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
	"github.com/edossier/api/pkg/pagination"
)

// ─────────────────────────────────────────────────────────────────────────────
// AUTH HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type AuthHandler struct{ uc *service.AuthService }

func NewAuthHandler(uc *service.AuthService) *AuthHandler { return &AuthHandler{uc: uc} }

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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	if err := h.uc.Logout(r.Context(), claims.UserID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, user)
}

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

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.Delete(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *UserHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	roleID := chi.URLParam(r, "roleId")
	if err := h.uc.RevokeRole(r.Context(), userID, roleID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	roles, err := h.uc.List(r.Context(), claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, roles)
}

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

func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, role)
}

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

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.Delete(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *RoleHandler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	permID := chi.URLParam(r, "permId")
	if err := h.uc.RemovePermission(r.Context(), roleID, permID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *StateHandler) ListStates(w http.ResponseWriter, r *http.Request) {
	states, err := h.uc.ListStates(r.Context())
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, states)
}

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

func (h *StateHandler) GetState(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetState(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

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

func (h *ZoneHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	zones, err := h.uc.ListZones(r.Context(), chi.URLParam(r, "stateId"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, zones)
}

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

func (h *SchoolHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

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

func (h *SchoolHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

func (h *SchoolHandler) ListFacilities(w http.ResponseWriter, r *http.Request) {
	facs, err := h.uc.ListFacilities(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, facs)
}

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

func (h *SchoolHandler) DeleteFacility(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteFacility(r.Context(), chi.URLParam(r, "facilityId")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

func (h *SchoolHandler) CountTotalSchools(w http.ResponseWriter, r *http.Request) {
	stateID := r.URL.Query().Get("state_id")
	count, err := h.uc.CountTotalSchools(r.Context(), stateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, count)
}

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type AcademicHandler struct{ uc *service.AcademicService }

func NewAcademicHandler(uc *service.AcademicService) *AcademicHandler {
	return &AcademicHandler{uc: uc}
}

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

func (h *AcademicHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetSession(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

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

func (h *AcademicHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteSession(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

func (h *AcademicHandler) ListTerms(w http.ResponseWriter, r *http.Request) {
	terms, err := h.uc.ListTerms(r.Context(), chi.URLParam(r, "sessionId"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, terms)
}

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

func (h *AcademicHandler) ActivateTerm(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	termID := chi.URLParam(r, "id")
	if err := h.uc.ActivateTerm(r.Context(), termID, sessionID); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

func (h *AcademicHandler) DeleteTerm(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteTerm(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// GetTerm returns a single term by id (top-level /terms/{id}).
func (h *AcademicHandler) GetTerm(w http.ResponseWriter, r *http.Request) {
	t, err := h.uc.GetTerm(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, t)
}

// ListAllTerms returns all terms across every session (top-level /terms).
func (h *AcademicHandler) ListAllTerms(w http.ResponseWriter, r *http.Request) {
	terms, err := h.uc.ListAllTerms(r.Context())
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, terms)
}

// CreateTermTopLevel creates a term via the top-level /terms endpoint, where the
// owning session is provided in the request body.
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

func (h *LevelHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	levels, err := h.uc.ListLevels(r.Context(), claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, levels)
}

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

func (h *LevelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	l, err := h.uc.GetLevel(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, l)
}

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

func (h *LevelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteLevel(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *LevelHandler) DeleteSubLevel(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteSubLevel(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *SubjectHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	subjects, err := h.uc.ListByState(r.Context(), claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, subjects)
}

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

func (h *SubjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

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

func (h *SubjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *PersonnelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	p, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, p)
}

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

func (h *PersonnelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *StudentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, s)
}

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

func (h *StudentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *StudentHandler) ListProgressions(w http.ResponseWriter, r *http.Request) {
	lps, err := h.uc.ListProgressions(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, lps)
}

// func (h *AvatarHandler) UploadStudentAvatar(w http.ResponseWriter, r *http.Request) {
//     studentID := chi.URLParam(r, "id")
//     file, _, err := r.FormFile("file")
//     if err != nil {
//         presenter.Error(w, err)
//         return
//     }
//     defer file.Close()
//     claims := middleware.ClaimsFromCtx(r.Context())
//     updated, err := h.uc.UploadAvatar(r.Context(), studentID, file, claims.UserID)
//     if err != nil {
//         presenter.Error(w, err)
//         return
//     }
//     presenter.JSON(w, http.StatusOK, updated)
// }

// ─────────────────────────────────────────────────────────────────────────────
// GENDER HANDLER
// ─────────────────────────────────────────────────────────────────────────────

type GenderHandler struct{ uc service.GenderService }

func NewGenderHandler(uc service.GenderService) *GenderHandler { return &GenderHandler{uc: uc} }

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

func (h *ResultHandler) ComputePositions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := h.uc.ComputePositions(r.Context(), q.Get("term_id"), q.Get("sub_level_id"), q.Get("subject_id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, map[string]string{"message": "positions computed"})
}

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

func (h *ResultHandler) GetReportCard(w http.ResponseWriter, r *http.Request) {
	rc, err := h.uc.GetReportCard(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, rc)
}

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

func (h *ResultHandler) GetStudentAllReports(w http.ResponseWriter, r *http.Request) {
	rcs, err := h.uc.GetStudentAllReports(r.Context(), chi.URLParam(r, "studentId"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, rcs)
}

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

func (h *ResultHandler) Publish(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Publish(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *ResultHandler) ListGradeConfigs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	schoolID := r.URL.Query().Get("school_id")
	configs, err := h.uc.ListGradeConfigs(r.Context(), schoolID, claims.StateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, configs)
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
func (h *AvatarHandler) UploadPersonnelAvatar(w http.ResponseWriter, r *http.Request) {
	// Get User/School context (Assume middleware sets Claims)
	claims := middleware.ClaimsFromCtx(r.Context())

	file, header, err := r.FormFile("file")
	if err != nil {
		presenter.Error(w, apperror.BadRequest("file is required"))
		return
	}
	defer file.Close()

	url, err := h.avatarService.UploadPersonnelAvatar(r.Context(), claims.SchoolID, claims.UserID, file, header.Filename)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, map[string]string{"avatar_url": url})
}

// POST /api/v1/avatar/student
func (h *AvatarHandler) UploadStudentAvatar(w http.ResponseWriter, r *http.Request) {
	// Get StudentID from form or query
	studentID := r.FormValue("student_id")

	file, header, err := r.FormFile("file")
	if err != nil {
		presenter.Error(w, apperror.BadRequest("file is required"))
		return
	}
	defer file.Close()

	claims := middleware.ClaimsFromCtx(r.Context())

	url, err := h.avatarService.UploadStudentAvatar(r.Context(), claims.SchoolID, studentID, file, header.Filename)
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

func (h *AttendanceHandler) GetPersonnelAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.uc.GetPersonnelAttendance(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

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

func (h *AttendanceHandler) DeletePersonnelAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.DeletePersonnelAttendance(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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

func (h *AttendanceHandler) GetStudentAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.uc.GetStudentAttendance(r.Context(), id)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, resp)
}

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

func (h *AttendanceHandler) DeleteStudentAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.uc.DeleteStudentAttendance(r.Context(), id); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

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
