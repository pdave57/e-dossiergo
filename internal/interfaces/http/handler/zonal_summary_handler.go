package handler

import (
	"net/http"

	"github.com/edossier/api/internal/application/service"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/internal/interfaces/http/middleware"
	"github.com/edossier/api/internal/interfaces/presenter"
	"github.com/edossier/api/pkg/apperror"
)

// ZonalSummaryHandler handles HTTP requests for zonal summary reports.
type ZonalSummaryHandler struct {
	zonalSummaryUC *service.ZonalSummaryService
}

// NewZonalSummaryHandler returns a new ZonalSummaryHandler.
func NewZonalSummaryHandler(zonalSummaryUC *service.ZonalSummaryService) *ZonalSummaryHandler {
	return &ZonalSummaryHandler{
		zonalSummaryUC: zonalSummaryUC,
	}
}

// GetZoneSummaryReport handles GET /reports/zonal/summary
// It retrieves the active session from the authenticated user's school (or state) and returns the zone summary report.
//
//	@Summary		Get zonal summary report
//	@Description	Retrieve zone summary report for the current active session
//	@Tags			reports
//	@Produce		json
//	@Param			school_id	query		string	false	"School ID (required for school-level access)"
//	@Param			state_id	query		string	false	"State ID (required for state-level access)"
//	@Success		200	{array}		domain.ZoneSummaryReport
//	@Failure		400	{object}	apperror.AppError
//	@Failure		401	{object}	apperror.AppError
//	@Failure		403	{object}	apperror.AppError
//	@Router			/reports/zonal/summary [get]
func (h *ZonalSummaryHandler) GetZoneSummaryReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims := middleware.ClaimsFromCtx(ctx)
	if claims == nil {
		presenter.Error(w, apperror.Unauthorized("unauthorized"))
		return
	}

	stateID := claims.StateID
	if stateID == "" {
		stateID = r.URL.Query().Get("state_id")
	}
	schoolID := claims.SchoolID
	if schoolID == "" {
		schoolID = r.URL.Query().Get("school_id")
	}

	var reports []domain.ZoneSummaryReport
	var err error

	if schoolID != "" {
		reports, err = h.zonalSummaryUC.GetZoneSummaryReport(ctx, schoolID, stateID)
	} else if stateID != "" {
		reports, err = h.zonalSummaryUC.GetZoneSummaryReportByState(ctx, stateID)
	} else {
		presenter.Error(w, apperror.BadRequest("school_id or state_id is required"))
		return
	}

	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, reports)
}
