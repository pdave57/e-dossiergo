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
func (h *ReportHandler) GetPublicTeachingPersonnel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	data, err := h.reportUC.GetTotalTeachingPersonnel(ctx)
	if err != nil {
		presenter.Error(w, err)
		return
	}

	presenter.JSON(w, http.StatusOK, data)
}
