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

// GetOSCReport returns the out-of-school children report grouped by LGA.
func (s *ReportService) GetOSCReport(ctx context.Context, stateID string) ([]domain.OSCReportRow, error) {
	return s.reportRepo.GetOSCReport(ctx, stateID)
}

// GetOSCChartData returns chart-ready OSC data grouped by LGA.
func (s *ReportService) GetOSCChartData(ctx context.Context, stateID string) ([]domain.OSCChartPoint, error) {
	return s.reportRepo.GetOSCChartData(ctx, stateID)
}
