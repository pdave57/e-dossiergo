package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC SESSION REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type sessionRepo struct{ db *sql.DB }

func NewAcademicSessionRepository(db *sql.DB) domain.AcademicSessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s *domain.AcademicSession) error {
	s.ID = uuid.NewString()
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO academic_sessions
		 (id,school_id,name,start_year,end_year,status,start_date,end_date,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.SchoolID, s.Name, s.StartYear, s.EndYear, s.Status,
		s.StartDate, s.EndDate, s.CreatedAt, s.UpdatedAt, s.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("academic session name already exists for this school")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *sessionRepo) GetByID(ctx context.Context, id string) (*domain.AcademicSession, error) {
	return scanSession(r.db.QueryRowContext(ctx,
		sessionSelect+" WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *sessionRepo) GetActive(ctx context.Context, schoolID string) (*domain.AcademicSession, error) {
	return scanSession(r.db.QueryRowContext(ctx,
		sessionSelect+" WHERE school_id=$1 AND status='ACTIVE' AND deleted_at IS NULL", schoolID))
}

func (r *sessionRepo) Update(ctx context.Context, s *domain.AcademicSession) error {
	s.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE academic_sessions SET name=$1,start_year=$2,end_year=$3,status=$4,
		  start_date=$5,end_date=$6,updated_at=$7,updated_by=$8
		 WHERE id=$9 AND deleted_at IS NULL`,
		s.Name, s.StartYear, s.EndYear, s.Status,
		s.StartDate, s.EndDate, s.UpdatedAt, s.UpdatedBy, s.ID)
	return checkRowsAffected(res, err, "academic_session", s.ID)
}

func (r *sessionRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE academic_sessions SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "academic_session", id)
}

func (r *sessionRepo) List(ctx context.Context, schoolID string, p pagination.Params) ([]*domain.AcademicSession, int, error) {
	where := ` WHERE deleted_at IS NULL`
	args := []any{}
	if schoolID != "" {
		where = ` WHERE school_id=$1 AND deleted_at IS NULL`
		args = []any{schoolID}
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM academic_sessions`+where, args...).
		Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf("%s%s ORDER BY start_year DESC LIMIT $%d OFFSET $%d", sessionSelect, where, len(args)+1, len(args)+2),
		append(append([]any{}, args...), p.PerPage, p.Offset)...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.AcademicSession
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

// SetActive atomically makes one session ACTIVE and closes others.
func (r *sessionRepo) SetActive(ctx context.Context, id, schoolID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.Internal(err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err = tx.ExecContext(ctx,
		`UPDATE academic_sessions SET status='CLOSED',updated_at=NOW()
		 WHERE school_id=$1 AND status='ACTIVE' AND deleted_at IS NULL`, schoolID); err != nil {
		return apperror.Internal(err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE academic_sessions SET status='ACTIVE',updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if err = checkRowsAffected(res, err, "academic_session", id); err != nil {
		return err
	}
	return tx.Commit()
}

const sessionSelect = `
	SELECT id,school_id,name,start_year,end_year,status,start_date,end_date,
	       created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM academic_sessions`

func scanSession(s scanner) (*domain.AcademicSession, error) {
	as := &domain.AcademicSession{}
	err := s.Scan(&as.ID, &as.SchoolID, &as.Name, &as.StartYear, &as.EndYear,
		&as.Status, &as.StartDate, &as.EndDate,
		&as.CreatedAt, &as.UpdatedAt, &as.CreatedBy, &as.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("academic_session", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return as, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// TERM REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type termRepo struct{ db *sql.DB }

func NewTermRepository(db *sql.DB) domain.TermRepository { return &termRepo{db: db} }

func (r *termRepo) Create(ctx context.Context, t *domain.Term) error {
	t.ID = uuid.NewString()
	now := time.Now()
	t.CreatedAt, t.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO terms (id,session_id,term_number,name,start_date,end_date,is_active,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.SessionID, t.Number, t.Name, t.StartDate, t.EndDate,
		t.IsActive, t.CreatedAt, t.UpdatedAt, t.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict(fmt.Sprintf("term %d already exists for this session", t.Number))
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *termRepo) GetByID(ctx context.Context, id string) (*domain.Term, error) {
	return scanTerm(r.db.QueryRowContext(ctx, termSelect+" WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *termRepo) GetActiveTerm(ctx context.Context, sessionID string) (*domain.Term, error) {
	return scanTerm(r.db.QueryRowContext(ctx,
		termSelect+" WHERE session_id=$1 AND is_active=TRUE AND deleted_at IS NULL", sessionID))
}

func (r *termRepo) Update(ctx context.Context, t *domain.Term) error {
	t.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE terms SET name=$1,start_date=$2,end_date=$3,is_active=$4,updated_at=$5,updated_by=$6
		 WHERE id=$7 AND deleted_at IS NULL`,
		t.Name, t.StartDate, t.EndDate, t.IsActive, t.UpdatedAt, t.UpdatedBy, t.ID)
	return checkRowsAffected(res, err, "term", t.ID)
}

func (r *termRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE terms SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "term", id)
}

func (r *termRepo) ListAll(ctx context.Context) ([]*domain.Term, error) {
	rows, err := r.db.QueryContext(ctx,
		termSelect+" WHERE deleted_at IS NULL ORDER BY session_id, term_number")
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Term
	for rows.Next() {
		t, err := scanTerm(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *termRepo) ListBySession(ctx context.Context, sessionID string) ([]*domain.Term, error) {
	rows, err := r.db.QueryContext(ctx,
		termSelect+" WHERE session_id=$1 AND deleted_at IS NULL ORDER BY term_number", sessionID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Term
	for rows.Next() {
		t, err := scanTerm(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *termRepo) SetActive(ctx context.Context, id, sessionID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.Internal(err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err = tx.ExecContext(ctx,
		`UPDATE terms SET is_active=FALSE WHERE session_id=$1`, sessionID); err != nil {
		return apperror.Internal(err)
	}
	res, err := tx.ExecContext(ctx, `UPDATE terms SET is_active=TRUE WHERE id=$1`, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if err = checkRowsAffected(res, nil, "term", id); err != nil {
		return err
	}
	return tx.Commit()
}

const termSelect = `
	SELECT id,session_id,term_number,name,start_date,end_date,is_active,
	       created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM terms`

func scanTerm(s scanner) (*domain.Term, error) {
	t := &domain.Term{}
	err := s.Scan(&t.ID, &t.SessionID, &t.Number, &t.Name,
		&t.StartDate, &t.EndDate, &t.IsActive,
		&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy, &t.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("term", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return t, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type levelRepo struct{ db *sql.DB }

func NewLevelRepository(db *sql.DB) domain.LevelRepository { return &levelRepo{db: db} }

func (r *levelRepo) Create(ctx context.Context, l *domain.Level) error {
	if l.ID == "" {
		l.ID = uuid.NewString()
	}
	now := time.Now()
	l.CreatedAt, l.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO levels (id,state_id,name,code,type,ord,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		l.ID, l.SchoolID, l.Name, l.Code, l.Type, l.Order,
		l.CreatedAt, l.UpdatedAt, l.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("level code already exists for this school")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *levelRepo) GetByID(ctx context.Context, id string) (*domain.Level, error) {
	return scanLevel(r.db.QueryRowContext(ctx,
		levelSelect+" WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *levelRepo) Update(ctx context.Context, l *domain.Level) error {
	l.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE levels SET name=$1,code=$2,type=$3,ord=$4,updated_at=$5,updated_by=$6
		 WHERE id=$7 AND deleted_at IS NULL`,
		l.Name, l.Code, l.Type, l.Order, l.UpdatedAt, l.UpdatedBy, l.ID)
	return checkRowsAffected(res, err, "level", l.ID)
}

func (r *levelRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE levels SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "level", id)
}

func (r *levelRepo) ListBySchool(ctx context.Context, schoolID string) ([]*domain.Level, error) {
	rows, err := r.db.QueryContext(ctx,
		levelSelect+" WHERE state_id=$1 AND deleted_at IS NULL ORDER BY ord,name", schoolID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Level
	for rows.Next() {
		l, err := scanLevel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// GetNextLevel returns the level with ord = current+1 of the same type.
func (r *levelRepo) GetNextLevel(ctx context.Context, currentLevelID string) (*domain.Level, error) {
	return scanLevel(r.db.QueryRowContext(ctx, `
		SELECT id,state_id,name,code,type,ord,created_at,updated_at,
		       COALESCE(created_by,''),COALESCE(updated_by,'')
		FROM levels
		WHERE state_id=(SELECT state_id FROM levels WHERE id=$1)
		  AND type=(SELECT type FROM levels WHERE id=$1)
		  AND ord=(SELECT ord+1 FROM levels WHERE id=$1)
		  AND deleted_at IS NULL`, currentLevelID))
}

const levelSelect = `
	SELECT id,state_id,name,code,type,ord,created_at,updated_at,
	       COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM levels`

func scanLevel(s scanner) (*domain.Level, error) {
	l := &domain.Level{}
	err := s.Scan(&l.ID, &l.SchoolID, &l.Name, &l.Code, &l.Type, &l.Order,
		&l.CreatedAt, &l.UpdatedAt, &l.CreatedBy, &l.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("level", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return l, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SUB-LEVEL REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type subLevelRepo struct{ db *sql.DB }

func NewSubLevelRepository(db *sql.DB) domain.SubLevelRepository { return &subLevelRepo{db: db} }

func (r *subLevelRepo) Create(ctx context.Context, s *domain.SubLevel) error {
	s.ID = uuid.NewString()
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sub_levels (id,school_id,level_id,name,code,capacity,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		s.ID, s.SchoolID, s.LevelID, s.Name, s.Code, s.Capacity,
		s.CreatedAt, s.UpdatedAt, s.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("sub-level code already exists for this school/level")
		}
		if isFKViolation(err) {
			return apperror.BadRequest("referenced school or level does not exist")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *subLevelRepo) GetByID(ctx context.Context, id string) (*domain.SubLevel, error) {
	return scanSubLevel(r.db.QueryRowContext(ctx,
		subLevelSelect+" WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *subLevelRepo) Update(ctx context.Context, s *domain.SubLevel) error {
	s.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE sub_levels SET name=$1,code=$2,capacity=$3,updated_at=$4,updated_by=$5
		 WHERE id=$6 AND deleted_at IS NULL`,
		s.Name, s.Code, s.Capacity, s.UpdatedAt, s.UpdatedBy, s.ID)
	return checkRowsAffected(res, err, "sub_level", s.ID)
}

func (r *subLevelRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE sub_levels SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "sub_level", id)
}

func (r *subLevelRepo) ListByLevel(ctx context.Context, schoolID, levelID string) ([]*domain.SubLevel, error) {
	query := subLevelSelect + " WHERE deleted_at IS NULL"
	args := []any{}
	if schoolID != "" {
		args = append(args, schoolID)
		query += fmt.Sprintf(" AND school_id=$%d", len(args))
	}
	if levelID != "" {
		args = append(args, levelID)
		query += fmt.Sprintf(" AND level_id=$%d", len(args))
	}
	query += " ORDER BY code"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.SubLevel
	for rows.Next() {
		s, err := scanSubLevel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *subLevelRepo) CountEnrolled(ctx context.Context, subLevelID, sessionID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM enrollments
		 WHERE sub_level_id=$1 AND session_id=$2 AND status='ACTIVE'`,
		subLevelID, sessionID).Scan(&n)
	return n, err
}

const subLevelSelect = `
	SELECT id,school_id,level_id,name,code,capacity,created_at,updated_at,
	       COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM sub_levels`

func scanSubLevel(s scanner) (*domain.SubLevel, error) {
	sl := &domain.SubLevel{}
	err := s.Scan(&sl.ID, &sl.SchoolID, &sl.LevelID, &sl.Name, &sl.Code, &sl.Capacity,
		&sl.CreatedAt, &sl.UpdatedAt, &sl.CreatedBy, &sl.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("sub_level", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return sl, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL LEVEL REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type schoolLevelRepo struct{ db *sql.DB }

func NewSchoolLevelRepository(db *sql.DB) domain.SchoolLevelRepository {
	return &schoolLevelRepo{db: db}
}

func (r *schoolLevelRepo) Upsert(ctx context.Context, sl *domain.SchoolLevel) error {
	if sl.ID == "" {
		sl.ID = uuid.NewString()
	}
	now := time.Now()
	sl.CreatedAt, sl.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO school_levels (id,school_id,level_id,session_id,is_active,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 ON CONFLICT (school_id,level_id,session_id)
		 DO UPDATE SET is_active=EXCLUDED.is_active,updated_at=NOW()`,
		sl.ID, sl.SchoolID, sl.LevelID, sl.SessionID, sl.IsActive,
		sl.CreatedAt, sl.UpdatedAt, sl.CreatedBy)
	return err
}

func (r *schoolLevelRepo) ListBySchool(ctx context.Context, schoolID, sessionID string) ([]*domain.SchoolLevel, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,school_id,level_id,session_id,is_active,created_at,updated_at,
		        COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM school_levels WHERE school_id=$1 AND session_id=$2`,
		schoolID, sessionID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.SchoolLevel
	for rows.Next() {
		sl := &domain.SchoolLevel{}
		if err := rows.Scan(&sl.ID, &sl.SchoolID, &sl.LevelID, &sl.SessionID, &sl.IsActive,
			&sl.CreatedAt, &sl.UpdatedAt, &sl.CreatedBy, &sl.UpdatedBy); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, sl)
	}
	return out, rows.Err()
}

func (r *schoolLevelRepo) Delete(ctx context.Context, schoolID, levelID, sessionID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM school_levels WHERE school_id=$1 AND level_id=$2 AND session_id=$3`,
		schoolID, levelID, sessionID)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// SUBJECT REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type subjectRepo struct{ db *sql.DB }

func NewSubjectRepository(db *sql.DB) domain.SubjectRepository { return &subjectRepo{db: db} }

func (r *subjectRepo) Create(ctx context.Context, s *domain.Subject) error {
	s.ID = uuid.NewString()
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subjects (id,state_id,name,code,category,level_type,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		s.ID, s.StateID, s.Name, s.Code, s.Category, s.LevelType,
		s.CreatedAt, s.UpdatedAt, s.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("subject code already exists for this state")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *subjectRepo) GetByID(ctx context.Context, id string) (*domain.Subject, error) {
	return scanSubject(r.db.QueryRowContext(ctx,
		subjectSelect+" WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *subjectRepo) GetByCode(ctx context.Context, code string) (*domain.Subject, error) {
	return scanSubject(r.db.QueryRowContext(ctx,
		subjectSelect+" WHERE code=$1 AND deleted_at IS NULL", code))
}

func (r *subjectRepo) Update(ctx context.Context, s *domain.Subject) error {
	s.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE subjects SET name=$1,code=$2,category=$3,level_type=$4,updated_at=$5,updated_by=$6
		 WHERE id=$7 AND deleted_at IS NULL`,
		s.Name, s.Code, s.Category, s.LevelType, s.UpdatedAt, s.UpdatedBy, s.ID)
	return checkRowsAffected(res, err, "subject", s.ID)
}

func (r *subjectRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE subjects SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "subject", id)
}

func (r *subjectRepo) ListByState(ctx context.Context, stateID string) ([]*domain.Subject, error) {
	rows, err := r.db.QueryContext(ctx,
		subjectSelect+" WHERE state_id=$1 AND deleted_at IS NULL ORDER BY level_type,name", stateID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Subject
	for rows.Next() {
		s, err := scanSubject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

const subjectSelect = `
	SELECT id,state_id,name,code,category,level_type,created_at,updated_at,
	       COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM subjects`

func scanSubject(s scanner) (*domain.Subject, error) {
	sub := &domain.Subject{}
	err := s.Scan(&sub.ID, &sub.StateID, &sub.Name, &sub.Code, &sub.Category, &sub.LevelType,
		&sub.CreatedAt, &sub.UpdatedAt, &sub.CreatedBy, &sub.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("subject", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return sub, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL SUBJECT REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type schoolSubjectRepo struct{ db *sql.DB }

func NewSchoolSubjectRepository(db *sql.DB) domain.SchoolSubjectRepository {
	return &schoolSubjectRepo{db: db}
}

func (r *schoolSubjectRepo) Create(ctx context.Context, ss *domain.SchoolSubject) error {
	ss.ID = uuid.NewString()
	now := time.Now()
	ss.CreatedAt, ss.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO school_subjects
		 (id,school_id,subject_id,level_id,session_id,teacher_id,is_active,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,$10)`,
		ss.ID, ss.SchoolID, ss.SubjectID, ss.LevelID, ss.SessionID,
		ss.TeacherID, ss.IsActive, ss.CreatedAt, ss.UpdatedAt, ss.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("subject already assigned to this school/level/session")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *schoolSubjectRepo) GetByID(ctx context.Context, id string) (*domain.SchoolSubject, error) {
	ss := &domain.SchoolSubject{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,school_id,subject_id,level_id,session_id,COALESCE(teacher_id,''),
		        is_active,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM school_subjects WHERE id=$1`, id).
		Scan(&ss.ID, &ss.SchoolID, &ss.SubjectID, &ss.LevelID, &ss.SessionID,
			&ss.TeacherID, &ss.IsActive, &ss.CreatedAt, &ss.UpdatedAt, &ss.CreatedBy, &ss.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("school_subject", id)
	}
	return ss, err
}

func (r *schoolSubjectRepo) Update(ctx context.Context, ss *domain.SchoolSubject) error {
	ss.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE school_subjects SET teacher_id=NULLIF($1,''),is_active=$2,updated_at=$3,updated_by=$4
		 WHERE id=$5`,
		ss.TeacherID, ss.IsActive, ss.UpdatedAt, ss.UpdatedBy, ss.ID)
	return checkRowsAffected(res, err, "school_subject", ss.ID)
}

func (r *schoolSubjectRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM school_subjects WHERE id=$1`, id)
	return checkRowsAffected(res, err, "school_subject", id)
}

func (r *schoolSubjectRepo) ListBySchool(ctx context.Context, schoolID, sessionID, levelID string) ([]*domain.SchoolSubject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,school_id,subject_id,level_id,session_id,COALESCE(teacher_id,''),
		        is_active,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM school_subjects
		 WHERE school_id=$1 AND session_id=$2 AND level_id=$3`,
		schoolID, sessionID, levelID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.SchoolSubject
	for rows.Next() {
		ss := &domain.SchoolSubject{}
		if err := rows.Scan(&ss.ID, &ss.SchoolID, &ss.SubjectID, &ss.LevelID, &ss.SessionID,
			&ss.TeacherID, &ss.IsActive, &ss.CreatedAt, &ss.UpdatedAt, &ss.CreatedBy, &ss.UpdatedBy); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, ss)
	}
	return out, rows.Err()
}
