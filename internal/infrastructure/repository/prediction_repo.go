// Package repository — PostgreSQL implementation of domain.PredictionRepository.
// All queries are read-only aggregates. No writes.
package repository

import (
	"database/sql"
	"fmt"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
)

type predictionRepo struct{ db *sql.DB }

// NewPredictionRepository returns a domain.PredictionRepository backed by PostgreSQL.
func NewPredictionRepository(db *sql.DB) domain.PredictionRepository {
	return &predictionRepo{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// FACILITY SIGNAL
// ─────────────────────────────────────────────────────────────────────────────

func (r *predictionRepo) GetFacilitySignal(schoolID string) (*domain.FacilitySignal, error) {
	const q = `
		SELECT
			COUNT(*)                                                             AS total,
			COUNT(*) FILTER (WHERE condition = 'GOOD')                           AS good_count,
			COUNT(*) FILTER (WHERE condition = 'FAIR')                           AS fair_count,
			COUNT(*) FILTER (WHERE condition = 'POOR')                           AS poor_count,
			COUNT(*) FILTER (WHERE condition = 'DEFUNCT')                        AS defunct_count,
			COALESCE(BOOL_OR(type = 'LIBRARY'), false)                           AS has_library,
			COALESCE(BOOL_OR(type = 'LABORATORY'), false)                        AS has_lab,
			COALESCE(BOOL_OR(type = 'ICT_CENTER'), false)                        AS has_ict,
			COALESCE(BOOL_OR(type = 'SPORT_FIELD'), false)                       AS has_sport
		FROM school_facilities
		WHERE school_id = $1
		  AND deleted_at IS NULL`

	sig := &domain.FacilitySignal{SchoolID: schoolID}
	err := r.db.QueryRow(q, schoolID).Scan(
		&sig.TotalFacilities,
		&sig.GoodCount,
		&sig.FairCount,
		&sig.PoorCount,
		&sig.DefunctCount,
		&sig.HasLibrary,
		&sig.HasLab,
		&sig.HasICT,
		&sig.HasSportField,
	)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("facility signal: %w", err))
	}
	return sig, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL SIGNAL
// ─────────────────────────────────────────────────────────────────────────────

func (r *predictionRepo) GetPersonnelSignal(schoolID string) (*domain.PersonnelSignal, error) {
	const q = `
		SELECT
			COUNT(*)                                                                        AS total_staff,
			COUNT(*) FILTER (WHERE status = 'ACTIVE')                                       AS active_staff,
			COUNT(*) FILTER (WHERE status = 'ACTIVE' AND role = 'TEACHER')                 AS teacher_count,
			COUNT(*) FILTER (WHERE status = 'ACTIVE'
				AND qualification ILIKE ANY(ARRAY['%B.Ed%','%B.Sc%','%NCE%','%M.Ed%','%Ph.D%','%HND%']))
				                                                                            AS qualified_count,
			COUNT(*) FILTER (WHERE status = 'ACTIVE'
				AND qualification ILIKE ANY(ARRAY['%M.Ed%','%Ph.D%','%M.Sc%']))            AS postgrad_count
		FROM personnel
		WHERE school_id = $1
		  AND deleted_at IS NULL`

	sig := &domain.PersonnelSignal{SchoolID: schoolID}
	err := r.db.QueryRow(q, schoolID).Scan(
		&sig.TotalStaff,
		&sig.ActiveStaff,
		&sig.TeacherCount,
		&sig.QualifiedCount,
		&sig.PostgradCount,
	)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("personnel signal: %w", err))
	}
	return sig, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// HISTORICAL SIGNAL — SCHOOL LEVEL
// ─────────────────────────────────────────────────────────────────────────────

func (r *predictionRepo) GetSchoolHistoricalSignal(schoolID string) (*domain.HistoricalSignal, error) {
	const q = `
		SELECT
			COALESCE(AVG(ss.total_score), 0)                                        AS avg_score,
			COALESCE(AVG(CASE WHEN ss.total_score >= 40 THEN 1.0 ELSE 0.0 END), 0) AS pass_rate,
			COALESCE(AVG(CASE WHEN ss.total_score >= 70 THEN 1.0 ELSE 0.0 END), 0) AS distinction_rate,
			COUNT(DISTINCT ss.term_id)                                              AS terms_recorded
		FROM score_sheets ss
		WHERE ss.school_id = $1`

	sig := &domain.HistoricalSignal{SchoolID: schoolID}
	err := r.db.QueryRow(q, schoolID).Scan(
		&sig.AvgScore,
		&sig.PassRate,
		&sig.DistinctionRate,
		&sig.TermsRecorded,
	)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("school historical signal: %w", err))
	}
	return sig, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// HISTORICAL SIGNAL — STUDENT LEVEL
// ─────────────────────────────────────────────────────────────────────────────

func (r *predictionRepo) GetStudentHistoricalSignal(studentID, schoolID string) (*domain.HistoricalSignal, error) {
	const q = `
		SELECT
			COALESCE(AVG(ss.total_score), 0)                                        AS avg_score,
			COALESCE(AVG(CASE WHEN ss.total_score >= 40 THEN 1.0 ELSE 0.0 END), 0) AS pass_rate,
			COALESCE(AVG(CASE WHEN ss.total_score >= 70 THEN 1.0 ELSE 0.0 END), 0) AS distinction_rate,
			COUNT(DISTINCT ss.term_id)                                              AS terms_recorded
		FROM score_sheets ss
		WHERE ss.student_id = $1
		  AND ss.school_id  = $2`

	sig := &domain.HistoricalSignal{StudentID: studentID, SchoolID: schoolID}
	err := r.db.QueryRow(q, studentID, schoolID).Scan(
		&sig.AvgScore,
		&sig.PassRate,
		&sig.DistinctionRate,
		&sig.TermsRecorded,
	)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("student historical signal: %w", err))
	}
	return sig, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ENROLLED STUDENTS
// ─────────────────────────────────────────────────────────────────────────────

func (r *predictionRepo) GetEnrolledStudents(schoolID, sessionID string) ([]domain.StudentRow, error) {
	q := `
		SELECT s.id, s.first_name, s.last_name
		FROM students s
		JOIN enrollments e ON e.student_id = s.id
		WHERE e.school_id = $1
		  AND e.status    = 'ACTIVE'
		  AND s.deleted_at IS NULL`

	args := []any{schoolID}
	if sessionID != "" {
		q += " AND e.session_id = $2"
		args = append(args, sessionID)
	}
	q += " ORDER BY s.last_name, s.first_name"

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("enrolled students: %w", err))
	}
	defer rows.Close()

	var out []domain.StudentRow
	for rows.Next() {
		var sr domain.StudentRow
		if err := rows.Scan(&sr.ID, &sr.FirstName, &sr.LastName); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, sr)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────────────────────────────────────

func (r *predictionRepo) GetSchoolName(schoolID string) (string, error) {
	var name string
	err := r.db.QueryRow(`SELECT name FROM schools WHERE id = $1`, schoolID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", apperror.NotFound("school", schoolID)
	}
	if err != nil {
		return "", apperror.Internal(err)
	}
	return name, nil
}

func (r *predictionRepo) GetEnrollmentCount(schoolID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM enrollments WHERE school_id = $1 AND status = 'ACTIVE'`,
		schoolID,
	).Scan(&count)
	return count, err
}