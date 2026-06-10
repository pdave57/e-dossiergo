// Package service contains all application business logic.
// Use cases orchestrate domain entities via repository interfaces.
package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/crypto"
	"github.com/edossier/api/pkg/token"
	"github.com/edossier/api/pkg/validator"
)

// AuthService handles authentication and token lifecycle.
type AuthService struct {
	users         domain.UserRepository
	userRoles     domain.UserRoleRepository
	roles         domain.RoleRepository
	refreshTokens domain.RefreshTokenRepository
	tokenMaker    *token.Maker
}

func NewAuthService(
	users domain.UserRepository,
	userRoles domain.UserRoleRepository,
	roles domain.RoleRepository,
	refreshTokens domain.RefreshTokenRepository,
	tokenMaker *token.Maker,
) *AuthService {
	return &AuthService{
		users:         users,
		userRoles:     userRoles,
		roles:         roles,
		refreshTokens: refreshTokens,
		tokenMaker:    tokenMaker,
	}
}

// Register creates a new user account.
func (uc *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	v := validator.New().
		Required(req.StateID, "state_id").
		Required(req.Email, "email").
		ValidEmail(req.Email, "email").
		Required(req.FirstName, "first_name").
		Required(req.LastName, "last_name").
		StrongPassword(req.Password, "password")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	user := &domain.User{
		StateID:      req.StateID,
		SchoolID:     req.SchoolID,
		Email:        req.Email,
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       domain.UserStatusActive,
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return uc.buildAuthResponse(ctx, user)
}

// Login authenticates a user and issues tokens.
func (uc *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	v := validator.New().
		Required(req.Email, "email").
		Required(req.Password, "password")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	user, err := uc.users.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't leak whether email exists
		return nil, apperror.Unauthorized("invalid credentials")
	}
	if !user.IsActive() {
		return nil, apperror.Unauthorized("account is not active")
	}
	if err = crypto.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}

	_ = uc.users.UpdateLastLogin(ctx, user.ID) // best-effort

	return uc.buildAuthResponse(ctx, user)
}

// Refresh issues a new access token from a valid refresh token.
func (uc *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	userID, err := uc.tokenMaker.VerifyRefresh(req.RefreshToken)
	if err != nil {
		return nil, apperror.Unauthorized("invalid or expired refresh token")
	}

	stored, err := uc.refreshTokens.GetByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Unauthorized("refresh token not found")
	}

	// Verify hash matches
	incoming := fmt.Sprintf("%x", sha256.Sum256([]byte(req.RefreshToken)))
	if stored.TokenHash != incoming {
		return nil, apperror.Unauthorized("refresh token mismatch")
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.buildAuthResponse(ctx, user)
}

// Logout revokes the user's refresh token.
func (uc *AuthService) Logout(ctx context.Context, userID string) error {
	return uc.refreshTokens.Revoke(ctx, userID)
}

// ChangePassword validates the current password and updates it.
func (uc *AuthService) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	v := validator.New().
		Required(req.CurrentPassword, "current_password").
		StrongPassword(req.NewPassword, "new_password")
	if !v.Valid() {
		return apperror.Validation(v.Errors())
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err = crypto.CheckPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return apperror.Unauthorized("current password is incorrect")
	}

	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return apperror.Internal(err)
	}
	user.PasswordHash = hash
	return uc.users.Update(ctx, user)
}

// buildAuthResponse mints both tokens, stores the refresh, and builds the DTO.
func (uc *AuthService) buildAuthResponse(ctx context.Context, user *domain.User) (*dto.AuthResponse, error) {
	roles, _ := uc.userRoles.GetRolesForUser(ctx, user.ID)
	roleCodes := make([]string, len(roles))
	for i, r := range roles {
		roleCodes[i] = r.Code
	}

	accessTok, exp, err := uc.tokenMaker.CreateAccessToken(user.ID, user.StateID, user.SchoolID, roleCodes)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	refreshTok, refreshExp, err := uc.tokenMaker.CreateRefreshToken(user.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Store hashed refresh token
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(refreshTok)))
	_ = uc.refreshTokens.Save(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: refreshExp,
	})

	return &dto.AuthResponse{
		AccessToken:  accessTok,
		RefreshToken: refreshTok,
		ExpiresAt:    exp,
		User: dto.UserResponse{
			ID:          user.ID,
			StateID:     user.StateID,
			SchoolID:    user.SchoolID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Status:      string(user.Status),
			Roles:       roleCodes,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// USER MANAGEMENT USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type UserService struct {
	users     domain.UserRepository
	userRoles domain.UserRoleRepository
	roles     domain.RoleRepository
}

func NewUserService(
	users domain.UserRepository,
	userRoles domain.UserRoleRepository,
	roles domain.RoleRepository,
) *UserService {
	return &UserService{users: users, userRoles: userRoles, roles: roles}
}

func (uc *UserService) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	user, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	roles, _ := uc.userRoles.GetRolesForUser(ctx, user.ID)
	return toUserResponse(user, roles), nil
}

