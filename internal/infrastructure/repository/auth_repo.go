package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// USER REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type userRepo struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) domain.UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	u.Status = domain.UserStatusActive

	q := `INSERT INTO users
		(id,state_id,school_id,email,password_hash,first_name,last_name,status,created_at,updated_at,created_by)
		VALUES ($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.ExecContext(ctx, q,
		u.ID, u.StateID, u.SchoolID, u.Email, u.PasswordHash,
		u.FirstName, u.LastName, u.Status,
		u.CreatedAt, u.UpdatedAt, u.CreatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("email already registered")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	q := `SELECT id,state_id,COALESCE(school_id,''),email,password_hash,
		         first_name,last_name,status,last_login_at,created_at,updated_at,
		         COALESCE(created_by,''),COALESCE(updated_by,'')
		  FROM users WHERE id=$1 AND deleted_at IS NULL`
	return scanUser(r.db.QueryRowContext(ctx, q, id))
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := `SELECT id,state_id,COALESCE(school_id,''),email,password_hash,
		         first_name,last_name,status,last_login_at,created_at,updated_at,
		         COALESCE(created_by,''),COALESCE(updated_by,'')
		  FROM users WHERE email=$1 AND deleted_at IS NULL`
	return scanUser(r.db.QueryRowContext(ctx, q, email))
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now()
	q := `UPDATE users SET school_id=NULLIF($1,''),first_name=$2,last_name=$3,
		  status=$4,updated_at=$5,updated_by=$6 WHERE id=$7 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q,
		u.SchoolID, u.FirstName, u.LastName, u.Status, u.UpdatedAt, u.UpdatedBy, u.ID)
	return checkRowsAffected(res, err, "user", u.ID)
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	q := `UPDATE users SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	return checkRowsAffected(res, err, "user", id)
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at=NOW() WHERE id=$1`, userID)
	return err
}

