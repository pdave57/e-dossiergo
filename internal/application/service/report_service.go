package service

import (
	"context"

	"github.com/edossier/api/internal/domain"
)

// ReportService orchestrates reporting and analytics data gathering.
type ReportService struct {
	reportRepo domain.ReportRepository
}

// NewReportService creates a new ReportService.
func NewReportService(reportRepo domain.ReportRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
	}
}

// GetDashboardStats returns aggregated stats for the dashboard.
func (s *ReportService) GetDashboardStats(ctx context.Context, stateID, schoolID string) (*domain.DashboardStats, error) {
	return s.reportRepo.GetDashboardStats(ctx, stateID, schoolID)
}

// GetTotalTeachingPersonnel returns the total number of active teachers system-wide.
func (s *ReportService) GetTotalTeachingPersonnel(ctx context.Context) (*domain.PublicTeachingPersonnel, error) {
	total, err := s.reportRepo.GetTotalTeachingPersonnel(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.PublicTeachingPersonnel{Total: total}, nil
}