func (uc *UserService) List(ctx context.Context, f domain.UserFilter, p interface{}) ([]*dto.UserResponse, int, error) {
	return nil, 0, nil // delegated to handler with pagination
}

func (uc *UserService) Update(ctx context.Context, id string, req dto.UpdateUserRequest, updatedBy string) error {
	user, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	if req.SchoolID != "" {
		user.SchoolID = req.SchoolID
	}
	if req.Status != "" {
		user.Status = domain.UserStatus(req.Status)
	}
	user.UpdatedBy = updatedBy
	return uc.users.Update(ctx, user)
}

func (uc *UserService) Delete(ctx context.Context, id string) error {
	return uc.users.Delete(ctx, id)
}

func (uc *UserService) AssignRole(ctx context.Context, userID string, req dto.AssignRoleRequest, assignedBy string) error {
	return uc.userRoles.Assign(ctx, &domain.UserRole{
		UserID:     userID,
		RoleID:     req.RoleID,
		SchoolID:   req.SchoolID,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
	})
}

func (uc *UserService) RevokeRole(ctx context.Context, userID, roleID string) error {
	return uc.userRoles.Revoke(ctx, userID, roleID)
}

func (uc *UserService) GetRoles(ctx context.Context, userID string) ([]*dto.RoleResponse, error) {
	roles, err := uc.userRoles.GetRolesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.RoleResponse, len(roles))
	for i, r := range roles {
		out[i] = toRoleResponse(r, nil)
	}
	return out, nil
}

func toUserResponse(u *domain.User, roles []*domain.Role) *dto.UserResponse {
	roleCodes := make([]string, len(roles))
	for i, r := range roles {
		roleCodes[i] = r.Code
	}
	return &dto.UserResponse{
		ID:          u.ID,
		StateID:     u.StateID,
		SchoolID:    u.SchoolID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Status:      string(u.Status),
		Roles:       roleCodes,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ROLE / PERMISSION USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type RoleService struct {
	roles       domain.RoleRepository
	permissions domain.PermissionRepository
}

func NewRoleService(roles domain.RoleRepository, permissions domain.PermissionRepository) *RoleService {
	return &RoleService{roles: roles, permissions: permissions}
}

func (uc *RoleService) Create(ctx context.Context, stateID string, req dto.CreateRoleRequest, createdBy string) (*dto.RoleResponse, error) {
	v := validator.New().Required(req.Name, "name").Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	role := &domain.Role{
		StateID:     stateID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	if err := uc.roles.Create(ctx, role); err != nil {
		return nil, err
	}
	return toRoleResponse(role, nil), nil
}

func (uc *RoleService) GetByID(ctx context.Context, id string) (*dto.RoleResponse, error) {
	role, err := uc.roles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	perms, _ := uc.roles.GetPermissions(ctx, id)
	return toRoleResponse(role, perms), nil
}

func (uc *RoleService) List(ctx context.Context, stateID string) ([]*dto.RoleResponse, error) {
	roles, err := uc.roles.List(ctx, stateID)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.RoleResponse, len(roles))
	for i, r := range roles {
		out[i] = toRoleResponse(r, nil)
	}
	return out, nil
}

func (uc *RoleService) Update(ctx context.Context, id string, req dto.UpdateRoleRequest, updatedBy string) error {
	role, err := uc.roles.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return apperror.Forbidden("system roles cannot be modified")
	}
	role.Name = req.Name
	role.Description = req.Description
	role.UpdatedBy = updatedBy
	return uc.roles.Update(ctx, role)
}

func (uc *RoleService) Delete(ctx context.Context, id string) error {
	role, err := uc.roles.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return apperror.Forbidden("system roles cannot be deleted")
	}
	return uc.roles.Delete(ctx, id)
}

func (uc *RoleService) AddPermission(ctx context.Context, roleID string, req dto.AddPermissionRequest, grantedBy string) error {
	return uc.roles.AddPermission(ctx, &domain.RolePermission{
		RoleID:       roleID,
		PermissionID: req.PermissionID,
		GrantedBy:    grantedBy,
		GrantedAt:    time.Now(),
	})
}

func (uc *RoleService) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	return uc.roles.RemovePermission(ctx, roleID, permissionID)
}

func (uc *RoleService) ListPermissions(ctx context.Context) ([]*dto.PermissionResponse, error) {
	perms, err := uc.permissions.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.PermissionResponse, len(perms))
	for i, p := range perms {
		out[i] = &dto.PermissionResponse{ID: p.ID, Resource: p.Resource, Action: p.Action, Description: p.Description}
	}
	return out, nil
}

func toRoleResponse(r *domain.Role, perms []*domain.Permission) *dto.RoleResponse {
	resp := &dto.RoleResponse{
		ID:          r.ID,
		StateID:     r.StateID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
	}
	if len(perms) > 0 {
		resp.Permissions = make([]dto.PermissionResponse, len(perms))
		for i, p := range perms {
			resp.Permissions[i] = dto.PermissionResponse{ID: p.ID, Resource: p.Resource, Action: p.Action, Description: p.Description}
		}
	}
	return resp
}
