package handler

import (
	"net/http"

	"github.com/edossier/api/internal/application/service"
	"github.com/edossier/api/internal/interfaces/http/middleware"
	"github.com/edossier/api/internal/interfaces/presenter"
	"github.com/edossier/api/pkg/apperror"
)

// ReportHandler handles HTTP requests for reports and analytics.
type ReportHandler struct {
	reportUC *service.ReportService
}

// NewReportHandler returns a new ReportHandler.
func NewReportHandler(reportUC *service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportUC: reportUC,
	}
}

// GetDashboardStats handles GET /reports/dashboard
// It retrieves the state_id and optionally school_id from the context to scope the data.
//
//	@Summary		Get dashboard stats
//	@Description	Retrieve dashboard statistics for the authenticated user's state/school
//	@Tags			reports
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		401	{object}	apperror.AppError
//	@Failure		403	{object}	apperror.AppError
//	@Router			/reports/dashboard [get]
func (h *ReportHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims := middleware.ClaimsFromCtx(ctx)
	if claims == nil {
		presenter.Error(w, apperror.Unauthorized("unauthorized"))
		return
	}

	stateID := claims.StateID
	schoolID := claims.SchoolID // Might be empty for State Admins

	stats, err := h.reportUC.GetDashboardStats(ctx, stateID, schoolID)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, stats)
}

// GetPublicTeachingPersonnel handles GET /reports/public/teaching-personnel
//
//	@Summary		Get public teaching personnel count
//	@Description	Get total count of active teaching personnel (public endpoint)
//	@Tags			reports
//	@Produce		json
//	@Success		200	{object}	int
//	@Router			/reports/public/teaching-personnel [get]
func (h *ReportHandler) GetPublicTeachingPersonnel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.reportUC.GetTotalTeachingPersonnel(ctx)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, data)
}

// GetOSCReport handles GET /reports/osc
// It retrieves the out-of-school children report grouped by LGA for the authenticated user's state.
//
//	@Summary		Get out-of-school children report
//	@Description	Retrieve OSC report grouped by LGA using 2006 census data projected to current year
//	@Tags			reports
//	@Produce		json
//	@Success		200	{array}		domain.OSCReportRow
//	@Failure		401	{object}	apperror.AppError
//	@Failure		403	{object}	apperror.AppError
//	@Router			/reports/osc [get]
func (h *ReportHandler) GetOSCReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims := middleware.ClaimsFromCtx(ctx)
	if claims == nil {
		presenter.Error(w, apperror.Unauthorized("unauthorized"))
		return
	}

	stateID := claims.StateID
	data, err := h.reportUC.GetOSCReport(ctx, stateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, data)
}

// GetOSCChart handles GET /reports/osc/chart
// It retrieves chart-ready out-of-school children data grouped by LGA.
//
//	@Summary		Get OSC chart data
//	@Description	Retrieve chart-ready data for visualizing OSC per LGA
//	@Tags			reports
//	@Produce		json
//	@Success		200	{array}		domain.OSCChartPoint
//	@Failure		401	{object}	apperror.AppError
//	@Failure		403	{object}	apperror.AppError
//	@Router			/reports/osc/chart [get]
func (h *ReportHandler) GetOSCChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims := middleware.ClaimsFromCtx(ctx)
	if claims == nil {
		presenter.Error(w, apperror.Unauthorized("unauthorized"))
		return
	}

	stateID := claims.StateID
	data, err := h.reportUC.GetOSCChartData(ctx, stateID)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, data)
}