func (r *userRepo) List(ctx context.Context, f domain.UserFilter, p pagination.Params) ([]*domain.User, int, error) {
	where, args := []string{"deleted_at IS NULL"}, []any{}
	idx := 1
	if f.StateID != "" {
		where = append(where, fmt.Sprintf("state_id=$%d", idx))
		args = append(args, f.StateID); idx++
	}
	if f.SchoolID != "" {
		where = append(where, fmt.Sprintf("school_id=$%d", idx))
		args = append(args, f.SchoolID); idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", idx))
		args = append(args, f.Status); idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf(
			"(email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+f.Search+"%"); idx++
	}

	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users "+clause, args...).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}

	args = append(args, p.PerPage, p.Offset)
	q := fmt.Sprintf(`SELECT id,state_id,COALESCE(school_id,''),email,password_hash,
		first_name,last_name,status,last_login_at,created_at,updated_at,created_by,updated_by
		FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, clause, idx, idx+1)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

// scanUser works on both *sql.Row and *sql.Rows via the scanner interface.
func scanUser(s scanner) (*domain.User, error) {
	u := &domain.User{}
	err := s.Scan(
		&u.ID, &u.StateID, &u.SchoolID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Status, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("user", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return u, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ROLE REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type roleRepo struct{ db *sql.DB }

func NewRoleRepository(db *sql.DB) domain.RoleRepository { return &roleRepo{db: db} }

func (r *roleRepo) Create(ctx context.Context, role *domain.Role) error {
	if role.ID == "" {
		role.ID = uuid.NewString()
	}
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now
	q := `INSERT INTO roles (id,state_id,name,code,description,is_system,created_at,updated_at,created_by)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.ExecContext(ctx, q,
		role.ID, role.StateID, role.Name, role.Code, role.Description,
		role.IsSystem, role.CreatedAt, role.UpdatedAt, role.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("role code already exists in this state")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *roleRepo) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	q := `SELECT id,state_id,name,code,COALESCE(description,''),is_system,created_at,updated_at,
		         COALESCE(created_by,''),COALESCE(updated_by,'')
		  FROM roles WHERE id=$1 AND deleted_at IS NULL`
	return scanRole(r.db.QueryRowContext(ctx, q, id))
}

func (r *roleRepo) GetByCode(ctx context.Context, code string) (*domain.Role, error) {
	q := `SELECT id,state_id,name,code,COALESCE(description,''),is_system,created_at,updated_at,
		         COALESCE(created_by,''),COALESCE(updated_by,'')
		  FROM roles WHERE code=$1 AND deleted_at IS NULL`
	return scanRole(r.db.QueryRowContext(ctx, q, code))
}

func (r *roleRepo) Update(ctx context.Context, role *domain.Role) error {
	role.UpdatedAt = time.Now()
	q := `UPDATE roles SET name=$1,description=$2,updated_at=$3,updated_by=$4 WHERE id=$5 AND is_system=FALSE AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, role.Name, role.Description, role.UpdatedAt, role.UpdatedBy, role.ID)
	return checkRowsAffected(res, err, "role", role.ID)
}

func (r *roleRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE roles SET deleted_at=NOW() WHERE id=$1 AND is_system=FALSE AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "role", id)
}

func (r *roleRepo) List(ctx context.Context, stateID string) ([]*domain.Role, error) {
	q := `SELECT id,state_id,name,code,COALESCE(description,''),is_system,created_at,updated_at,
		         COALESCE(created_by,''),COALESCE(updated_by,'')
		  FROM roles WHERE state_id=$1 AND deleted_at IS NULL ORDER BY name`
	rows, err := r.db.QueryContext(ctx, q, stateID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

func (r *roleRepo) AddPermission(ctx context.Context, rp *domain.RolePermission) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO role_permissions (role_id,permission_id,granted_at,granted_by)
		 VALUES ($1,$2,NOW(),$3) ON CONFLICT DO NOTHING`,
		rp.RoleID, rp.PermissionID, rp.GrantedBy)
	return err
}

func (r *roleRepo) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM role_permissions WHERE role_id=$1 AND permission_id=$2`, roleID, permissionID)
	return err
}

func (r *roleRepo) GetPermissions(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	q := `SELECT p.id,p.resource,p.action,COALESCE(p.description,'')
		  FROM permissions p
		  JOIN role_permissions rp ON rp.permission_id=p.id
		  WHERE rp.role_id=$1 ORDER BY p.resource,p.action`
	return queryPermissions(r.db, ctx, q, roleID)
}

func (r *roleRepo) GetPermissionsForUser(ctx context.Context, userID string) ([]*domain.Permission, error) {
	q := `SELECT DISTINCT p.id,p.resource,p.action,COALESCE(p.description,'')
		  FROM permissions p
		  JOIN role_permissions rp ON rp.permission_id=p.id
		  JOIN user_roles ur ON ur.role_id=rp.role_id
		  WHERE ur.user_id=$1`
	return queryPermissions(r.db, ctx, q, userID)
}

func queryPermissions(db *sql.DB, ctx context.Context, q, arg string) ([]*domain.Permission, error) {
	rows, err := db.QueryContext(ctx, q, arg)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Permission
	for rows.Next() {
		p := &domain.Permission{}
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action, &p.Description); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanRole(s scanner) (*domain.Role, error) {
	r := &domain.Role{}
	err := s.Scan(&r.ID, &r.StateID, &r.Name, &r.Code, &r.Description,
		&r.IsSystem, &r.CreatedAt, &r.UpdatedAt, &r.CreatedBy, &r.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("role", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return r, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PERMISSION REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type permRepo struct{ db *sql.DB }

func NewPermissionRepository(db *sql.DB) domain.PermissionRepository { return &permRepo{db: db} }

func (r *permRepo) Create(ctx context.Context, p *domain.Permission) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO permissions (id,resource,action,description) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (resource,action) DO NOTHING`,
		p.ID, p.Resource, p.Action, p.Description)
	return err
}

