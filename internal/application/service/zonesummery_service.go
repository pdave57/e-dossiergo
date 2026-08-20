package service

import (
	"context"

	"github.com/edossier/api/internal/domain"
)

// ZonalSummaryService orchestrates zonal summary report data gathering.
type ZonalSummaryService struct {
	zonalReportRepo domain.ZonalReportRepository
	sessionRepo     domain.AcademicSessionRepository
}

// NewZonalSummaryService creates a new ZonalSummaryService.
func NewZonalSummaryService(zonalReportRepo domain.ZonalReportRepository, sessionRepo domain.AcademicSessionRepository) *ZonalSummaryService {
	return &ZonalSummaryService{
		zonalReportRepo: zonalReportRepo,
		sessionRepo:     sessionRepo,
	}
}

// GetZoneSummaryReport returns the zone summary report scoped to the given state.
func (s *ZonalSummaryService) GetZoneSummaryReport(ctx context.Context, schoolID, stateID string) ([]domain.ZoneSummaryReport, error) {
	return s.zonalReportRepo.GetZoneSummaryReport(ctx, "", stateID)
}

// GetZoneSummaryReportByState returns the zone summary report for the given state.
func (s *ZonalSummaryService) GetZoneSummaryReportByState(ctx context.Context, stateID string) ([]domain.ZoneSummaryReport, error) {
	return s.zonalReportRepo.GetZoneSummaryReport(ctx, "", stateID)
}
