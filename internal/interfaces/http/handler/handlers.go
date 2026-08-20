// Package handler contains all HTTP request handlers for e-Dossier.
package handler

import (
	"net/http"
	"strconv"
	"time"

	chi "github.com/go-chi/chi/v5"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/application/service"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/internal/interfaces/http/middleware"
	"github.com/edossier/api/internal/interfaces/presenter"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/logger"
	"github.com/edossier/api/pkg/pagination"
)

// ─────────────────────────────────────────────────────────────────────────────
// AUTH HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type AuthHandler struct{ uc *service.AuthService }

func NewAuthHandler(uc *service.AuthService) *AuthHandler { return &AuthHandler{uc: uc} }

// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User registration info"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} apperror.AppError
// @Failure 409 {object} apperror.AppError
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	resp, err := h.uc.Register(r.Context(), req)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, resp)
}

// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	resp, err := h.uc.Login(r.Context(), req)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	resp, err := h.uc.Refresh(r.Context(), req)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary Logout user
// @Description Logout user by invalidating refresh token
// @Tags auth
// @Produce json
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	if err := h.uc.Logout(r.Context(), claims.UserID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Change user password
// @Description Change the authenticated user's password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Password change info"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ChangePasswordRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	if err := h.uc.ChangePassword(r.Context(), claims.UserID, req); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Get current user
// @Description Get the currently authenticated user profile
// @Tags auth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} apperror.AppError
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	presenter.JSON(w, http.StatusOK, claims)
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT AUTH HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type StudentAuthHandler struct{ uc *service.StudentAuthService }

func NewStudentAuthHandler(uc *service.StudentAuthService) *StudentAuthHandler {
	return &StudentAuthHandler{uc: uc}
}

// @Summary Student login
// @Description Authenticate student using school code and enrollment number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.StudentLoginRequest true "Student login credentials"
// @Success 200 {object} dto.StudentLoginResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Router /auth/student-login [post]
func (h *StudentAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.StudentLoginRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	resp, err := h.uc.Login(r.Context(), req)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// USER HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type UserHandler struct {
	uc       *service.UserService
	userRepo domain.UserRepository
}

func NewUserHandler(uc *service.UserService, userRepo domain.UserRepository) *UserHandler {
	return &UserHandler{uc: uc, userRepo: userRepo}
}

// @Summary List users
// @Description List users with optional filters
// @Tags users
// @Produce json
// @Param school_id query string false "Filter by school ID"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name or email"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} pagination.Meta
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	p := pagination.Parse(r)
	f := domain.UserFilter{
		StateID:  claims.StateID,
		SchoolID: r.URL.Query().Get("school_id"),
		Status:   r.URL.Query().Get("status"),
		Search:   r.URL.Query().Get("search"),
	}
	users, total, err := h.userRepo.List(r.Context(), f, p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, users, pagination.BuildMeta(p, total))
}

// @Summary Get user by ID
// @Description Get a single user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, user)
}

// @Summary Update user
// @Description Update user details
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "Updated user info"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateUserRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	if err := h.uc.Update(r.Context(), id, req, claims.UserID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Delete user
// @Description Delete a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.Delete(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Assign role to user
// @Description Assign a role to a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.AssignRoleRequest true "Role assignment info"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /users/{id}/roles [post]
func (h *UserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.AssignRoleRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	if err := h.uc.AssignRole(r.Context(), id, req, claims.UserID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Revoke role from user
// @Description Revoke a role from a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Param roleId path string true "Role ID"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /users/{id}/roles/{roleId} [delete]
func (h *UserHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	roleID := chi.URLParam(r, "roleId")
	if err := h.uc.RevokeRole(r.Context(), userID, roleID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Get user roles
// @Description Get roles assigned to a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} []dto.RoleResponse
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /users/{id}/roles [get]
func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	roles, err := h.uc.GetRoles(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, roles)
}

// ─────────────────────────────────────────────────────────────────────────────
// ROLE HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type RoleHandler struct{ uc *service.RoleService }

func NewRoleHandler(uc *service.RoleService) *RoleHandler { return &RoleHandler{uc: uc} }

// @Summary List roles
// @Description List all roles
// @Tags roles
// @Produce json
// @Success 200 {object} []dto.RoleResponse
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /roles [get]
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	roles, err := h.uc.List(r.Context(), claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, roles)
}

// @Summary Create role
// @Description Create a new role
// @Tags roles
// @Accept json
// @Produce json
// @Param request body dto.CreateRoleRequest true "Role info"
// @Success 201 {object} dto.RoleResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /roles [post]
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRoleRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	role, err := h.uc.Create(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, role)
}

// @Summary Get role by ID
// @Description Get a single role by ID
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} dto.RoleResponse
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /roles/{id} [get]
func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, role)
}