func (r *permRepo) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	p := &domain.Permission{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,resource,action,COALESCE(description,'') FROM permissions WHERE id=$1`, id).
		Scan(&p.ID, &p.Resource, &p.Action, &p.Description)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("permission", id)
	}
	return p, err
}

func (r *permRepo) List(ctx context.Context) ([]*domain.Permission, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,resource,action,COALESCE(description,'') FROM permissions ORDER BY resource,action`)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Permission
	for rows.Next() {
		p := &domain.Permission{}
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action, &p.Description); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *permRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM permissions WHERE id=$1`, id)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// USER ROLE REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type userRoleRepo struct{ db *sql.DB }

func NewUserRoleRepository(db *sql.DB) domain.UserRoleRepository { return &userRoleRepo{db: db} }

func (r *userRoleRepo) Assign(ctx context.Context, ur *domain.UserRole) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_roles (user_id,role_id,school_id,assigned_at,assigned_by)
		 VALUES ($1,$2,NULLIF($3,''),NOW(),$4) ON CONFLICT (user_id,role_id) DO NOTHING`,
		ur.UserID, ur.RoleID, ur.SchoolID, ur.AssignedBy)
	return err
}

func (r *userRoleRepo) Revoke(ctx context.Context, userID, roleID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_roles WHERE user_id=$1 AND role_id=$2`, userID, roleID)
	return err
}

func (r *userRoleRepo) GetRolesForUser(ctx context.Context, userID string) ([]*domain.Role, error) {
	q := `SELECT r.id,r.state_id,r.name,r.code,COALESCE(r.description,''),r.is_system,
		         r.created_at,r.updated_at,COALESCE(r.created_by,''),COALESCE(r.updated_by,'')
		  FROM roles r
		  JOIN user_roles ur ON ur.role_id=r.id
		  WHERE ur.user_id=$1 AND r.deleted_at IS NULL`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

func (r *userRoleRepo) HasRole(ctx context.Context, userID, roleCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
		   SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id
		   WHERE ur.user_id=$1 AND r.code=$2 AND r.deleted_at IS NULL
		 )`, userID, roleCode).Scan(&exists)
	return exists, err
}

func (r *userRoleRepo) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
		  SELECT 1
		  FROM user_roles ur
		  JOIN role_permissions rp ON rp.role_id = ur.role_id
		  JOIN permissions p       ON p.id = rp.permission_id
		  WHERE ur.user_id=$1 AND p.resource=$2 AND p.action=$3
		)`, userID, resource, action).Scan(&exists)
	return exists, err
}

// ─────────────────────────────────────────────────────────────────────────────
// REFRESH TOKEN REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type refreshTokenRepo struct{ db *sql.DB }

func NewRefreshTokenRepository(db *sql.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Save(ctx context.Context, rt *domain.RefreshToken) error {
	if rt.ID == "" {
		rt.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id,user_id,token_hash,expires_at,created_at,revoked)
		 VALUES ($1,$2,$3,$4,NOW(),FALSE)
		 ON CONFLICT (user_id) DO UPDATE SET token_hash=$3,expires_at=$4,revoked=FALSE`,
		rt.ID, rt.UserID, rt.TokenHash, rt.ExpiresAt)
	return err
}

func (r *refreshTokenRepo) GetByUserID(ctx context.Context, userID string) (*domain.RefreshToken, error) {
	rt := &domain.RefreshToken{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,user_id,token_hash,expires_at,created_at,revoked
		 FROM refresh_tokens WHERE user_id=$1 AND revoked=FALSE AND expires_at>NOW()`, userID).
		Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked)
	if err == sql.ErrNoRows {
		return nil, apperror.Unauthorized("refresh token not found or expired")
	}
	return rt, err
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked=TRUE WHERE user_id=$1`, userID)
	return err
}

func (r *refreshTokenRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at<NOW() OR revoked=TRUE`)
	return err
}
