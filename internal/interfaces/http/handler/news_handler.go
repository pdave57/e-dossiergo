package handler

import (
	"net/http"
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

// NewsHandler handles HTTP requests for news items and announcements.
//
// List and GetByID are public; Create, Update and Delete are wrapped at the
// router with Authenticate + Authorize("news", …).
type NewsHandler struct {
	uc *service.NewsService
}

// NewNewsHandler returns a new NewsHandler.
func NewNewsHandler(uc *service.NewsService) *NewsHandler {
	return &NewsHandler{uc: uc}
}

// List handles GET /news — public, no authentication required.
//
//	@Summary		List news and announcements
//	@Description	Public, paginated feed of news items and announcements, newest first
//	@Tags			news
//	@Produce		json
//	@Param			state_id	query		string	false	"Filter by state ID"
//	@Param			type		query		string	false	"Filter by type"	Enums(NEWS, ANNOUNCEMENT)
//	@Param			search		query		string	false	"Search headline, sub-headline, and body"
//	@Param			from		query		string	false	"Earliest news date (YYYY-MM-DD)"
//	@Param			to			query		string	false	"Latest news date (YYYY-MM-DD)"
//	@Param			page		query		int		false	"Page number"
//	@Param			per_page	query		int		false	"Items per page"
//	@Success		200			{object}	pagination.Meta
//	@Failure		400			{object}	apperror.AppError
//	@Router			/news [get]
func (h *NewsHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()

	f := domain.NewsFilter{
		StateID: q.Get("state_id"),
		Type:    q.Get("type"),
		Search:  q.Get("search"),
	}

	from, err := parseNewsDate(q.Get("from"), "from")
	if err != nil {
		presenter.Error(w, err)
		return
	}
	to, err := parseNewsDate(q.Get("to"), "to")
	if err != nil {
		presenter.Error(w, err)
		return
	}
	f.From, f.To = from, to

	items, total, err := h.uc.List(r.Context(), f, p)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSONList(w, http.StatusOK, items, pagination.BuildMeta(p, total))
}

// GetByID handles GET /news/{id} — public, no authentication required.
//
//	@Summary		Get news item by ID
//	@Description	Public — retrieve a single news item or announcement
//	@Tags			news
//	@Produce		json
//	@Param			id	path		string	true	"News ID"
//	@Success		200	{object}	domain.NewsAnnouncement
//	@Failure		404	{object}	apperror.AppError
//	@Router			/news/{id} [get]
func (h *NewsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	n, err := h.uc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, n)
}

// Create handles POST /news — requires the news:create permission.
//
//	@Summary		Publish news or announcement
//	@Description	Requires the news:create permission. state_id defaults to the caller's state and type defaults to NEWS.
//	@Tags			news
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateNewsRequest	true	"News payload"
//	@Success		201		{object}	domain.NewsAnnouncement
//	@Failure		400		{object}	apperror.AppError
//	@Failure		401		{object}	apperror.AppError
//	@Failure		403		{object}	apperror.AppError
//	@Router			/news [post]
func (h *NewsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateNewsRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	if claims == nil {
		presenter.Error(w, apperror.Unauthorized("unauthorized"))
		return
	}
	// Prefer the state supplied in the request body; fall back to the
	// caller's state from the token (e.g. state admins).
	stateID := req.StateID
	if stateID == "" {
		stateID = claims.StateID
	}
	n, err := h.uc.Create(r.Context(), stateID, req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.Created(w, n)
}

// Update handles PUT /news/{id} — requires the news:update permission.
//
//	@Summary		Update news or announcement
//	@Description	Requires the news:update permission. The owning state cannot be changed.
//	@Tags			news
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"News ID"
//	@Param			request	body		dto.UpdateNewsRequest	true	"News payload"
//	@Success		200		{object}	domain.NewsAnnouncement
//	@Failure		400		{object}	apperror.AppError
//	@Failure		401		{object}	apperror.AppError
//	@Failure		403		{object}	apperror.AppError
//	@Failure		404		{object}	apperror.AppError
//	@Router			/news/{id} [put]
func (h *NewsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateNewsRequest
	if err := presenter.DecodeJSON(r, &req); err != nil {
		presenter.Error(w, err)
		return
	}
	claims := middleware.ClaimsFromCtx(r.Context())
	if claims == nil {
		presenter.Error(w, apperror.Unauthorized("unauthorized"))
		return
	}
	n, err := h.uc.Update(r.Context(), chi.URLParam(r, "id"), req, claims.UserID)
	if err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.JSON(w, http.StatusOK, n)
}

// Delete handles DELETE /news/{id} — requires the news:delete permission.
//
//	@Summary		Delete news or announcement
//	@Description	Requires the news:delete permission. Soft-deletes the record.
//	@Tags			news
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"News ID"
//	@Success		204	"No Content"
//	@Failure		401	{object}	apperror.AppError
//	@Failure		403	{object}	apperror.AppError
//	@Failure		404	{object}	apperror.AppError
//	@Router			/news/{id} [delete]
func (h *NewsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		presenter.Error(w, err)
		return
	}
	presenter.NoContent(w)
}

// parseNewsDate parses an optional YYYY-MM-DD query parameter.
func parseNewsDate(value, field string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, apperror.BadRequest("invalid " + field + " date, expected YYYY-MM-DD")
	}
	return &t, nil
}