// @Summary Update role
// @Description Update role details
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body dto.UpdateRoleRequest true "Updated role info"
// @Success 200 {object} dto.RoleResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /roles/{id} [put]
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateRoleRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	if err := h.uc.Update(r.Context(), id, req, claims.UserID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Delete role
// @Description Delete a role
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /roles/{id} [delete]
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.Delete(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Add permission to role
// @Description Add a permission to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body dto.AddPermissionRequest true "Permission info"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /roles/{id}/permissions [post]
func (h *RoleHandler) AddPermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.AddPermissionRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	if err := h.uc.AddPermission(r.Context(), id, req, claims.UserID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Remove permission from role
// @Description Remove a permission from a role
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Param permId path string true "Permission ID"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /roles/{id}/permissions/{permId} [delete]
func (h *RoleHandler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	permID := chi.URLParam(r, "permId")
	if err := h.uc.RemovePermission(r.Context(), roleID, permID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List permissions
// @Description List all permissions
// @Tags roles
// @Produce json
// @Success 200 {object} []dto.PermissionResponse
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /permissions [get]
func (h *RoleHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.uc.ListPermissions(r.Context())
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, perms)
}

// ─────────────────────────────────────────────────────────────────────────────
// STATE HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type StateHandler struct{ uc *service.StateService }

func NewStateHandler(uc *service.StateService) *StateHandler { return &StateHandler{uc: uc} }

// @Summary List states
// @Description List all states
// @Tags states
// @Produce json
// @Success 200 {object} []domain.State
// @Router /states [get]
func (h *StateHandler) ListStates(w http.ResponseWriter, r *http.Request) {
	states, err := h.uc.ListStates(r.Context())
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, states)
}

// @Summary Create state
// @Description Create a new state
// @Tags states
// @Accept json
// @Produce json
// @Param request body dto.CreateStateRequest true "State info"
// @Success 201 {object} domain.State
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /states [post]
func (h *StateHandler) CreateState(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateStateRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.CreateState(r.Context(), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, s)
}

// @Summary Get state by ID
// @Description Get a single state by ID
// @Tags states
// @Produce json
// @Param id path string true "State ID"
// @Success 200 {object} domain.State
// @Failure 404 {object} apperror.AppError
// @Router /states/{id} [get]
func (h *StateHandler) GetState(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetState(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Update state
// @Description Update state details
// @Tags states
// @Accept json
// @Produce json
// @Param id path string true "State ID"
// @Param request body dto.UpdateStateRequest true "Updated state info"
// @Success 200 {object} domain.State
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /states/{id} [put]
func (h *StateHandler) UpdateState(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateStateRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.UpdateState(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// ─────────────────────────────────────────────────────────────────────────────
// ZONE HANDLER
// ─────────────────────────────────────────────────────────────────────────────
type ZoneHandler struct{ uc *service.ZoneService }

func NewZoneHandler(uc *service.ZoneService) *ZoneHandler { return &ZoneHandler{uc: uc} }

// @Summary List zones
// @Description List zones for a state
// @Tags zones
// @Produce json
// @Param stateId path string true "State ID"
// @Success 200 {object} []domain.Zone
// @Router /states/{stateId}/zones [get]
func (h *ZoneHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	zones, err := h.uc.ListZones(r.Context(), chi.URLParam(r, "stateId"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, zones)
}

// @Summary Create zone
// @Description Create a new zone
// @Tags zones
// @Accept json
// @Produce json
// @Param stateId path string true "State ID"
// @Param request body dto.CreateZoneRequest true "Zone info"
// @Success 201 {object} domain.Zone
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /states/{stateId}/zones [post]
func (h *ZoneHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateZoneRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	z, err := h.uc.CreateZone(r.Context(), chi.URLParam(r, "stateId"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, z)
}

// @Summary Update zone
// @Description Update zone details
// @Tags zones
// @Accept json
// @Produce json
// @Param id path string true "Zone ID"
// @Param request body dto.UpdateZoneRequest true "Updated zone info"
// @Success 200 {object} domain.Zone
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /zones/{id} [put]
func (h *ZoneHandler) UpdateZone(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateZoneRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	z, err := h.uc.UpdateZone(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, z)
}

// @Summary Delete zone
// @Description Delete a zone
// @Tags zones
// @Produce json
// @Param id path string true "Zone ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /zones/{id} [delete]
func (h *ZoneHandler) DeleteZone(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteZone(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// ─────────────────────────────────────────────────────────────────────────────
// LGA HANDLER
// ─────────────────────────────────────────────────────────────────────────────
type LGAHandler struct{ uc *service.LGAService }

func NewLGAHandler(uc *service.LGAService) *LGAHandler { return &LGAHandler{uc: uc} }

// @Summary List LGAs
// @Description List LGAs for a state or zone
// @Tags lgas
// @Produce json
// @Param stateId path string true "State ID"
// @Param zone_id query string false "Filter by zone ID"
// @Success 200 {object} []domain.LGA
// @Router /states/{stateId}/lgas [get]
func (h *LGAHandler) ListLGAs(w http.ResponseWriter, r *http.Request) {
	stateID := chi.URLParam(r, "stateId")
	zoneID := r.URL.Query().Get("zone_id")
	var (
		lgas []*domain.LGA
		err  error
	)
	if zoneID != "" {
		lgas, err = h.uc.ListLGAsByZone(r.Context(), zoneID)
	} else {
		lgas, err = h.uc.ListLGAs(r.Context(), stateID)
	}
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, lgas)
}

// @Summary Create LGA
// @Description Create a new LGA
// @Tags lgas
// @Accept json
// @Produce json
// @Param stateId path string true "State ID"
// @Param request body dto.CreateLGARequest true "LGA info"
// @Success 201 {object} domain.LGA
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /states/{stateId}/lgas [post]
func (h *LGAHandler) CreateLGA(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLGARequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	l, err := h.uc.CreateLGA(r.Context(), chi.URLParam(r, "stateId"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, l)
}

// @Summary Update LGA
// @Description Update LGA details
// @Tags lgas
// @Accept json
// @Produce json
// @Param id path string true "LGA ID"
// @Param request body dto.UpdateLGARequest true "Updated LGA info"
// @Success 200 {object} domain.LGA
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /lgas/{id} [put]
func (h *LGAHandler) UpdateLGA(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateLGARequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	l, err := h.uc.UpdateLGA(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, l)
}

// @Summary Delete LGA
// @Description Delete an LGA
// @Tags lgas
// @Produce json
// @Param id path string true "LGA ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /lgas/{id} [delete]
func (h *LGAHandler) DeleteLGA(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteLGA(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type SchoolHandler struct{ uc *service.SchoolService }

func NewSchoolHandler(uc *service.SchoolService) *SchoolHandler { return &SchoolHandler{uc: uc} }

// @Summary List schools
// @Description List schools with optional filters
// @Tags schools
// @Produce json
// @Param state_id query string false "Filter by state ID"
// @Param zone_id query string false "Filter by zone ID"
// @Param lga_id query string false "Filter by LGA ID"
// @Param category query string false "Filter by category"
// @Param ownership query string false "Filter by ownership"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name or code"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} pagination.Meta
// @Router /schools [get]
func (h *SchoolHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	f := domain.SchoolFilter{
		StateID:   q.Get("state_id"),
		ZoneID:    q.Get("zone_id"),
		LGAID:     q.Get("lga_id"),
		Category:  q.Get("category"),
		Ownership: q.Get("ownership"),
		Status:    q.Get("status"),
		Search:    q.Get("search"),
	}
	schools, total, err := h.uc.List(r.Context(), f, p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, schools, pagination.BuildMeta(p, total))
}

// @Summary Create school
// @Description Create a new school
// @Tags schools
// @Accept json
// @Produce json
// @Param request body dto.CreateSchoolRequest true "School info"
// @Success 201 {object} domain.School
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /schools [post]
func (h *SchoolHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSchoolRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	// Prefer the state selected in the request body; fall back to the
	// caller's state from the token (e.g. state admins).
	stateID := req.StateID
	if stateID == "" {
		stateID = claims.StateID
	}
	s, err := h.uc.Create(r.Context(), stateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, s)
}

// @Summary Get school by ID
// @Description Get a single school by ID
// @Tags schools
// @Produce json
// @Param id path string true "School ID"
// @Success 200 {object} domain.School
// @Failure 404 {object} apperror.AppError
// @Router /schools/{id} [get]
func (h *SchoolHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Update school
// @Description Update school details
// @Tags schools
// @Accept json
// @Produce json
// @Param id path string true "School ID"
// @Param request body dto.UpdateSchoolRequest true "Updated school info"
// @Success 200 {object} domain.School
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /schools/{id} [put]
func (h *SchoolHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSchoolRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.Update(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Delete school
// @Description Delete a school
// @Tags schools
// @Produce json
// @Param id path string true "School ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /schools/{id} [delete]
func (h *SchoolHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List school facilities
// @Description List facilities for a school
// @Tags schools
// @Produce json
// @Param id path string true "School ID"
// @Success 200 {object} []domain.SchoolFacility
// @Router /schools/{id}/facilities [get]
func (h *SchoolHandler) ListFacilities(w http.ResponseWriter, r *http.Request) {
	facs, err := h.uc.ListFacilities(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, facs)
}

// @Summary Add facility to school
// @Description Add a facility to a school
// @Tags schools
// @Accept json
// @Produce json
// @Param id path string true "School ID"
// @Param request body dto.CreateFacilityRequest true "Facility info"
// @Success 201 {object} domain.SchoolFacility
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /schools/{id}/facilities [post]
func (h *SchoolHandler) AddFacility(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateFacilityRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	f, err := h.uc.AddFacility(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, f)
}

// @Summary Update school facility
// @Description Update a school facility
// @Tags schools
// @Accept json
// @Produce json
// @Param id path string true "School ID"
// @Param facilityId path string true "Facility ID"
// @Param request body dto.UpdateFacilityRequest true "Updated facility info"
// @Success 200 {object} domain.SchoolFacility
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /schools/{id}/facilities/{facilityId} [put]
func (h *SchoolHandler) UpdateFacility(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateFacilityRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	f, err := h.uc.UpdateFacility(r.Context(), chi.URLParam(r, "facilityId"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, f)
}

// @Summary Delete school facility
// @Description Delete a school facility
// @Tags schools
// @Produce json
// @Param id path string true "School ID"
// @Param facilityId path string true "Facility ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /schools/{id}/facilities/{facilityId} [delete]
func (h *SchoolHandler) DeleteFacility(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteFacility(r.Context(), chi.URLParam(r, "facilityId")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Count total schools
// @Description Count total schools by state
// @Tags schools
// @Produce json
// @Param state_id query string true "State ID"
// @Success 200 {object} int
// @Router /reports/schools/total [get]
func (h *SchoolHandler) CountTotalSchools(w http.ResponseWriter, r *http.Request) {
	stateID := r.URL.Query().Get("state_id")
	count, err := h.uc.CountTotalSchools(r.Context(), stateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, count)
}

// @Summary Upload school logo
// @Description Upload a logo for a school
// @Tags schools
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "School ID"
// @Param file formData file true "Logo image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /schools/{id}/logo [put]
func (h *SchoolHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "id")
	file, header, err := r.FormFile("file")
	if err != nil {
		presenter.Error(w, apperror.BadRequest("file is required"))
		return
	}
	defer file.Close()

	url, err := h.uc.UploadLogo(r.Context(), schoolID, file, header.Filename)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, map[string]string{"logo_url": url})
}

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type AcademicHandler struct{ uc *service.AcademicService }

func NewAcademicHandler(uc *service.AcademicService) *AcademicHandler {
	return &AcademicHandler{uc: uc}
}

// @Summary List academic sessions
// @Description List academic sessions for a school
// @Tags academic
// @Produce json
// @Param school_id query string false "School ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} pagination.Meta
// @Failure 400 {object} apperror.AppError
// @Router /sessions [get]
func (h *AcademicHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := claims.SchoolID
	if schoolID == "" {
		schoolID = r.URL.Query().Get("school_id")
	}
	if schoolID == "" {
		presenter.Error(w, apperror.BadRequest("school_id is required"))
		return
	}
	p := pagination.Parse(r)
	sessions, total, err := h.uc.ListSessions(r.Context(), schoolID, p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, sessions, pagination.BuildMeta(p, total))
}

// @Summary Create academic session
// @Description Create a new academic session
// @Tags academic
// @Accept json
// @Produce json
// @Param school_id query string false "School ID"
// @Param request body dto.CreateSessionRequest true "Session info"
// @Success 201 {object} domain.AcademicSession
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /sessions [post]
func (h *AcademicHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSessionRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := claims.SchoolID
	if schoolID == "" {
		schoolID = r.URL.Query().Get("school_id")
	}
	if schoolID == "" {
		presenter.Error(w, apperror.BadRequest("school_id is required"))
		return
	}
	s, err := h.uc.CreateSession(r.Context(), schoolID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, s)
}

// @Summary Get session by ID
// @Description Get a single academic session by ID
// @Tags academic
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} domain.AcademicSession
// @Failure 404 {object} apperror.AppError
// @Router /sessions/{id} [get]
func (h *AcademicHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetSession(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Get active session
// @Description Get the active academic session
// @Tags academic
// @Produce json
// @Param school_id query string false "School ID"
// @Success 200 {object} domain.AcademicSession
// @Failure 400 {object} apperror.AppError
// @Router /sessions/active [get]
func (h *AcademicHandler) GetActiveSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := claims.SchoolID
	if schoolID == "" {
		schoolID = r.URL.Query().Get("school_id")
	}
	if schoolID == "" {
		presenter.Error(w, apperror.BadRequest("school_id is required"))
		return
	}
	s, err := h.uc.GetActiveSession(r.Context(), schoolID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Update session
// @Description Update academic session details
// @Tags academic
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param request body dto.UpdateSessionRequest true "Updated session info"
// @Success 200 {object} domain.AcademicSession
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /sessions/{id} [put]
func (h *AcademicHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSessionRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.UpdateSession(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Activate session
// @Description Activate an academic session
// @Tags academic
// @Produce json
// @Param id path string true "Session ID"
// @Param school_id query string false "School ID"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /sessions/{id}/activate [post]
func (h *AcademicHandler) ActivateSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := claims.SchoolID
	if schoolID == "" {
		schoolID = r.URL.Query().Get("school_id")
	}
	if schoolID == "" {
		presenter.Error(w, apperror.BadRequest("school_id is required"))
		return
	}
	if err := h.uc.ActivateSession(r.Context(), chi.URLParam(r, "id"), schoolID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Delete session
// @Description Delete an academic session
// @Tags academic
// @Produce json
// @Param id path string true "Session ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /sessions/{id} [delete]
func (h *AcademicHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteSession(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List terms
// @Description List terms for a session
// @Tags academic
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} []domain.Term
// @Router /sessions/{sessionId}/terms [get]
func (h *AcademicHandler) ListTerms(w http.ResponseWriter, r *http.Request) {
	terms, err := h.uc.ListTerms(r.Context(), chi.URLParam(r, "sessionId"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, terms)
}

// @Summary Create term
// @Description Create a new term
// @Tags academic
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.CreateTermRequest true "Term info"
// @Success 201 {object} domain.Term
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /sessions/{sessionId}/terms [post]
func (h *AcademicHandler) CreateTerm(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTermRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	t, err := h.uc.CreateTerm(r.Context(), chi.URLParam(r, "sessionId"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, t)
}

// @Summary Update term
// @Description Update term details
// @Tags academic
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param id path string true "Term ID"
// @Param request body dto.UpdateTermRequest true "Updated term info"
// @Success 200 {object} domain.Term
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /sessions/{sessionId}/terms/{id} [put]
func (h *AcademicHandler) UpdateTerm(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateTermRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	t, err := h.uc.UpdateTerm(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, t)
}

// @Summary Activate term
// @Description Activate a term
// @Tags academic
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param id path string true "Term ID"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /sessions/{sessionId}/terms/{id}/activate [post]
func (h *AcademicHandler) ActivateTerm(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	termID := chi.URLParam(r, "id")
	if err := h.uc.ActivateTerm(r.Context(), termID, sessionID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Delete term
// @Description Delete a term
// @Tags academic
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param id path string true "Term ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /sessions/{sessionId}/terms/{id} [delete]
func (h *AcademicHandler) DeleteTerm(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteTerm(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// GetTerm returns a single term by id (top-level /terms/{id}).
// @Summary Get term by ID
// @Description Get a single term by ID
// @Tags academic
// @Produce json
// @Param id path string true "Term ID"
// @Success 200 {object} domain.Term
// @Failure 404 {object} apperror.AppError
// @Router /terms/{id} [get]
func (h *AcademicHandler) GetTerm(w http.ResponseWriter, r *http.Request) {
	t, err := h.uc.GetTerm(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, t)
}

// ListAllTerms returns all terms across every session (top-level /terms).
// @Summary List all terms
// @Description List all terms across every session
// @Tags academic
// @Produce json
// @Success 200 {object} []domain.Term
// @Router /terms [get]
func (h *AcademicHandler) ListAllTerms(w http.ResponseWriter, r *http.Request) {
	terms, err := h.uc.ListAllTerms(r.Context())
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, terms)
}

// GetActiveTerm returns the active term for an optional session.
// Query param: session_id (optional). If omitted, returns the currently active term globally.
// @Summary Get active term
// @Description Get the active term
// @Tags academic
// @Produce json
// @Param session_id query string false "Session ID"
// @Success 200 {object} domain.Term
// @Router /terms/active [get]
func (h *AcademicHandler) GetActiveTerm(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	t, err := h.uc.GetActiveTerm(r.Context(), sessionID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, t)
}

// CreateTermTopLevel creates a term via the top-level /terms endpoint, where the
// owning session is provided in the request body.
// @Summary Create term (top-level)
// @Description Create a term via the top-level endpoint
// @Tags academic
// @Accept json
// @Produce json
// @Param request body dto.CreateTermRequest true "Term info"
// @Success 201 {object} domain.Term
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /terms [post]
func (h *AcademicHandler) CreateTermTopLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTermRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	t, err := h.uc.CreateTermTopLevel(r.Context(), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, t)
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type LevelHandler struct{ uc *service.LevelService }

func NewLevelHandler(uc *service.LevelService) *LevelHandler { return &LevelHandler{uc: uc} }

// @Summary List levels
// @Description List all levels
// @Tags levels
// @Produce json
// @Success 200 {object} []domain.Level
// @Router /levels [get]
func (h *LevelHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	levels, err := h.uc.ListLevels(r.Context(), claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, levels)
}

// @Summary Create level
// @Description Create a new level
// @Tags levels
// @Accept json
// @Produce json
// @Param request body dto.CreateLevelRequest true "Level info"
// @Success 201 {object} domain.Level
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /levels [post]
func (h *LevelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLevelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	l, err := h.uc.CreateLevel(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, l)
}

// @Summary Get level by ID
// @Description Get a single level by ID
// @Tags levels
// @Produce json
// @Param id path string true "Level ID"
// @Success 200 {object} domain.Level
// @Failure 404 {object} apperror.AppError
// @Router /levels/{id} [get]
func (h *LevelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	l, err := h.uc.GetLevel(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, l)
}

// @Summary Update level
// @Description Update level details
// @Tags levels
// @Accept json
// @Produce json
// @Param id path string true "Level ID"
// @Param request body dto.UpdateLevelRequest true "Updated level info"
// @Success 200 {object} domain.Level
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /levels/{id} [put]
func (h *LevelHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateLevelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	l, err := h.uc.UpdateLevel(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, l)
}

// @Summary Delete level
// @Description Delete a level
// @Tags levels
// @Produce json
// @Param id path string true "Level ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /levels/{id} [delete]
func (h *LevelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteLevel(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List sub-levels
// @Description List sub-levels (class arms) for a school and level
// @Tags levels
// @Produce json
// @Param schoolId path string true "School ID"
// @Param levelId path string true "Level ID"
// @Success 200 {object} []domain.SubLevel
// @Router /levels/{levelId}/sub-levels [get]
func (h *LevelHandler) ListSubLevels(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	levelID := chi.URLParam(r, "levelId")
	sls, err := h.uc.ListSubLevels(r.Context(), schoolID, levelID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, sls)
}

// @Summary Create sub-level
// @Description Create a sub-level for a school
// @Tags levels
// @Accept json
// @Produce json
// @Param schoolId path string true "School ID"
// @Param request body dto.CreateSubLevelRequest true "Sub-level info"
// @Success 201 {object} domain.SubLevel
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /schools/{schoolId}/sub-levels [post]
func (h *LevelHandler) CreateSubLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubLevelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := chi.URLParam(r, "schoolId")
	sl, err := h.uc.CreateSubLevel(r.Context(), schoolID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, sl)
}

// CreateSubLevelGlobal handles POST /api/v1/sub-levels, reading the owning
// school_id and level_id from the query string so callers can create a
// sub-level from a page URL that already carries those identifiers.
// @Summary Create sub-level (global)
// @Description Create a sub-level via top-level endpoint
// @Tags levels
// @Accept json
// @Produce json
// @Param school_id query string true "School ID"
// @Param level_id query string true "Level ID"
// @Param request body dto.CreateSubLevelRequest true "Sub-level info"
// @Success 201 {object} domain.SubLevel
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /sub-levels [post]
func (h *LevelHandler) CreateSubLevelGlobal(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubLevelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	q := r.URL.Query()
	if s := q.Get("school_id"); s != "" {
		req.SchoolID = s
	}
	if l := q.Get("level_id"); l != "" {
		req.LevelID = l
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	sl, err := h.uc.CreateSubLevel(r.Context(), req.SchoolID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, sl)
}

// @Summary Update sub-level
// @Description Update sub-level details
// @Tags levels
// @Accept json
// @Produce json
// @Param id path string true "Sub-level ID"
// @Param request body dto.UpdateSubLevelRequest true "Updated sub-level info"
// @Success 200 {object} domain.SubLevel
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /sub-levels/{id} [put]
func (h *LevelHandler) UpdateSubLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSubLevelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	sl, err := h.uc.UpdateSubLevel(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, sl)
}

// @Summary Delete sub-level
// @Description Delete a sub-level
// @Tags levels
// @Produce json
// @Param id path string true "Sub-level ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /sub-levels/{id} [delete]
func (h *LevelHandler) DeleteSubLevel(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteSubLevel(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Upsert school level
// @Description Assign or update a level for a school
// @Tags levels
// @Accept json
// @Produce json
// @Param schoolId path string true "School ID"
// @Param request body dto.UpsertSchoolLevelRequest true "School level info"
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /schools/{schoolId}/levels [post]
func (h *LevelHandler) UpsertSchoolLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.UpsertSchoolLevelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := chi.URLParam(r, "schoolId")
	if err := h.uc.UpsertSchoolLevel(r.Context(), schoolID, req, claims.UserID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List school levels
// @Description List levels assigned to a school
// @Tags levels
// @Produce json
// @Param schoolId path string true "School ID"
// @Param session_id query string false "Filter by session ID"
// @Success 200 {object} []domain.SchoolLevel
// @Router /schools/{schoolId}/levels [get]
func (h *LevelHandler) ListSchoolLevels(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	sessionID := r.URL.Query().Get("session_id")
	sls, err := h.uc.ListSchoolLevels(r.Context(), schoolID, sessionID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, sls)
}

// ─────────────────────────────────────────────────────────────────────────────
// SUBJECT HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type SubjectHandler struct{ uc *service.SubjectService }

func NewSubjectHandler(uc *service.SubjectService) *SubjectHandler { return &SubjectHandler{uc: uc} }

// @Summary List subjects
// @Description List all subjects
// @Tags subjects
// @Produce json
// @Success 200 {object} []domain.Subject
// @Router /subjects [get]
func (h *SubjectHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	subjects, err := h.uc.ListByState(r.Context(), claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, subjects)
}

// @Summary Create subject
// @Description Create a new subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param request body dto.CreateSubjectRequest true "Subject info"
// @Success 201 {object} domain.Subject
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /subjects [post]
func (h *SubjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubjectRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.Create(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, s)
}

// @Summary Get subject by ID
// @Description Get a single subject by ID
// @Tags subjects
// @Produce json
// @Param id path string true "Subject ID"
// @Success 200 {object} domain.Subject
// @Failure 404 {object} apperror.AppError
// @Router /subjects/{id} [get]
func (h *SubjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Update subject
// @Description Update subject details
// @Tags subjects
// @Accept json
// @Produce json
// @Param id path string true "Subject ID"
// @Param request body dto.UpdateSubjectRequest true "Updated subject info"
// @Success 200 {object} domain.Subject
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /subjects/{id} [put]
func (h *SubjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSubjectRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.Update(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Delete subject
// @Description Delete a subject
// @Tags subjects
// @Produce json
// @Param id path string true "Subject ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /subjects/{id} [delete]
func (h *SubjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List school subjects
// @Description List subjects assigned to a school
// @Tags subjects
// @Produce json
// @Param schoolId path string true "School ID"
// @Param session_id query string false "Filter by session ID"
// @Param level_id query string false "Filter by level ID"
// @Success 200 {object} []domain.SchoolSubject
// @Router /schools/{schoolId}/subjects [get]
func (h *SubjectHandler) ListSchoolSubjects(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	sessionID := r.URL.Query().Get("session_id")
	levelID := r.URL.Query().Get("level_id")
	ss, err := h.uc.ListSchoolSubjects(r.Context(), schoolID, sessionID, levelID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, ss)
}

// @Summary Assign subject to school
// @Description Assign a subject to a school
// @Tags subjects
// @Accept json
// @Produce json
// @Param schoolId path string true "School ID"
// @Param request body dto.CreateSchoolSubjectRequest true "School subject info"
// @Success 201 {object} domain.SchoolSubject
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /schools/{schoolId}/subjects [post]
func (h *SubjectHandler) AssignToSchool(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSchoolSubjectRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	ss, err := h.uc.AssignToSchool(r.Context(), chi.URLParam(r, "schoolId"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, ss)
}

// @Summary Update school subject
// @Description Update a school subject assignment
// @Tags subjects
// @Accept json
// @Produce json
// @Param id path string true "School Subject ID"
// @Param request body dto.UpdateSchoolSubjectRequest true "Updated school subject info"
// @Success 200 {object} domain.SchoolSubject
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /school-subjects/{id} [put]
func (h *SubjectHandler) UpdateSchoolSubject(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSchoolSubjectRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	ss, err := h.uc.UpdateSchoolSubject(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, ss)
}

// @Summary Remove school subject
// @Description Remove a subject from a school
// @Tags subjects
// @Produce json
// @Param id path string true "School Subject ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /school-subjects/{id} [delete]
func (h *SubjectHandler) RemoveSchoolSubject(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.RemoveSchoolSubject(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type PersonnelHandler struct{ uc *service.PersonnelService }

func NewPersonnelHandler(uc *service.PersonnelService) *PersonnelHandler {
	return &PersonnelHandler{uc: uc}
}

// @Summary List personnel
// @Description List personnel with optional filters
// @Tags personnel
// @Produce json
// @Param school_id query string false "Filter by school ID"
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} pagination.Meta
// @Router /personnel [get]
func (h *PersonnelHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	claims := middleware.ClaimsFromCtx(r.Context())
	q := r.URL.Query()
	f := domain.PersonnelFilter{
		StateID:  claims.StateID,
		SchoolID: q.Get("school_id"),
		Role:     q.Get("role"),
		Status:   q.Get("status"),
		Search:   q.Get("search"),
	}
	staff, total, err := h.uc.List(r.Context(), f, p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, staff, pagination.BuildMeta(p, total))
}

// @Summary Create personnel
// @Description Create a new personnel record
// @Tags personnel
// @Accept json
// @Produce json
// @Param request body dto.CreatePersonnelRequest true "Personnel info"
// @Success 201 {object} domain.Personnel
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /personnel [post]
func (h *PersonnelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePersonnelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	p, err := h.uc.Create(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, p)
}

// @Summary Get personnel by ID
// @Description Get a single personnel by ID
// @Tags personnel
// @Produce json
// @Param id path string true "Personnel ID"
// @Success 200 {object} domain.Personnel
// @Failure 404 {object} apperror.AppError
// @Router /personnel/{id} [get]
func (h *PersonnelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	p, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, p)
}

// @Summary Update personnel
// @Description Update personnel details
// @Tags personnel
// @Accept json
// @Produce json
// @Param id path string true "Personnel ID"
// @Param request body dto.UpdatePersonnelRequest true "Updated personnel info"
// @Success 200 {object} domain.Personnel
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /personnel/{id} [put]
func (h *PersonnelHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePersonnelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	p, err := h.uc.Update(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, p)
}

// @Summary Delete personnel
// @Description Delete a personnel record
// @Tags personnel
// @Produce json
// @Param id path string true "Personnel ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /personnel/{id} [delete]
func (h *PersonnelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Transfer personnel
// @Description Transfer personnel to another school
// @Tags personnel
// @Accept json
// @Produce json
// @Param id path string true "Personnel ID"
// @Param request body dto.TransferPersonnelRequest true "Transfer info"
// @Success 201 {object} domain.PersonnelTransfer
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /personnel/{id}/transfer [post]
func (h *PersonnelHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req dto.TransferPersonnelRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	t, err := h.uc.Transfer(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, t)
}

// @Summary List personnel transfers
// @Description List transfer history for a personnel
// @Tags personnel
// @Produce json
// @Param id path string true "Personnel ID"
// @Success 200 {object} []domain.PersonnelTransfer
// @Router /personnel/{id}/transfers [get]
func (h *PersonnelHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	transfers, err := h.uc.ListTransfers(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, transfers)
}

// ─────────────────────────────────────────────────────────────────────────────
// TOTAL PERSONNEL HANDLER
// ─────────────────────────────────────────────────────────────────────────────

// @Summary Count total personnel
// @Description Count total personnel by state
// @Tags personnel
// @Produce json
// @Param state_id query string true "State ID"
// @Success 200 {object} int
// @Router /reports/personnel/total [get]
func (h *PersonnelHandler) CountTotalPersonnel(w http.ResponseWriter, r *http.Request) {
	stateID := r.URL.Query().Get("state_id")
	count, err := h.uc.CountTotalPersonnel(r.Context(), stateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, count)
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type StudentHandler struct{ uc *service.StudentService }

func NewStudentHandler(uc *service.StudentService) *StudentHandler { return &StudentHandler{uc: uc} }

// @Summary List students
// @Description List students with optional filters
// @Tags students
// @Produce json
// @Param school_id query string false "Filter by school ID"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} pagination.Meta
// @Router /students [get]
func (h *StudentHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	claims := middleware.ClaimsFromCtx(r.Context())
	q := r.URL.Query()
	f := domain.StudentFilter{
		StateID:  claims.StateID,
		SchoolID: q.Get("school_id"),
		Status:   q.Get("status"),
		Search:   q.Get("search"),
	}
	students, total, err := h.uc.List(r.Context(), f, p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, students, pagination.BuildMeta(p, total))
}

// @Summary Create student
// @Description Register a new student
// @Tags students
// @Accept json
// @Produce json
// @Param request body dto.CreateStudentRequest true "Student info"
// @Success 201 {object} domain.Student
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /students [post]
func (h *StudentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateStudentRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.Register(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, s)
}

// @Summary Get next serial number
// @Description Get the next student serial number for a school
// @Tags students
// @Produce json
// @Param school_id query string false "School ID"
// @Param year query int false "Enrollment year"
// @Success 200 {object} map[string]int
// @Router /students/next-serial [get]
func (h *StudentHandler) GetNextSerial(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	schoolID := q.Get("school_id")
	year := 0
	if y := q.Get("year"); y != "" {
		year, _ = strconv.Atoi(y)
	}
	serial, err := h.uc.GetNextSerial(r.Context(), schoolID, year)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, map[string]int{"serial_no": serial})
}

// @Summary Get student by ID
// @Description Get a single student by ID
// @Tags students
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} domain.Student
// @Failure 404 {object} apperror.AppError
// @Router /students/{id} [get]
func (h *StudentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Update student
// @Description Update student details
// @Tags students
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param request body dto.UpdateStudentRequest true "Updated student info"
// @Success 200 {object} domain.Student
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /students/{id} [put]
func (h *StudentHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateStudentRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	s, err := h.uc.Update(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

// @Summary Delete student
// @Description Delete a student
// @Tags students
// @Produce json
// @Param id path string true "Student ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /students/{id} [delete]
func (h *StudentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Enroll student
// @Description Enroll a student in a class
// @Tags enrollments
// @Accept json
// @Produce json
// @Param request body dto.EnrollStudentRequest true "Enrollment info"
// @Success 201 {object} domain.Enrollment
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /enrollments [post]
func (h *StudentHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	var req dto.EnrollStudentRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	e, err := h.uc.Enroll(r.Context(), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, e)
}

// @Summary List enrollments
// @Description List student enrollments with optional filters
// @Tags enrollments
// @Produce json
// @Param school_id query string false "Filter by school ID"
// @Param session_id query string false "Filter by session ID"
// @Param level_id query string false "Filter by level ID"
// @Param sub_level_id query string false "Filter by sub-level ID"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} pagination.Meta
// @Router /enrollments [get]
func (h *StudentHandler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	f := domain.EnrollmentFilter{
		SchoolID:   q.Get("school_id"),
		SessionID:  q.Get("session_id"),
		LevelID:    q.Get("level_id"),
		SubLevelID: q.Get("sub_level_id"),
		Status:     q.Get("status"),
	}
	enrollments, total, err := h.uc.ListEnrollments(r.Context(), f, p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, enrollments, pagination.BuildMeta(p, total))
}

// @Summary Update enrollment
// @Description Update an enrollment record
// @Tags enrollments
// @Accept json
// @Produce json
// @Param enrollmentId path string true "Enrollment ID"
// @Param request body dto.UpdateEnrollmentRequest true "Updated enrollment info"
// @Success 200 {object} domain.Enrollment
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /enrollments/{enrollmentId} [put]
func (h *StudentHandler) UpdateEnrollment(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateEnrollmentRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	e, err := h.uc.UpdateEnrollment(r.Context(), chi.URLParam(r, "enrollmentId"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, e)
}

// @Summary Record progression
// @Description Record a student level progression
// @Tags students
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param school_id query string false "School ID"
// @Param request body dto.RecordProgressionRequest true "Progression info"
// @Success 201 {object} domain.LevelProgression
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /students/{id}/progressions [post]
func (h *StudentHandler) RecordProgression(w http.ResponseWriter, r *http.Request) {
	var req dto.RecordProgressionRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := r.URL.Query().Get("school_id")
	lp, err := h.uc.RecordProgression(r.Context(), schoolID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, lp)
}

// @Summary List progressions
// @Description List level progressions for a student
// @Tags students
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} []domain.LevelProgression
// @Router /students/{id}/progressions [get]
func (h *StudentHandler) ListProgressions(w http.ResponseWriter, r *http.Request) {
	lps, err := h.uc.ListProgressions(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, lps)
}

// // @Summary Upload student avatar
// @Description Upload a profile picture for student
// @Tags avatar
// @Accept multipart/form-data
// @Produce json
// @Param student_id formData string true "Student ID"
// @Param file formData file true "Image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /avatar/students/{id} [put]
// func (h *AvatarHandler) UploadStudentAvatar(w http.ResponseWriter, r *http.Request) {
// 	studentID := r.FormValue("student_id")
// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		presenter.Error(w, apperror.BadRequest("file is required"))
// 		return
// 	}
// 	defer file.Close()
// 	url, err := h.avatarService.UploadStudentAvatar(r.Context(), studentID, file, header.Filename)
// 	if err != nil {
// 		presenter.Error(w, err)
// 		return
// 	}
// 	presenter.JSON(w, http.StatusOK, map[string]string{"avatar_url": url})
// }

// ─────────────────────────────────────────────────────────────────────────────
// GENDER HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type GenderHandler struct{ uc service.GenderService }

func NewGenderHandler(uc service.GenderService) *GenderHandler { return &GenderHandler{uc: uc} }

// @Summary Count by gender
// @Description Count students by gender for a state
// @Tags reports
// @Produce json
// @Param state_id query string true "State ID"
// @Success 200 {object} map[string]int
// @Router /reports/gender/total [get]
func (h *GenderHandler) CountByGender(w http.ResponseWriter, r *http.Request) {
	stateID := r.URL.Query().Get("state_id")
	counts, err := h.uc.CountByGender(r.Context(), stateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, counts)
}

// ─────────────────────────────────────────────────────────────────────────────
// TOTAL STUDENT HANDLER
// ─────────────────────────────────────────────────────────────────────────────

// @Summary Count total students
// @Description Count total students by state
// @Tags reports
// @Produce json
// @Param state_id query string true "State ID"
// @Success 200 {object} int
// @Router /reports/students/total [get]
func (h *StudentHandler) CountTotalStudents(w http.ResponseWriter, r *http.Request) {
	stateID := r.URL.Query().Get("state_id")
	count, err := h.uc.CountTotalStudents(r.Context(), stateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, count)
}

// ─────────────────────────────────────────────────────────────────────────────
// RESULT HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type ResultHandler struct{ uc *service.ResultService }

func NewResultHandler(uc *service.ResultService) *ResultHandler { return &ResultHandler{uc: uc} }

// @Summary Upsert score
// @Description Create or update a student score
// @Tags results
// @Accept json
// @Produce json
// @Param request body dto.UpsertScoreRequest true "Score info"
// @Success 200 {object} domain.ScoreSheet
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /results/scores [post]
func (h *ResultHandler) UpsertScore(w http.ResponseWriter, r *http.Request) {
	var req dto.UpsertScoreRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	ss, err := h.uc.UpsertScore(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, ss)
}

// @Summary Bulk upsert scores
// @Description Bulk create or update student scores
// @Tags results
// @Accept json
// @Produce json
// @Param request body dto.BulkUpsertScoreRequest true "Bulk score info"
// @Success 200 {object} map[string]any
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /results/scores/bulk [post]
func (h *ResultHandler) BulkUpsertScores(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkUpsertScoreRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	results, errs := h.uc.BulkUpsertScores(r.Context(), claims.StateID, req, claims.UserID)
	resp := map[string]any{"saved": len(results), "errors": len(errs), "data": results}
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		resp["error_messages"] = msgs
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary Get student scores
// @Description Get scores for a student
// @Tags results
// @Produce json
// @Param studentId path string true "Student ID"
// @Param session_id query string false "Filter by session ID"
// @Success 200 {object} []domain.ScoreSheet
// @Router /students/{studentId}/scores [get]
func (h *ResultHandler) GetStudentScores(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "studentId")
	sessionID := r.URL.Query().Get("session_id")
	scores, err := h.uc.GetStudentScores(r.Context(), studentID, sessionID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, scores)
}

// @Summary Compute positions
// @Description Compute student positions in class
// @Tags results
// @Produce json
// @Param term_id query string false "Term ID"
// @Param sub_level_id query string false "Sub-level ID"
// @Param subject_id query string false "Subject ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /results/scores/compute-positions [post]
func (h *ResultHandler) ComputePositions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := h.uc.ComputePositions(r.Context(), q.Get("term_id"), q.Get("sub_level_id"), q.Get("subject_id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, map[string]string{"message": "positions computed"})
}

// ComputePositionsBulk handles POST /results/scores/compute-positions-bulk
// It re-ranks all students in a class arm across all subjects for a given term.
//
//	@Summary		Compute class positions for all subjects
//	@Description	Re-rank all students in a sub-level across all subjects for a given term
//	@Tags			results
//	@Produce		json
//	@Param			term_id		query		string	true	"Term ID"
//	@Param			sub_level_id	query		string	true	"Sub-level (class arm) ID"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	apperror.AppError
//	@Failure		401	{object}	apperror.AppError
//	@Failure		403	{object}	apperror.AppError
//	@Router			/results/scores/compute-positions-bulk [post]
func (h *ResultHandler) ComputePositionsBulk(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := h.uc.ComputePositionsBulk(r.Context(), q.Get("term_id"), q.Get("sub_level_id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, map[string]string{"message": "positions computed for all subjects"})
}

// ComputeClassSubjectStats handles GET /results/scores/class-stats
// It returns per-subject highest, lowest, and average scores for a sub-level/term.
//
//	@Summary		Get class subject stats
//	@Description	Return highest, lowest, and average scores per subject for a class arm
//	@Tags			results
//	@Produce		json
//	@Param			term_id		query		string	true	"Term ID"
//	@Param			sub_level_id	query		string	true	"Sub-level (class arm) ID"
//	@Success		200	{array}		domain.ClassSubjectStat
//	@Failure		400	{object}	apperror.AppError
//	@Failure		401	{object}	apperror.AppError
//	@Failure		403	{object}	apperror.AppError
//	@Router			/results/scores/class-stats [get]
func (h *ResultHandler) ComputeClassSubjectStats(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	stats, err := h.uc.ComputeClassSubjectStats(r.Context(), q.Get("term_id"), q.Get("sub_level_id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, stats)
}

// @Summary Generate report cards
// @Description Generate report cards for a class
// @Tags results
// @Accept json
// @Produce json
// @Param request body dto.GenerateReportCardRequest true "Report card generation info"
// @Success 200 {object} map[string]int
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /results/report-cards/generate [post]
func (h *ResultHandler) GenerateReportCards(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateReportCardRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	count, err := h.uc.GenerateReportCards(r.Context(), claims.StateID, req)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, map[string]int{"generated": count})
}

// @Summary Get report card
// @Description Get a single report card by ID
// @Tags results
// @Produce json
// @Param id path string true "Report Card ID"
// @Success 200 {object} domain.ReportCard
// @Failure 404 {object} apperror.AppError
// @Router /results/report-cards/{id} [get]
func (h *ResultHandler) GetReportCard(w http.ResponseWriter, r *http.Request) {
	rc, err := h.uc.GetReportCard(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, rc)
}

// @Summary Get student report card
// @Description Get a student report card for a term
// @Tags results
// @Produce json
// @Param studentId path string true "Student ID"
// @Param term_id query string false "Term ID"
// @Success 200 {object} domain.ReportCard
// @Failure 404 {object} apperror.AppError
// @Router /students/{studentId}/report-card [get]
func (h *ResultHandler) GetStudentReportCard(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "studentId")
	termID := r.URL.Query().Get("term_id")
	rc, err := h.uc.GetStudentReportCard(r.Context(), studentID, termID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, rc)
}

// @Summary List report cards
// @Description List report cards with optional filters
// @Tags results
// @Produce json
// @Param school_id query string false "Filter by school ID"
// @Param term_id query string false "Filter by term ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} pagination.Meta
// @Router /results/report-cards [get]
func (h *ResultHandler) ListReportCards(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	rcs, total, err := h.uc.ListReportCards(r.Context(), q.Get("school_id"), q.Get("term_id"), p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, rcs, pagination.BuildMeta(p, total))
}

// @Summary Get all student reports
// @Description Get all report cards for a student
// @Tags results
// @Produce json
// @Param studentId path string true "Student ID"
// @Success 200 {object} []domain.ReportCard
// @Router /students/{studentId}/report-cards [get]
func (h *ResultHandler) GetStudentAllReports(w http.ResponseWriter, r *http.Request) {
	rcs, err := h.uc.GetStudentAllReports(r.Context(), chi.URLParam(r, "studentId"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, rcs)
}

// @Summary Update report card remarks
// @Description Update remarks on a report card
// @Tags results
// @Accept json
// @Produce json
// @Param id path string true "Report Card ID"
// @Param request body dto.UpdateReportCardRemarksRequest true "Remarks info"
// @Success 200 {object} domain.ReportCard
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /results/report-cards/{id}/remarks [put]
func (h *ResultHandler) UpdateRemarks(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateReportCardRemarksRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	rc, err := h.uc.UpdateRemarks(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, rc)
}

// @Summary Publish report card
// @Description Publish a report card
// @Tags results
// @Produce json
// @Param id path string true "Report Card ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /results/report-cards/{id}/publish [post]
func (h *ResultHandler) Publish(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Publish(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary Upsert score config
// @Description Create or update score configuration
// @Tags results
// @Accept json
// @Produce json
// @Param request body dto.UpsertScoreConfigRequest true "Score config info"
// @Success 200 {object} domain.ScoreConfig
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /results/score-config [post]
func (h *ResultHandler) UpsertScoreConfig(w http.ResponseWriter, r *http.Request) {
	var req dto.UpsertScoreConfigRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	sc, err := h.uc.UpsertScoreConfig(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, sc)
}

// @Summary Upsert grade config
// @Description Create or update grade configuration
// @Tags results
// @Accept json
// @Produce json
// @Param request body dto.UpsertGradeConfigRequest true "Grade config info"
// @Success 200 {object} domain.GradeConfig
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /results/grade-config [post]
func (h *ResultHandler) UpsertGradeConfig(w http.ResponseWriter, r *http.Request) {
	var req dto.UpsertGradeConfigRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	gc, err := h.uc.UpsertGradeConfig(r.Context(), claims.StateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, gc)
}

// @Summary List grade configs
// @Description List grade configurations
// @Tags results
// @Produce json
// @Param school_id query string false "Filter by school ID"
// @Param level_id query string false "Filter by level ID"
// @Success 200 {object} []domain.GradeConfig
// @Router /results/grade-config [get]
func (h *ResultHandler) ListGradeConfigs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := r.URL.Query().Get("school_id")
	levelID := r.URL.Query().Get("level_id")
	configs, err := h.uc.ListGradeConfigs(r.Context(), schoolID, levelID, claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, configs)
}

// @Summary Delete grade config
// @Description Delete a grade configuration
// @Tags results
// @Produce json
// @Param id path string true "Grade Config ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /results/grade-config/{id} [delete]
func (h *ResultHandler) DeleteGradeConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.DeleteGradeConfig(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// AVATAR HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type AvatarHandler struct {
	avatarService *service.AvatarService
}

func NewAvatarHandler(s *service.AvatarService) *AvatarHandler {
	return &AvatarHandler{avatarService: s}
}

// POST /api/v1/avatar/personnel
// @Summary Upload personnel avatar
// @Description Upload a profile picture for personnel
// @Tags avatar
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Personnel ID"
// @Param file formData file true "Image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /avatar/personnel/{id} [put]
func (h *AvatarHandler) UploadPersonnelAvatar(w http.ResponseWriter, r *http.Request) {
	personnelID := chi.URLParam(r, "id")

	file, header, err := r.FormFile("file")
	if err != nil {
		presenter.Error(w, apperror.BadRequest("file is required"))
		return
	}
	defer file.Close()

	url, err := h.avatarService.UploadPersonnelAvatar(r.Context(), personnelID, file, header.Filename)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, map[string]string{"avatar_url": url})
}

// POST /api/v1/avatar/student
func (h *AvatarHandler) UploadStudentAvatar(w http.ResponseWriter, r *http.Request) {
	studentID := r.FormValue("student_id")

	file, header, err := r.FormFile("file")
	if err != nil {
		presenter.Error(w, apperror.BadRequest("file is required"))
		return
	}
	defer file.Close()

	url, err := h.avatarService.UploadStudentAvatar(r.Context(), studentID, file, header.Filename)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, map[string]string{"avatar_url": url})
}

// ─────────────────────────────────────────────────────────────────────────────
// ATTENDANCE HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type AttendanceHandler struct{ uc *service.AttendanceService }

func NewAttendanceHandler(uc *service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{uc: uc}
}

// ── Personnel Attendance ────────────────────────────────────────────────────

// @Summary Record personnel attendance
// @Description Record attendance for personnel
// @Tags attendance
// @Accept json
// @Produce json
// @Param request body dto.PersonnelAttendanceRequest true "Attendance info"
// @Success 201 {object} dto.PersonnelAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /attendance/personnel [post]
func (h *AttendanceHandler) RecordPersonnelAttendance(w http.ResponseWriter, r *http.Request) {
	var req dto.PersonnelAttendanceRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	resp, err := h.uc.RecordPersonnelAttendance(r.Context(), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, resp)
}

// @Summary Get personnel attendance
// @Description Get a personnel attendance record by ID
// @Tags attendance
// @Produce json
// @Param id path string true "Attendance ID"
// @Success 200 {object} dto.PersonnelAttendanceResponse
// @Failure 404 {object} apperror.AppError
// @Router /attendance/personnel/{id} [get]
func (h *AttendanceHandler) GetPersonnelAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.uc.GetPersonnelAttendance(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary Update personnel attendance
// @Description Update personnel attendance record
// @Tags attendance
// @Accept json
// @Produce json
// @Param id path string true "Attendance ID"
// @Param request body dto.UpdatePersonnelAttendanceRequest true "Updated attendance info"
// @Success 200 {object} dto.PersonnelAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /attendance/personnel/{id} [put]
func (h *AttendanceHandler) UpdatePersonnelAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdatePersonnelAttendanceRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	resp, err := h.uc.UpdatePersonnelAttendance(r.Context(), id, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary Delete personnel attendance
// @Description Delete a personnel attendance record
// @Tags attendance
// @Produce json
// @Param id path string true "Attendance ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /attendance/personnel/{id} [delete]
func (h *AttendanceHandler) DeletePersonnelAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.DeletePersonnelAttendance(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List personnel attendance by school and date
// @Description List personnel attendance for a school on a specific date
// @Tags attendance
// @Produce json
// @Param school_id query string true "School ID"
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} []dto.PersonnelAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Router /attendance/personnel/school [get]
func (h *AttendanceHandler) ListPersonnelAttendanceBySchoolAndDate(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("school_id")
	dateStr := r.URL.Query().Get("date")
	if schoolID == "" || dateStr == "" {
		presenter.Error(w, apperror.BadRequest("school_id and date are required"))
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid date format, expected YYYY-MM-DD"))
		return
	}
	resp, err := h.uc.ListPersonnelAttendanceBySchoolAndDate(r.Context(), schoolID, date)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary List personnel attendance by range
// @Description List personnel attendance for a date range
// @Tags attendance
// @Produce json
// @Param id path string true "Personnel ID"
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} []dto.PersonnelAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Router /attendance/personnel/{id}/range [get]
func (h *AttendanceHandler) ListPersonnelAttendanceByPersonnelAndRange(w http.ResponseWriter, r *http.Request) {
	personnelID := chi.URLParam(r, "id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		presenter.Error(w, apperror.BadRequest("from and to query parameters are required"))
		return
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid from date format, expected YYYY-MM-DD"))
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid to date format, expected YYYY-MM-DD"))
		return
	}
	resp, err := h.uc.ListPersonnelAttendanceByPersonnelAndRange(r.Context(), personnelID, from, to)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// ── Student Attendance ──────────────────────────────────────────────────────

// @Summary Record student attendance
// @Description Record attendance for a student
// @Tags attendance
// @Accept json
// @Produce json
// @Param request body dto.StudentAttendanceRequest true "Attendance info"
// @Success 201 {object} dto.StudentAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /attendance/students [post]
func (h *AttendanceHandler) RecordStudentAttendance(w http.ResponseWriter, r *http.Request) {
	var req dto.StudentAttendanceRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	resp, err := h.uc.RecordStudentAttendance(r.Context(), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, resp)
}

// @Summary Bulk record student attendance
// @Description Bulk record attendance for students
// @Tags attendance
// @Accept json
// @Produce json
// @Param request body dto.BulkStudentAttendanceRequest true "Bulk attendance info"
// @Success 201 {object} map[string]any
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /attendance/students/bulk [post]
func (h *AttendanceHandler) BulkRecordStudentAttendance(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkStudentAttendanceRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	resp, err := h.uc.BulkRecordStudentAttendance(r.Context(), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, map[string]any{"records": resp})
}

// @Summary Get student attendance
// @Description Get a student attendance record by ID
// @Tags attendance
// @Produce json
// @Param id path string true "Attendance ID"
// @Success 200 {object} dto.StudentAttendanceResponse
// @Failure 404 {object} apperror.AppError
// @Router /attendance/students/{id} [get]
func (h *AttendanceHandler) GetStudentAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.uc.GetStudentAttendance(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary Update student attendance
// @Description Update student attendance record
// @Tags attendance
// @Accept json
// @Produce json
// @Param id path string true "Attendance ID"
// @Param request body dto.UpdateStudentAttendanceRequest true "Updated attendance info"
// @Success 200 {object} dto.StudentAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /attendance/students/{id} [put]
func (h *AttendanceHandler) UpdateStudentAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateStudentAttendanceRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	resp, err := h.uc.UpdateStudentAttendance(r.Context(), id, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary Delete student attendance
// @Description Delete a student attendance record
// @Tags attendance
// @Produce json
// @Param id path string true "Attendance ID"
// @Success 204
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Router /attendance/students/{id} [delete]
func (h *AttendanceHandler) DeleteStudentAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.DeleteStudentAttendance(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// @Summary List student attendance by school and date
// @Description List student attendance for a school on a specific date
// @Tags attendance
// @Produce json
// @Param school_id query string true "School ID"
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} []dto.StudentAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Router /attendance/students/school [get]
func (h *AttendanceHandler) ListStudentAttendanceBySchoolAndDate(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("school_id")
	dateStr := r.URL.Query().Get("date")
	if schoolID == "" || dateStr == "" {
		presenter.Error(w, apperror.BadRequest("school_id and date are required"))
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid date format, expected YYYY-MM-DD"))
		return
	}
	resp, err := h.uc.ListStudentAttendanceBySchoolAndDate(r.Context(), schoolID, date)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary List student attendance by range
// @Description List student attendance for a date range
// @Tags attendance
// @Produce json
// @Param id path string true "Student ID"
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} []dto.StudentAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Router /attendance/students/student/{id}/range [get]
func (h *AttendanceHandler) ListStudentAttendanceByStudentAndRange(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		presenter.Error(w, apperror.BadRequest("from and to query parameters are required"))
		return
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid from date format, expected YYYY-MM-DD"))
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid to date format, expected YYYY-MM-DD"))
		return
	}
	resp, err := h.uc.ListStudentAttendanceByStudentAndRange(r.Context(), studentID, from, to)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// @Summary List student attendance by school and range
// @Description List student attendance for a school in a date range
// @Tags attendance
// @Produce json
// @Param school_id query string true "School ID"
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} []dto.StudentAttendanceResponse
// @Failure 400 {object} apperror.AppError
// @Router /attendance/students/school/range [get]
func (h *AttendanceHandler) ListStudentAttendanceBySchoolAndRange(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("school_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if schoolID == "" || fromStr == "" || toStr == "" {
		presenter.Error(w, apperror.BadRequest("school_id, from and to query parameters are required"))
		return
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid from date format, expected YYYY-MM-DD"))
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		presenter.Error(w, apperror.BadRequest("invalid to date format, expected YYYY-MM-DD"))
		return
	}
	resp, err := h.uc.ListStudentAttendanceBySchoolAndRange(r.Context(), schoolID, from, to)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

// PredictionHandler serves prediction endpoints.
type PredictionHandler struct{ uc *service.PredictionService }

func NewPredictionHandler(uc *service.PredictionService) *PredictionHandler {
	return &PredictionHandler{uc: uc}
}

// SchoolReport godoc
// GET /api/v1/predictions/schools/{schoolId}
//
// Returns a school-level performance prediction only (fast path — no per-student loop).
// Query param: session_id (optional) — scope to a specific academic session.
//
// Required permission: results:read
// @Summary School prediction report
// @Description Get school-level performance prediction
// @Tags predictions
// @Produce json
// @Param schoolId path string true "School ID"
// @Param session_id query string false "Session ID"
// @Success 200 {object} domain.SchoolPrediction
// @Failure 400 {object} apperror.AppError
// @Router /predictions/schools/{schoolId} [get]
func (h *PredictionHandler) SchoolReport(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	if schoolID == "" {
		presenter.Error(w, apperror.BadRequest("school_id is required"))
		return
	}

	pred, err := h.uc.SchoolOnlyReport(schoolID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, pred)
}

// FullReport godoc
// GET /api/v1/predictions/schools/{schoolId}/full
//
// Returns the school prediction plus individual predictions for every
// actively enrolled student. Can be slow for large schools — use
// SchoolReport for summary cards, this for the detailed drill-down view.
//
// Query params:
//
//	session_id (optional) — scope enrollments to a specific session
//
// Required permission: results:read
// @Summary Full prediction report
// @Description Get full school prediction report
// @Tags predictions
// @Produce json
// @Param schoolId path string true "School ID"
// @Success 200 {object} domain.PredictionReport
// @Failure 400 {object} apperror.AppError
// @Router /predictions/schools/{schoolId}/full [get]
func (h *PredictionHandler) FullReport(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	if schoolID == "" {
		presenter.Error(w, apperror.BadRequest("school_id is required"))
		return
	}

	sessionID := r.URL.Query().Get("session_id")

	// Confirm the caller is scoped to this school or is a state-level user.
	claims := middleware.ClaimsFromCtx(r.Context())
	if claims.SchoolID != "" && claims.SchoolID != schoolID {
		presenter.Error(w, apperror.Forbidden("you can only access predictions for your own school"))
		return
	}

	report, err := h.uc.GenerateReport(schoolID, sessionID)
	if err != nil {
		logger.FromContext(r.Context()).Error("prediction generate report failed", "school_id", schoolID, "session_id", sessionID, "error", err)
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, report)
}

// StudentReport godoc
// GET /api/v1/predictions/schools/{schoolId}/students/{studentId}
//
// Returns a prediction for a single student, useful for drilling into
// one record from the full report without regenerating everything.
//
// Required permission: results:read
// @Summary Student prediction report
// @Description Get prediction report for a student
// @Tags predictions
// @Produce json
// @Param schoolId path string true "School ID"
// @Param studentId path string true "Student ID"
// @Success 200 {object} domain.StudentPrediction
// @Failure 400 {object} apperror.AppError
// @Router /predictions/schools/{schoolId}/students/{studentId} [get]
func (h *PredictionHandler) StudentReport(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	studentID := chi.URLParam(r, "studentId")

	if schoolID == "" || studentID == "" {
		presenter.Error(w, apperror.BadRequest("school_id and student_id are required"))
		return
	}

	// Generate the full report and find the student in it.
	// (A dedicated single-student path would be a further optimisation.)
	report, err := h.uc.GenerateReport(schoolID, "")
	if err != nil {
		presenter.Error(w, err)
		return
	}

	for _, sp := range report.Students {
		if sp.StudentID == studentID {
			presenter.JSON(w, http.StatusOK, sp)
			return
		}
	}

	presenter.Error(w, apperror.NotFound("student prediction", studentID))
}

// ─────────────────────────────────────────────────────────────────────────────
// RECOMMENDATION HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type RecommendationHandler struct {
	uc *service.RecommendationService
}

func NewRecommendationHandler(uc *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{uc: uc}
}

// @Summary Get recommendations
// @Description Get course recommendations
// @Tags recommendations
// @Produce json
// @Success 200 {object} []ml.RecommendResponse
// @Failure 401 {object} apperror.AppError
// @Failure 403 {object} apperror.AppError
// @Router /recommendations [get]
func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	if claims == nil {
		presenter.Error(w, apperror.Unauthorized("unauthorized"))
		return
	}

	result, err := h.uc.GetRecommendations(r.Context())
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, result)
}
