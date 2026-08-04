package repository

import (
	"context"
	"database/sql"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
)

type recommendationRepo struct{ db *sql.DB }

func NewRecommendationRepository(db *sql.DB) domain.RecommendationRepository {
	return &recommendationRepo{db: db}
}

func (r *recommendationRepo) ListSchoolsWithAggregates(ctx context.Context) ([]domain.SchoolRecommendationRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			s.id,
			s.name,
			COALESCE(z.name, '') AS zone_name,
			COALESCE(l.name, '') AS lga_name,
			s.category,
			COALESCE(teacher_counts.total_teachers, 0) AS total_teachers,
			COALESCE(teacher_counts.qualified_teachers, 0) AS qualified_teachers,
			COALESCE(s.total_students, 0) AS total_students,
			COALESCE(s.number_of_classrooms, 0) AS total_classrooms,
			COALESCE(s.number_of_classrooms, 0) AS functional_classrooms,
			COALESCE(facility_flags.has_library, false) AS has_library,
			COALESCE(facility_flags.has_laboratory, false) AS has_laboratory,
			COALESCE(facility_flags.has_toilet, false) AS has_toilet,
			COALESCE(facility_flags.has_electricity, false) AS has_electricity,
			COALESCE(facility_flags.has_water, false) AS has_water,
			COALESCE(facility_flags.has_internet, false) AS has_internet,
			COALESCE(subject_counts.subjects_offered, 0) AS subjects_offered,
			COALESCE(expected.subject_count, 0) AS expected_subjects,
			0 AS books_per_student,
			perf.avg_pass_rate
		FROM schools s
		LEFT JOIN zones z ON z.id = s.zone_id
		LEFT JOIN lgas l ON l.id = s.lga_id
		LEFT JOIN (
			SELECT
				school_id,
				COUNT(CASE WHEN role = 'TEACHER' THEN 1 END) AS total_teachers,
				COUNT(CASE WHEN role = 'TEACHER' AND qualification IN ('NCE','B.Ed','M.Ed','Ph.D') THEN 1 END) AS qualified_teachers
			FROM personnel
			WHERE deleted_at IS NULL
			GROUP BY school_id
		) teacher_counts ON teacher_counts.school_id = s.id
		LEFT JOIN (
			SELECT
				school_id,
				BOOL_OR(type = 'LIBRARY') AS has_library,
				BOOL_OR(type = 'LABORATORY') AS has_laboratory,
				BOOL_OR(type = 'TOILET') AS has_toilet,
				BOOL_OR(type = 'GENERATOR') AS has_electricity,
				BOOL_OR(type = 'BOREHOLE') AS has_water,
				BOOL_OR(type = 'ICT_CENTER') AS has_internet
			FROM school_facilities
			WHERE deleted_at IS NULL
			GROUP BY school_id
		) facility_flags ON facility_flags.school_id = s.id
		LEFT JOIN (
			SELECT school_id, COUNT(*) AS subjects_offered
			FROM school_subjects
			WHERE is_active = TRUE
			GROUP BY school_id
		) subject_counts ON subject_counts.school_id = s.id
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS subject_count
			FROM subjects sub
			WHERE sub.deleted_at IS NULL
			  AND (
				  (s.category = 'PRIMARY' AND sub.level_type = 'PRIMARY')
				  OR (s.category = 'JUNIOR_SECONDARY' AND sub.level_type = 'JSS')
				  OR (s.category = 'SENIOR_SECONDARY' AND sub.level_type = 'SSS')
				  OR (s.category = 'VOCATIONAL' AND sub.level_type = 'VOCATIONAL')
				  OR (s.category = 'COMBINED' AND sub.level_type IN ('JSS','SSS'))
			  )
		) expected ON true
		LEFT JOIN (
			SELECT school_id, AVG(total_score) AS avg_pass_rate
			FROM score_sheets
			WHERE total_score > 0
			GROUP BY school_id
		) perf ON perf.school_id = s.id
		WHERE s.deleted_at IS NULL
		ORDER BY s.name
	`)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()

	var out []domain.SchoolRecommendationRow
	for rows.Next() {
		var r domain.SchoolRecommendationRow
		var avgPassRate sql.NullFloat64
		err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.ZoneName,
			&r.LGAName,
			&r.Category,
			&r.TotalTeachers,
			&r.QualifiedTeachers,
			&r.TotalStudents,
			&r.TotalClassrooms,
			&r.FunctionalClassrooms,
			&r.HasLibrary,
			&r.HasLaboratory,
			&r.HasToilet,
			&r.HasElectricity,
			&r.HasWater,
			&r.HasInternet,
			&r.SubjectsOffered,
			&r.ExpectedSubjects,
			&r.BooksPerStudent,
			&avgPassRate,
		)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if avgPassRate.Valid {
			val := avgPassRate.Float64
			r.AvgPassRate = &val
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
