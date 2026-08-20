package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/edossier/api/internal/domain"
)

type postgresZonalReportRepository struct {
	db *gorm.DB
}

func NewZonalReportRepository(db *gorm.DB) domain.ZonalReportRepository {
	return &postgresZonalReportRepository{db: db}
}

func (r *postgresZonalReportRepository) GetZoneSummaryReport(ctx context.Context, _sessionID, stateID string) ([]domain.ZoneSummaryReport, error) {
	var results []domain.ZoneSummaryReport

	const (
		roleTeacher            = "TEACHER"
		statusActiveStaff      = "ACTIVE"
		statusActiveStudent    = "ACTIVE"
		statusActiveEnrollment = "ACTIVE"
	)

	args := []interface{}{
		roleTeacher,
		statusActiveStaff,
		statusActiveEnrollment,
		statusActiveStudent,
	}

	stateFilter := ""
	if stateID != "" {
		stateFilter = "AND z.state_id = ?"
		args = append(args, stateID)
	}

	query := fmt.Sprintf(`
		WITH zone_schools AS (
			SELECT
				s.zone_id,
				COUNT(s.id) AS school_count
			FROM schools s
			WHERE s.zone_id IS NOT NULL
			GROUP BY s.zone_id
		),

		zone_teachers AS (
			SELECT
				s.zone_id,
				COUNT(DISTINCT p.id) AS teaching_staff_count
			FROM personnel p
			INNER JOIN schools s ON s.id = p.school_id
			WHERE p.role = ?
			  AND p.status = ?
			  AND s.zone_id IS NOT NULL
			GROUP BY s.zone_id
		),

		zone_students AS (
			SELECT
				s.zone_id,
				COUNT(DISTINCT e.student_id) AS student_count
			FROM enrollments e
			INNER JOIN schools s ON s.id = e.school_id
			INNER JOIN students st ON st.id = e.student_id
			WHERE e.status = ?
			  AND st.status = ?
			  AND s.zone_id IS NOT NULL
			GROUP BY s.zone_id
		)

		SELECT
			z.name AS zone,
			COALESCE(zs.school_count, 0) AS school,
			COALESCE(zt.teaching_staff_count, 0) AS teaching_staff,
			COALESCE(zstu.student_count, 0) AS students,
			CASE
				WHEN COALESCE(zt.teaching_staff_count, 0) = 0 THEN NULL
				ELSE ROUND(
					(1.0 * COALESCE(zstu.student_count, 0)) /
					NULLIF(COALESCE(zt.teaching_staff_count, 0), 0),
					2
				)
			END AS students_teachers_ratio
		FROM zones z
		LEFT JOIN zone_schools zs ON zs.zone_id = z.id
		LEFT JOIN zone_teachers zt ON zt.zone_id = z.id
		LEFT JOIN zone_students zstu ON zstu.zone_id = z.id
		WHERE 1=1 %s
		ORDER BY z.name;
	`, stateFilter)

	err := r.db.Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("query zone summary report: %w", err)
	}

	return results, nil
}
