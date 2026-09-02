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

func (r *postgresReportRepository) GetOSCReport(ctx context.Context, stateID string) ([]domain.OSCReportRow, error) {
	stateFilter := ""
	args := []interface{}{}
	if stateID != "" {
		stateFilter = "AND l.state_id = $1"
		args = append(args, stateID)
	}

	query := fmt.Sprintf(`
		WITH lga_school_students AS (
			SELECT
				sc.lga_id,
				COALESCE(SUM(sc.total_students), 0) AS school_student_count
			FROM schools sc
			WHERE sc.deleted_at IS NULL
			GROUP BY sc.lga_id
		),
		lga_enrollments AS (
			SELECT
				sc.lga_id,
				COUNT(DISTINCT e.student_id) AS enrollment_count
			FROM enrollments e
			JOIN schools sc ON sc.id = e.school_id
			WHERE e.status = 'ACTIVE'
			  AND sc.deleted_at IS NULL
			GROUP BY sc.lga_id
		)
		SELECT
			l.id AS lga_id,
			l.name AS lga_name,
			l.state_id,
			COALESCE(ROUND(lp.population_4_14 * POW(1.0 + lp.annual_growth_rate,
				EXTRACT(YEAR FROM NOW())::INT - lp.base_year)), 0) AS sep,
			COALESCE(le.enrollment_count, 0) AS enrollment_count,
			COALESCE(ls.school_student_count, 0) AS school_student_count,
			COALESCE(le.enrollment_count, 0) + COALESCE(ls.school_student_count, 0) AS tcs
		FROM lgas l
		LEFT JOIN lga_population_profiles lp ON lp.lga_id = l.id AND lp.deleted_at IS NULL
		LEFT JOIN lga_enrollments le ON le.lga_id = l.id
		LEFT JOIN lga_school_students ls ON ls.lga_id = l.id
		WHERE l.deleted_at IS NULL %s
		ORDER BY l.name
	`, stateFilter)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query osc report: %w", err)
	}
	defer rows.Close()

	var results []domain.OSCReportRow
	for rows.Next() {
		var row domain.OSCReportRow
		err := rows.Scan(
			&row.LGAID, &row.LGAName, &row.StateID,
			&row.SEP, &row.EnrollmentCount, &row.SchoolStudentCount, &row.TCS,
		)
		if err != nil {
			return nil, fmt.Errorf("scan osc report row: %w", err)
		}
		row.OSC = row.SEP - row.TCS
		if row.SEP > 0 {
			row.OSCPct = (float64(row.OSC) / float64(row.SEP)) * 100
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

func (r *postgresReportRepository) GetOSCChartData(ctx context.Context, stateID string) ([]domain.OSCChartPoint, error) {
	reportRows, err := r.GetOSCReport(ctx, stateID)
	if err != nil {
		return nil, err
	}

	chartData := make([]domain.OSCChartPoint, len(reportRows))
	for i, row := range reportRows {
		chartData[i] = domain.OSCChartPoint{
			LGAName: row.LGAName,
			SEP:     row.SEP,
			TCS:     row.TCS,
			OSC:     row.OSC,
			OSCPct:  row.OSCPct,
		}
	}

	return chartData, nil
}

