package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL ATTENDANCE REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type personnelAttendanceRepo struct{ db *sql.DB }

func NewPersonnelAttendanceRepository(db *sql.DB) domain.PersonnelAttendanceRepository {
	return &personnelAttendanceRepo{db: db}
}

func (r *personnelAttendanceRepo) Create(ctx context.Context, a *domain.PersonnelAttendance) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO personnel_attendance (id,personnel_id,school_id,attendance_date,status,remarks,recorded_by,created_at,updated_at,created_by,updated_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 ON CONFLICT (personnel_id, attendance_date) DO UPDATE SET
			status=$5, remarks=$6, recorded_by=$7, updated_at=$9, updated_by=$11`,
		a.ID, a.PersonnelID, a.SchoolID, a.AttendanceDate, a.Status, a.Remarks, a.RecordedBy,
		a.CreatedAt, a.UpdatedAt, a.CreatedBy, a.UpdatedBy)
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *personnelAttendanceRepo) GetByID(ctx context.Context, id string) (*domain.PersonnelAttendance, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,personnel_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM personnel_attendance WHERE id=$1`, id)
	return scanPersonnelAttendance(row)
}

func (r *personnelAttendanceRepo) GetByPersonnelAndDate(ctx context.Context, personnelID string, date time.Time) (*domain.PersonnelAttendance, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,personnel_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM personnel_attendance WHERE personnel_id=$1 AND attendance_date=$2`,
		personnelID, date)
	return scanPersonnelAttendance(row)
}

func (r *personnelAttendanceRepo) Update(ctx context.Context, a *domain.PersonnelAttendance) error {
	a.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE personnel_attendance SET status=$1, remarks=$2, recorded_by=$3, updated_at=$4, updated_by=$5 WHERE id=$6`,
		a.Status, a.Remarks, a.RecordedBy, a.UpdatedAt, a.UpdatedBy, a.ID)
	return checkRowsAffected(res, err, "personnel_attendance", a.ID)
}

func (r *personnelAttendanceRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM personnel_attendance WHERE id=$1`, id)
	return checkRowsAffected(res, err, "personnel_attendance", id)
}

func (r *personnelAttendanceRepo) ListBySchoolAndDate(ctx context.Context, schoolID string, date time.Time) ([]*domain.PersonnelAttendance, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,personnel_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM personnel_attendance WHERE school_id=$1 AND attendance_date=$2 ORDER BY created_at`,
		schoolID, date)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectPersonnelAttendance(rows)
}

func (r *personnelAttendanceRepo) ListByPersonnelAndRange(ctx context.Context, personnelID string, from, to time.Time) ([]*domain.PersonnelAttendance, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,personnel_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM personnel_attendance WHERE personnel_id=$1 AND attendance_date >= $2 AND attendance_date <= $3 ORDER BY attendance_date`,
		personnelID, from, to)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectPersonnelAttendance(rows)
}

func collectPersonnelAttendance(rows *sql.Rows) ([]*domain.PersonnelAttendance, error) {
	var out []*domain.PersonnelAttendance
	for rows.Next() {
		a, err := scanPersonnelAttendance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanPersonnelAttendance(s scanner) (*domain.PersonnelAttendance, error) {
	a := &domain.PersonnelAttendance{}
	err := s.Scan(
		&a.ID, &a.PersonnelID, &a.SchoolID, &a.AttendanceDate, &a.Status, &a.Remarks, &a.RecordedBy,
		&a.CreatedAt, &a.UpdatedAt, &a.CreatedBy, &a.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("personnel_attendance", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return a, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT ATTENDANCE REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type studentAttendanceRepo struct{ db *sql.DB }

func NewStudentAttendanceRepository(db *sql.DB) domain.StudentAttendanceRepository {
	return &studentAttendanceRepo{db: db}
}

func (r *studentAttendanceRepo) Create(ctx context.Context, a *domain.StudentAttendance) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO student_attendance (id,student_id,school_id,attendance_date,status,remarks,recorded_by,created_at,updated_at,created_by,updated_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 ON CONFLICT (student_id, attendance_date) DO UPDATE SET
			status=$5, remarks=$6, recorded_by=$7, updated_at=$9, updated_by=$11`,
		a.ID, a.StudentID, a.SchoolID, a.AttendanceDate, a.Status, a.Remarks, a.RecordedBy,
		a.CreatedAt, a.UpdatedAt, a.CreatedBy, a.UpdatedBy)
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *studentAttendanceRepo) GetByID(ctx context.Context, id string) (*domain.StudentAttendance, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,student_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM student_attendance WHERE id=$1`, id)
	return scanStudentAttendance(row)
}

func (r *studentAttendanceRepo) GetByStudentAndDate(ctx context.Context, studentID string, date time.Time) (*domain.StudentAttendance, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,student_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM student_attendance WHERE student_id=$1 AND attendance_date=$2`,
		studentID, date)
	return scanStudentAttendance(row)
}

func (r *studentAttendanceRepo) Update(ctx context.Context, a *domain.StudentAttendance) error {
	a.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE student_attendance SET status=$1, remarks=$2, recorded_by=$3, updated_at=$4, updated_by=$5 WHERE id=$6`,
		a.Status, a.Remarks, a.RecordedBy, a.UpdatedAt, a.UpdatedBy, a.ID)
	return checkRowsAffected(res, err, "student_attendance", a.ID)
}

func (r *studentAttendanceRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM student_attendance WHERE id=$1`, id)
	return checkRowsAffected(res, err, "student_attendance", id)
}

func (r *studentAttendanceRepo) ListBySchoolAndDate(ctx context.Context, schoolID string, date time.Time) ([]*domain.StudentAttendance, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,student_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM student_attendance WHERE school_id=$1 AND attendance_date=$2 ORDER BY created_at`,
		schoolID, date)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectStudentAttendance(rows)
}

func (r *studentAttendanceRepo) ListByStudentAndRange(ctx context.Context, studentID string, from, to time.Time) ([]*domain.StudentAttendance, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,student_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM student_attendance WHERE student_id=$1 AND attendance_date >= $2 AND attendance_date <= $3 ORDER BY attendance_date`,
		studentID, from, to)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectStudentAttendance(rows)
}

func (r *studentAttendanceRepo) ListBySchoolAndRange(ctx context.Context, schoolID string, from, to time.Time) ([]*domain.StudentAttendance, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,student_id,school_id,attendance_date,status,COALESCE(remarks,''),recorded_by,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM student_attendance WHERE school_id=$1 AND attendance_date >= $2 AND attendance_date <= $3 ORDER BY attendance_date, student_id`,
		schoolID, from, to)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectStudentAttendance(rows)
}

func collectStudentAttendance(rows *sql.Rows) ([]*domain.StudentAttendance, error) {
	var out []*domain.StudentAttendance
	for rows.Next() {
		a, err := scanStudentAttendance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanStudentAttendance(s scanner) (*domain.StudentAttendance, error) {
	a := &domain.StudentAttendance{}
	err := s.Scan(
		&a.ID, &a.StudentID, &a.SchoolID, &a.AttendanceDate, &a.Status, &a.Remarks, &a.RecordedBy,
		&a.CreatedAt, &a.UpdatedAt, &a.CreatedBy, &a.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("student_attendance", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return a, nil
}
