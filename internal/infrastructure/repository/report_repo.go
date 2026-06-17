package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/edossier/api/internal/domain"
)

type postgresReportRepository struct {
	db *sql.DB
}

// NewReportRepository returns a Postgres implementation of ReportRepository.
func NewReportRepository(db *sql.DB) domain.ReportRepository {
	return &postgresReportRepository{db: db}
}

func (r *postgresReportRepository) GetDashboardStats(ctx context.Context, stateID, schoolID string) (*domain.DashboardStats, error) {
	var stats domain.DashboardStats

	// Determine optional school filter
	schoolFilter := ""
	args := []interface{}{stateID}
	
	if schoolID != "" {
		schoolFilter = " AND id = $2"
		args = append(args, schoolID)
	}

	// 1. Total Schools
	qSchools := fmt.Sprintf("SELECT COUNT(*) FROM schools WHERE state_id = $1%s", schoolFilter)
	if err := r.db.QueryRowContext(ctx, qSchools, args...).Scan(&stats.TotalSchools); err != nil {
		return nil, fmt.Errorf("count schools: %w", err)
	}

	// For relations like enrollments/personnel, we need to join or filter on school_id if provided.
	relSchoolFilter := ""
	if schoolID != "" {
		relSchoolFilter = " AND school_id = $2"
	}

	// 2. Total Students (Active enrollments)
	qStudents := fmt.Sprintf(`
		SELECT COUNT(*) FROM enrollments e 
		JOIN schools s ON e.school_id = s.id 
		WHERE s.state_id = $1 AND e.status = 'ACTIVE'%s
	`, relSchoolFilter)
	if err := r.db.QueryRowContext(ctx, qStudents, args...).Scan(&stats.TotalStudents); err != nil {
		return nil, fmt.Errorf("count students: %w", err)
	}

	// 3. Teaching Personnel
	qPersonnel := fmt.Sprintf(`
		SELECT COUNT(*) FROM personnel 
		WHERE state_id = $1 AND role = 'TEACHER' AND status = 'ACTIVE'%s
	`, relSchoolFilter)
	if err := r.db.QueryRowContext(ctx, qPersonnel, args...).Scan(&stats.TeachingPersonnel); err != nil {
		return nil, fmt.Errorf("count personnel: %w", err)
	}

	// 4. Result Completion (Report cards published or generated)
	qResults := fmt.Sprintf(`
		SELECT COUNT(*) FROM report_cards rc 
		JOIN schools s ON rc.school_id = s.id 
		WHERE s.state_id = $1%s
	`, relSchoolFilter)
	if err := r.db.QueryRowContext(ctx, qResults, args...).Scan(&stats.ResultCompletion); err != nil {
		return nil, fmt.Errorf("count results: %w", err)
	}

	return &stats, nil
}

func (r *postgresReportRepository) GetTotalTeachingPersonnel(ctx context.Context) (int, error) {
	var total int
	query := "SELECT COUNT(*) FROM personnel WHERE role = 'TEACHER' AND status = 'ACTIVE'"
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, fmt.Errorf("count total teaching personnel: %w", err)
	}
	return total, nil
}
