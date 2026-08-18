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

// GetZoneSummaryReport returns the zone summary report for the active session of the given school,
// scoped to the school's state.
func (s *ZonalSummaryService) GetZoneSummaryReport(ctx context.Context, schoolID, stateID string) ([]domain.ZoneSummaryReport, error) {
	session, err := s.sessionRepo.GetActive(ctx, schoolID)
	if err != nil {
		return nil, err
	}

	return s.zonalReportRepo.GetZoneSummaryReport(ctx, session.ID, stateID)
}

// GetZoneSummaryReportByState returns the zone summary report for the active session across the given state.
func (s *ZonalSummaryService) GetZoneSummaryReportByState(ctx context.Context, stateID string) ([]domain.ZoneSummaryReport, error) {
	session, err := s.sessionRepo.GetActiveForState(ctx, stateID)
	if err != nil {
		return nil, err
	}

	return s.zonalReportRepo.GetZoneSummaryReport(ctx, session.ID, stateID)
}
