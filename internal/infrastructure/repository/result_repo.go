package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type personnelRepo struct{ db *sql.DB }

func NewPersonnelRepository(db *sql.DB) domain.PersonnelRepository { return &personnelRepo{db: db} }

func (r *personnelRepo) Create(ctx context.Context, p *domain.Personnel) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	now := time.Now()
	p.CreatedAt, p.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO personnel
		 (id,state_id,school_id,staff_id,first_name,middle_name,last_name,gender,
		  date_of_birth,email,phone,address,role,status,qualification,specialization,
		  date_of_employment,lga_id,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,NULLIF($10,''),NULLIF($11,''),
		         NULLIF($12,''),$13,$14,NULLIF($15,''),NULLIF($16,''),$17,NULLIF($18,''),$19,$20,$21)`,
		p.ID, p.StateID, p.SchoolID, p.StaffID, p.FirstName, p.MiddleName, p.LastName,
		p.Gender, p.DateOfBirth, p.Email, p.Phone, p.Address,
		p.Role, p.Status, p.Qualification, p.Specialization,
		p.DateOfEmployment, p.LGAID, p.CreatedAt, p.UpdatedAt, p.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("staff ID already registered")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *personnelRepo) GetByID(ctx context.Context, id string) (*domain.Personnel, error) {
	return scanPersonnel(r.db.QueryRowContext(ctx,
		personnelSelect+" WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *personnelRepo) GetByStaffID(ctx context.Context, staffID string) (*domain.Personnel, error) {
	return scanPersonnel(r.db.QueryRowContext(ctx,
		personnelSelect+" WHERE staff_id=$1 AND deleted_at IS NULL", staffID))
}

func (r *personnelRepo) Update(ctx context.Context, p *domain.Personnel) error {
	p.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE personnel SET school_id=$1,first_name=$2,middle_name=NULLIF($3,''),last_name=$4,
		  gender=$5,date_of_birth=$6,email=NULLIF($7,''),phone=NULLIF($8,''),
		  address=NULLIF($9,''),role=$10,status=$11,qualification=NULLIF($12,''),
		  specialization=NULLIF($13,''),date_of_employment=$14,updated_at=$15,updated_by=$16
		 WHERE id=$17 AND deleted_at IS NULL`,
		p.SchoolID, p.FirstName, p.MiddleName, p.LastName,
		p.Gender, p.DateOfBirth, p.Email, p.Phone,
		p.Address, p.Role, p.Status, p.Qualification,
		p.Specialization, p.DateOfEmployment, p.UpdatedAt, p.UpdatedBy, p.ID)
	return checkRowsAffected(res, err, "personnel", p.ID)
}

func (r *personnelRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE personnel SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "personnel", id)
}

func (r *personnelRepo) List(ctx context.Context, f domain.PersonnelFilter, p pagination.Params) ([]*domain.Personnel, int, error) {
	where, args := []string{"deleted_at IS NULL"}, []any{}
	idx := 1
	if f.StateID != "" {
		where = append(where, fmt.Sprintf("state_id=$%d", idx)); args = append(args, f.StateID); idx++
	}
	if f.SchoolID != "" {
		where = append(where, fmt.Sprintf("school_id=$%d", idx)); args = append(args, f.SchoolID); idx++
	}
	if f.Role != "" {
		where = append(where, fmt.Sprintf("role=$%d", idx)); args = append(args, f.Role); idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", idx)); args = append(args, f.Status); idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf(
			"(first_name ILIKE $%d OR last_name ILIKE $%d OR staff_id ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+f.Search+"%"); idx++
	}
	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM personnel "+clause, args...).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}
	args = append(args, p.PerPage, p.Offset)
	q := fmt.Sprintf(personnelSelect+" %s ORDER BY last_name,first_name LIMIT $%d OFFSET $%d",
		clause, idx, idx+1)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Personnel
	for rows.Next() {
		p, err := scanPersonnel(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

const personnelSelect = `
	SELECT id,state_id,school_id,staff_id,first_name,COALESCE(middle_name,''),last_name,gender,
	       date_of_birth,COALESCE(email,''),COALESCE(phone,''),COALESCE(address,''),
	       role,status,COALESCE(qualification,''),COALESCE(specialization,''),
	       date_of_employment,COALESCE(lga_id,''),created_at,updated_at,
	       COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM personnel`

func scanPersonnel(s scanner) (*domain.Personnel, error) {
	p := &domain.Personnel{}
	err := s.Scan(
		&p.ID, &p.StateID, &p.SchoolID, &p.StaffID,
		&p.FirstName, &p.MiddleName, &p.LastName, &p.Gender,
		&p.DateOfBirth, &p.Email, &p.Phone, &p.Address,
		&p.Role, &p.Status, &p.Qualification, &p.Specialization,
		&p.DateOfEmployment, &p.LGAID,
		&p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("personnel", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return p, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL TRANSFER REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type personnelTransferRepo struct{ db *sql.DB }

func NewPersonnelTransferRepository(db *sql.DB) domain.PersonnelTransferRepository {
	return &personnelTransferRepo{db: db}
}

func (r *personnelTransferRepo) Create(ctx context.Context, t *domain.PersonnelTransfer) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	now := time.Now()
	t.CreatedAt, t.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO personnel_transfers
		 (id,personnel_id,from_school_id,to_school_id,transfer_date,reason,approved_by,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,''),$8,$9,$10)`,
		t.ID, t.PersonnelID, t.FromSchoolID, t.ToSchoolID,
		t.TransferDate, t.Reason, t.ApprovedBy,
		t.CreatedAt, t.UpdatedAt, t.CreatedBy)
	return err
}

func (r *personnelTransferRepo) GetByID(ctx context.Context, id string) (*domain.PersonnelTransfer, error) {
	t := &domain.PersonnelTransfer{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,personnel_id,from_school_id,to_school_id,transfer_date,
		        COALESCE(reason,''),COALESCE(approved_by,''),created_at,updated_at
		 FROM personnel_transfers WHERE id=$1`, id).
		Scan(&t.ID, &t.PersonnelID, &t.FromSchoolID, &t.ToSchoolID,
			&t.TransferDate, &t.Reason, &t.ApprovedBy, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("transfer", id)
	}
	return t, err
}

func (r *personnelTransferRepo) ListByPersonnel(ctx context.Context, personnelID string) ([]*domain.PersonnelTransfer, error) {
	return r.listTransfers(ctx, "personnel_id=$1", personnelID)
}

func (r *personnelTransferRepo) ListBySchool(ctx context.Context, schoolID string) ([]*domain.PersonnelTransfer, error) {
	return r.listTransfers(ctx, "from_school_id=$1 OR to_school_id=$1", schoolID)
}

func (r *personnelTransferRepo) listTransfers(ctx context.Context, where, arg string) ([]*domain.PersonnelTransfer, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,personnel_id,from_school_id,to_school_id,transfer_date,
		        COALESCE(reason,''),COALESCE(approved_by,''),created_at,updated_at
		 FROM personnel_transfers WHERE `+where+` ORDER BY transfer_date DESC`, arg)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.PersonnelTransfer
	for rows.Next() {
		t := &domain.PersonnelTransfer{}
		if err := rows.Scan(&t.ID, &t.PersonnelID, &t.FromSchoolID, &t.ToSchoolID,
			&t.TransferDate, &t.Reason, &t.ApprovedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type studentRepo struct{ db *sql.DB }

func NewStudentRepository(db *sql.DB) domain.StudentRepository { return &studentRepo{db: db} }

func (r *studentRepo) Create(ctx context.Context, s *domain.Student) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO students
		 (id,state_id,admission_no,first_name,middle_name,last_name,gender,date_of_birth,
		  state_of_origin,lga_id,religion,phone,email,address,
		  guardian_name,guardian_phone,guardian_relation,status,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8,NULLIF($9,''),NULLIF($10,''),
		         NULLIF($11,''),NULLIF($12,''),NULLIF($13,''),NULLIF($14,''),
		         $15,$16,NULLIF($17,''),$18,$19,$20,$21)`,
		s.ID, s.StateID, s.AdmissionNo, s.FirstName, s.MiddleName, s.LastName,
		s.Gender, s.DateOfBirth, s.StateOfOrigin, s.LGAID,
		s.Religion, s.Phone, s.Email, s.Address,
		s.GuardianName, s.GuardianPhone, s.GuardianRelation,
		s.Status, s.CreatedAt, s.UpdatedAt, s.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("admission number already exists")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *studentRepo) GetByID(ctx context.Context, id string) (*domain.Student, error) {
	return scanStudent(r.db.QueryRowContext(ctx,
		studentSelect+" WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *studentRepo) GetByAdmissionNo(ctx context.Context, admNo string) (*domain.Student, error) {
	return scanStudent(r.db.QueryRowContext(ctx,
		studentSelect+" WHERE admission_no=$1 AND deleted_at IS NULL", admNo))
}

func (r *studentRepo) Update(ctx context.Context, s *domain.Student) error {
	s.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE students SET first_name=$1,middle_name=NULLIF($2,''),last_name=$3,
		  gender=$4,date_of_birth=$5,state_of_origin=NULLIF($6,''),lga_id=NULLIF($7,''),
		  religion=NULLIF($8,''),phone=NULLIF($9,''),email=NULLIF($10,''),address=NULLIF($11,''),
		  guardian_name=$12,guardian_phone=$13,guardian_relation=NULLIF($14,''),
		  status=$15,updated_at=$16,updated_by=$17
		 WHERE id=$18 AND deleted_at IS NULL`,
		s.FirstName, s.MiddleName, s.LastName,
		s.Gender, s.DateOfBirth, s.StateOfOrigin, s.LGAID,
		s.Religion, s.Phone, s.Email, s.Address,
		s.GuardianName, s.GuardianPhone, s.GuardianRelation,
		s.Status, s.UpdatedAt, s.UpdatedBy, s.ID)
	return checkRowsAffected(res, err, "student", s.ID)
}

func (r *studentRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE students SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "student", id)
}

func (r *studentRepo) List(ctx context.Context, f domain.StudentFilter, p pagination.Params) ([]*domain.Student, int, error) {
	where, args := []string{"s.deleted_at IS NULL"}, []any{}
	idx := 1
	if f.StateID != "" {
		where = append(where, fmt.Sprintf("s.state_id=$%d", idx)); args = append(args, f.StateID); idx++
	}
	if f.SchoolID != "" {
		// students are linked to schools via enrollment
		where = append(where, fmt.Sprintf(`EXISTS(
			SELECT 1 FROM enrollments e WHERE e.student_id=s.id AND e.school_id=$%d
		)`, idx))
		args = append(args, f.SchoolID); idx++
	}
	if f.LGAID != "" {
		where = append(where, fmt.Sprintf(`EXISTS(
			SELECT 1 FROM enrollments e
			JOIN schools sc ON sc.id=e.school_id
			WHERE e.student_id=s.id AND sc.lga_id=$%d
		)`, idx))
		args = append(args, f.LGAID); idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("s.status=$%d", idx)); args = append(args, f.Status); idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf(
			"(s.first_name ILIKE $%d OR s.last_name ILIKE $%d OR s.admission_no ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+f.Search+"%"); idx++
	}
	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM students s "+clause, args...).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}
	args = append(args, p.PerPage, p.Offset)
	q := fmt.Sprintf(`SELECT `+studentSelectAlias+` FROM students s %s ORDER BY s.last_name,s.first_name LIMIT $%d OFFSET $%d`,
		clause, idx, idx+1)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Student
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

// GetAllStudents returns every distinct student enrolled in schools that match
// the provided filters. Both lgaID and schoolID are optional — pass "" to skip
// that filter. When both are provided they are ANDed together (i.e. a specific
// school inside the given LGA). Results are ordered by last_name, first_name.
func (r *studentRepo) GetAllStudents(ctx context.Context, lgaID, schoolID string) ([]*domain.Student, error) {
	where := []string{"s.deleted_at IS NULL"}
	args := []any{}
	idx := 1

	// Both filters work via the enrollment → school join
	needJoin := lgaID != "" || schoolID != ""
	if lgaID != "" {
		where = append(where, fmt.Sprintf("sc.lga_id = $%d", idx))
		args = append(args, lgaID)
		idx++
	}
	if schoolID != "" {
		where = append(where, fmt.Sprintf("e.school_id = $%d", idx))
		args = append(args, schoolID)
		idx++
	}

	clause := strings.Join(where, " AND ")

	var q string
	if needJoin {
		q = `SELECT DISTINCT ` + studentSelectAlias + `
		FROM students s
		JOIN enrollments e ON e.student_id = s.id
		JOIN schools    sc ON sc.id        = e.school_id
		WHERE ` + clause + `
		ORDER BY s.last_name, s.first_name`
	} else {
		// No filters — return all non-deleted students
		q = `SELECT ` + studentSelectAlias + `
		FROM students s
		WHERE s.deleted_at IS NULL
		ORDER BY s.last_name, s.first_name`
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Student
	for rows.Next() {
		st, err := scanStudentAlias(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// studentSelect is used when querying the students table directly (no alias).
const studentSelect = `
	SELECT id,state_id,admission_no,first_name,COALESCE(middle_name,''),last_name,gender,date_of_birth,
	       COALESCE(state_of_origin,''),COALESCE(lga_id,''),COALESCE(religion,''),
	       COALESCE(phone,''),COALESCE(email,''),COALESCE(address,''),
	       guardian_name,guardian_phone,COALESCE(guardian_relation,''),
	       status,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM students`

// studentSelectAlias is used when students is aliased as "s" (e.g. in JOINs).
const studentSelectAlias = `
	s.id,s.state_id,s.admission_no,s.first_name,COALESCE(s.middle_name,''),s.last_name,s.gender,s.date_of_birth,
	COALESCE(s.state_of_origin,''),COALESCE(s.lga_id,''),COALESCE(s.religion,''),
	COALESCE(s.phone,''),COALESCE(s.email,''),COALESCE(s.address,''),
	s.guardian_name,s.guardian_phone,COALESCE(s.guardian_relation,''),
	s.status,s.created_at,s.updated_at,COALESCE(s.created_by,''),COALESCE(s.updated_by,'')`

func scanStudent(s scanner) (*domain.Student, error) {
	return scanStudentAlias(s)
}

// scanStudentAlias scans a student row regardless of whether the underlying
// query used an aliased or un-aliased column list.
func scanStudentAlias(s scanner) (*domain.Student, error) {
	st := &domain.Student{}
	err := s.Scan(
		&st.ID, &st.StateID, &st.AdmissionNo,
		&st.FirstName, &st.MiddleName, &st.LastName,
		&st.Gender, &st.DateOfBirth,
		&st.StateOfOrigin, &st.LGAID, &st.Religion,
		&st.Phone, &st.Email, &st.Address,
		&st.GuardianName, &st.GuardianPhone, &st.GuardianRelation,
		&st.Status, &st.CreatedAt, &st.UpdatedAt, &st.CreatedBy, &st.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("student", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return st, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ENROLLMENT REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type enrollmentRepo struct{ db *sql.DB }

func NewEnrollmentRepository(db *sql.DB) domain.EnrollmentRepository { return &enrollmentRepo{db: db} }

func (r *enrollmentRepo) Create(ctx context.Context, e *domain.Enrollment) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	now := time.Now()
	e.CreatedAt, e.UpdatedAt, e.EnrolledAt = now, now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO enrollments
		 (id,student_id,school_id,session_id,level_id,sub_level_id,status,enrolled_at,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.StudentID, e.SchoolID, e.SessionID, e.LevelID, e.SubLevelID,
		e.Status, e.EnrolledAt, e.CreatedAt, e.UpdatedAt, e.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("student already enrolled for this session")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *enrollmentRepo) GetByID(ctx context.Context, id string) (*domain.Enrollment, error) {
	return scanEnrollment(r.db.QueryRowContext(ctx,
		enrollmentSelect+" WHERE id=$1", id))
}

func (r *enrollmentRepo) GetActiveByStudent(ctx context.Context, studentID, sessionID string) (*domain.Enrollment, error) {
	return scanEnrollment(r.db.QueryRowContext(ctx,
		enrollmentSelect+" WHERE student_id=$1 AND session_id=$2 AND status='ACTIVE'",
		studentID, sessionID))
}

func (r *enrollmentRepo) ExistsForSession(ctx context.Context, studentID, sessionID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM enrollments WHERE student_id=$1 AND session_id=$2)`,
		studentID, sessionID).Scan(&exists)
	return exists, err
}

func (r *enrollmentRepo) Update(ctx context.Context, e *domain.Enrollment) error {
	e.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE enrollments SET sub_level_id=$1,status=$2,updated_at=$3,updated_by=$4 WHERE id=$5`,
		e.SubLevelID, e.Status, e.UpdatedAt, e.UpdatedBy, e.ID)
	return checkRowsAffected(res, err, "enrollment", e.ID)
}

func (r *enrollmentRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM enrollments WHERE id=$1`, id)
	return checkRowsAffected(res, err, "enrollment", id)
}

func (r *enrollmentRepo) List(ctx context.Context, f domain.EnrollmentFilter, p pagination.Params) ([]*domain.Enrollment, int, error) {
	where, args := []string{"1=1"}, []any{}
	idx := 1
	if f.SchoolID != "" {
		where = append(where, fmt.Sprintf("school_id=$%d", idx)); args = append(args, f.SchoolID); idx++
	}
	if f.SessionID != "" {
		where = append(where, fmt.Sprintf("session_id=$%d", idx)); args = append(args, f.SessionID); idx++
	}
	if f.LevelID != "" {
		where = append(where, fmt.Sprintf("level_id=$%d", idx)); args = append(args, f.LevelID); idx++
	}
	if f.SubLevelID != "" {
		where = append(where, fmt.Sprintf("sub_level_id=$%d", idx)); args = append(args, f.SubLevelID); idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", idx)); args = append(args, f.Status); idx++
	}
	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enrollments "+clause, args...).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}
	args = append(args, p.PerPage, p.Offset)
	q := fmt.Sprintf(enrollmentSelect+" %s ORDER BY enrolled_at DESC LIMIT $%d OFFSET $%d",
		clause, idx, idx+1)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Enrollment
	for rows.Next() {
		e, err := scanEnrollment(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

const enrollmentSelect = `
	SELECT id,student_id,school_id,session_id,level_id,sub_level_id,status,enrolled_at,
	       created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM enrollments`

func scanEnrollment(s scanner) (*domain.Enrollment, error) {
	e := &domain.Enrollment{}
	err := s.Scan(&e.ID, &e.StudentID, &e.SchoolID, &e.SessionID,
		&e.LevelID, &e.SubLevelID, &e.Status, &e.EnrolledAt,
		&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("enrollment", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return e, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL PROGRESSION REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type levelProgressionRepo struct{ db *sql.DB }

func NewLevelProgressionRepository(db *sql.DB) domain.LevelProgressionRepository {
	return &levelProgressionRepo{db: db}
}

func (r *levelProgressionRepo) Create(ctx context.Context, lp *domain.LevelProgression) error {
	if lp.ID == "" {
		lp.ID = uuid.NewString()
	}
	now := time.Now()
	lp.CreatedAt, lp.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO level_progressions
		 (id,student_id,school_id,from_session_id,to_session_id,from_level_id,to_level_id,
		  decision,decided_by,decision_date,remarks,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,NULLIF($7,''),$8,$9,$10,NULLIF($11,''),$12,$13,$14)`,
		lp.ID, lp.StudentID, lp.SchoolID, lp.FromSessionID, lp.ToSessionID,
		lp.FromLevelID, lp.ToLevelID, lp.Decision, lp.DecidedBy,
		lp.DecisionDate, lp.Remarks, lp.CreatedAt, lp.UpdatedAt, lp.CreatedBy)
	return err
}

func (r *levelProgressionRepo) GetByID(ctx context.Context, id string) (*domain.LevelProgression, error) {
	lp := &domain.LevelProgression{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id,student_id,school_id,from_session_id,COALESCE(to_session_id,''),
		       from_level_id,COALESCE(to_level_id,''),decision,decided_by,decision_date,
		       COALESCE(remarks,''),created_at,updated_at
		FROM level_progressions WHERE id=$1`, id).
		Scan(&lp.ID, &lp.StudentID, &lp.SchoolID, &lp.FromSessionID, &lp.ToSessionID,
			&lp.FromLevelID, &lp.ToLevelID, &lp.Decision, &lp.DecidedBy,
			&lp.DecisionDate, &lp.Remarks, &lp.CreatedAt, &lp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("progression", id)
	}
	return lp, err
}

func (r *levelProgressionRepo) ListByStudent(ctx context.Context, studentID string) ([]*domain.LevelProgression, error) {
	return r.list(ctx, "student_id=$1", studentID)
}

func (r *levelProgressionRepo) ListBySession(ctx context.Context, schoolID, sessionID string) ([]*domain.LevelProgression, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,student_id,school_id,from_session_id,COALESCE(to_session_id,''),
		       from_level_id,COALESCE(to_level_id,''),decision,decided_by,decision_date,
		       COALESCE(remarks,''),created_at,updated_at
		FROM level_progressions WHERE school_id=$1 AND from_session_id=$2 ORDER BY decision_date`,
		schoolID, sessionID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectProgressions(rows)
}

func (r *levelProgressionRepo) list(ctx context.Context, where, arg string) ([]*domain.LevelProgression, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,student_id,school_id,from_session_id,COALESCE(to_session_id,''),
		       from_level_id,COALESCE(to_level_id,''),decision,decided_by,decision_date,
		       COALESCE(remarks,''),created_at,updated_at
		FROM level_progressions WHERE `+where+` ORDER BY decision_date DESC`, arg)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectProgressions(rows)
}

func collectProgressions(rows *sql.Rows) ([]*domain.LevelProgression, error) {
	var out []*domain.LevelProgression
	for rows.Next() {
		lp := &domain.LevelProgression{}
		if err := rows.Scan(&lp.ID, &lp.StudentID, &lp.SchoolID,
			&lp.FromSessionID, &lp.ToSessionID, &lp.FromLevelID, &lp.ToLevelID,
			&lp.Decision, &lp.DecidedBy, &lp.DecisionDate,
			&lp.Remarks, &lp.CreatedAt, &lp.UpdatedAt); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, lp)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// SCORE SHEET REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type scoreSheetRepo struct{ db *sql.DB }

func NewScoreSheetRepository(db *sql.DB) domain.ScoreSheetRepository { return &scoreSheetRepo{db: db} }

func (r *scoreSheetRepo) Upsert(ctx context.Context, ss *domain.ScoreSheet) error {
	if ss.ID == "" {
		ss.ID = uuid.NewString()
	}
	ss.RecordedAt = time.Now()
	ss.UpdatedAt = time.Now()
	if ss.CreatedAt.IsZero() {
		ss.CreatedAt = ss.UpdatedAt
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO score_sheets
		(id,enrollment_id,student_id,school_id,session_id,term_id,subject_id,
		 ca1_score,ca2_score,ca3_score,exam_score,total_score,grade,remark,
		 position,recorded_by,recorded_at,created_at,updated_at,created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		ON CONFLICT (student_id,subject_id,term_id) DO UPDATE SET
		  ca1_score=EXCLUDED.ca1_score, ca2_score=EXCLUDED.ca2_score,
		  ca3_score=EXCLUDED.ca3_score, exam_score=EXCLUDED.exam_score,
		  total_score=EXCLUDED.total_score, grade=EXCLUDED.grade,
		  remark=EXCLUDED.remark, recorded_by=EXCLUDED.recorded_by,
		  recorded_at=EXCLUDED.recorded_at, updated_at=NOW()`,
		ss.ID, ss.EnrollmentID, ss.StudentID, ss.SchoolID, ss.SessionID, ss.TermID, ss.SubjectID,
		ss.CA1Score, ss.CA2Score, ss.CA3Score, ss.ExamScore, ss.TotalScore,
		ss.Grade, ss.Remark, ss.Position, ss.RecordedBy, ss.RecordedAt,
		ss.CreatedAt, ss.UpdatedAt, ss.CreatedBy)
	return err
}

func (r *scoreSheetRepo) GetByID(ctx context.Context, id string) (*domain.ScoreSheet, error) {
	return scanScoreSheet(r.db.QueryRowContext(ctx, scoreSheetSelect+" WHERE id=$1", id))
}

func (r *scoreSheetRepo) GetByStudentSubjectTerm(ctx context.Context, studentID, subjectID, termID string) (*domain.ScoreSheet, error) {
	return scanScoreSheet(r.db.QueryRowContext(ctx,
		scoreSheetSelect+" WHERE student_id=$1 AND subject_id=$2 AND term_id=$3",
		studentID, subjectID, termID))
}

func (r *scoreSheetRepo) ListByStudent(ctx context.Context, studentID, sessionID string) ([]*domain.ScoreSheet, error) {
	rows, err := r.db.QueryContext(ctx,
		scoreSheetSelect+" WHERE student_id=$1 AND session_id=$2 ORDER BY term_id,subject_id",
		studentID, sessionID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectScoreSheets(rows)
}

func (r *scoreSheetRepo) List(ctx context.Context, f domain.ScoreSheetFilter, p pagination.Params) ([]*domain.ScoreSheet, int, error) {
	where, args := []string{"1=1"}, []any{}
	idx := 1
	if f.SchoolID != "" {
		where = append(where, fmt.Sprintf("school_id=$%d", idx)); args = append(args, f.SchoolID); idx++
	}
	if f.SessionID != "" {
		where = append(where, fmt.Sprintf("session_id=$%d", idx)); args = append(args, f.SessionID); idx++
	}
	if f.TermID != "" {
		where = append(where, fmt.Sprintf("term_id=$%d", idx)); args = append(args, f.TermID); idx++
	}
	if f.SubjectID != "" {
		where = append(where, fmt.Sprintf("subject_id=$%d", idx)); args = append(args, f.SubjectID); idx++
	}
	clause := "WHERE " + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM score_sheets "+clause, args...).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}
	args = append(args, p.PerPage, p.Offset)
	q := fmt.Sprintf(scoreSheetSelect+" %s ORDER BY student_id,subject_id LIMIT $%d OFFSET $%d",
		clause, idx, idx+1)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()
	sheets, err := collectScoreSheets(rows)
	return sheets, total, err
}

// ComputePositions ranks students in a sub-level for a given subject and term.
func (r *scoreSheetRepo) ComputePositions(ctx context.Context, termID, subLevelID, subjectID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE score_sheets ss
		SET position = ranked.pos
		FROM (
		  SELECT ss2.id,
		         RANK() OVER (ORDER BY ss2.total_score DESC) AS pos
		  FROM score_sheets ss2
		  JOIN enrollments e ON e.id = ss2.enrollment_id
		  WHERE ss2.term_id=$1 AND e.sub_level_id=$2 AND ss2.subject_id=$3
		) ranked
		WHERE ss.id = ranked.id`, termID, subLevelID, subjectID)
	return err
}

const scoreSheetSelect = `
	SELECT id,enrollment_id,student_id,school_id,session_id,term_id,subject_id,
	       ca1_score,ca2_score,ca3_score,exam_score,total_score,
	       COALESCE(grade,''),COALESCE(remark,''),position,
	       COALESCE(recorded_by,''),COALESCE(recorded_at,NOW()),
	       created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM score_sheets`

func scanScoreSheet(s scanner) (*domain.ScoreSheet, error) {
	ss := &domain.ScoreSheet{}
	err := s.Scan(
		&ss.ID, &ss.EnrollmentID, &ss.StudentID, &ss.SchoolID, &ss.SessionID, &ss.TermID, &ss.SubjectID,
		&ss.CA1Score, &ss.CA2Score, &ss.CA3Score, &ss.ExamScore, &ss.TotalScore,
		&ss.Grade, &ss.Remark, &ss.Position, &ss.RecordedBy, &ss.RecordedAt,
		&ss.CreatedAt, &ss.UpdatedAt, &ss.CreatedBy, &ss.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("score_sheet", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return ss, nil
}

func collectScoreSheets(rows *sql.Rows) ([]*domain.ScoreSheet, error) {
	var out []*domain.ScoreSheet
	for rows.Next() {
		ss, err := scanScoreSheet(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ss)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// GRADE CONFIG REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type gradeConfigRepo struct{ db *sql.DB }

func NewGradeConfigRepository(db *sql.DB) domain.GradeConfigRepository { return &gradeConfigRepo{db: db} }

func (r *gradeConfigRepo) Upsert(ctx context.Context, gc *domain.GradeConfig) error {
	if gc.ID == "" {
		gc.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO grade_configs (id,state_id,school_id,grade,min_score,max_score,remark,points,created_at,updated_at,created_by)
		VALUES ($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,NOW(),NOW(),$9)
		ON CONFLICT (state_id,school_id,grade) DO UPDATE SET
		  min_score=EXCLUDED.min_score, max_score=EXCLUDED.max_score,
		  remark=EXCLUDED.remark, points=EXCLUDED.points, updated_at=NOW()`,
		gc.ID, gc.StateID, gc.SchoolID, gc.Grade, gc.MinScore, gc.MaxScore,
		gc.Remark, gc.Points, gc.CreatedBy)
	return err
}

func (r *gradeConfigRepo) ListBySchool(ctx context.Context, schoolID string) ([]*domain.GradeConfig, error) {
	rows, err := r.db.QueryContext(ctx,
		gradeConfigSelect+" WHERE school_id=$1 ORDER BY min_score DESC", schoolID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectGradeConfigs(rows)
}

func (r *gradeConfigRepo) ListStateDefault(ctx context.Context, stateID string) ([]*domain.GradeConfig, error) {
	rows, err := r.db.QueryContext(ctx,
		gradeConfigSelect+" WHERE state_id=$1 AND school_id IS NULL ORDER BY min_score DESC", stateID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectGradeConfigs(rows)
}

func (r *gradeConfigRepo) EvaluateGrade(ctx context.Context, score float64, schoolID, stateID string) (*domain.GradeConfig, error) {
	// Try school-specific first, fall back to state default
	gc := &domain.GradeConfig{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id,state_id,COALESCE(school_id,''),grade,min_score,max_score,remark,points,
		       created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		FROM grade_configs
		WHERE state_id=$1 AND (school_id=$2 OR school_id IS NULL)
		  AND $3 BETWEEN min_score AND max_score
		ORDER BY school_id DESC NULLS LAST
		LIMIT 1`, stateID, schoolID, score).
		Scan(&gc.ID, &gc.StateID, &gc.SchoolID, &gc.Grade, &gc.MinScore, &gc.MaxScore,
			&gc.Remark, &gc.Points, &gc.CreatedAt, &gc.UpdatedAt, &gc.CreatedBy, &gc.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("grade_config", "for score")
	}
	return gc, err
}

const gradeConfigSelect = `
	SELECT id,state_id,COALESCE(school_id,''),grade,min_score,max_score,remark,points,
	       created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM grade_configs`

func collectGradeConfigs(rows *sql.Rows) ([]*domain.GradeConfig, error) {
	var out []*domain.GradeConfig
	for rows.Next() {
		gc := &domain.GradeConfig{}
		if err := rows.Scan(&gc.ID, &gc.StateID, &gc.SchoolID, &gc.Grade,
			&gc.MinScore, &gc.MaxScore, &gc.Remark, &gc.Points,
			&gc.CreatedAt, &gc.UpdatedAt, &gc.CreatedBy, &gc.UpdatedBy); err != nil {
			return nil, apperror.Internal(err)
		}
		out = append(out, gc)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// SCORE CONFIG REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type scoreConfigRepo struct{ db *sql.DB }

func NewScoreConfigRepository(db *sql.DB) domain.ScoreConfigRepository { return &scoreConfigRepo{db: db} }

func (r *scoreConfigRepo) Upsert(ctx context.Context, sc *domain.ScoreConfig) error {
	if sc.ID == "" {
		sc.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO score_configs (id,state_id,school_id,ca1_max,ca2_max,ca3_max,exam_max,total_max,created_at,updated_at,created_by)
		VALUES ($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,NOW(),NOW(),$9)
		ON CONFLICT (state_id,school_id) DO UPDATE SET
		  ca1_max=EXCLUDED.ca1_max, ca2_max=EXCLUDED.ca2_max,
		  ca3_max=EXCLUDED.ca3_max, exam_max=EXCLUDED.exam_max,
		  total_max=EXCLUDED.total_max, updated_at=NOW()`,
		sc.ID, sc.StateID, sc.SchoolID, sc.CA1Max, sc.CA2Max, sc.CA3Max,
		sc.ExamMax, sc.TotalMax, sc.CreatedBy)
	return err
}

func (r *scoreConfigRepo) GetBySchool(ctx context.Context, schoolID string) (*domain.ScoreConfig, error) {
	sc := &domain.ScoreConfig{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,state_id,COALESCE(school_id,''),ca1_max,ca2_max,ca3_max,exam_max,total_max,
		        created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM score_configs WHERE school_id=$1`, schoolID).
		Scan(&sc.ID, &sc.StateID, &sc.SchoolID, &sc.CA1Max, &sc.CA2Max, &sc.CA3Max,
			&sc.ExamMax, &sc.TotalMax, &sc.CreatedAt, &sc.UpdatedAt, &sc.CreatedBy, &sc.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("score_config", schoolID)
	}
	return sc, err
}

func (r *scoreConfigRepo) GetStateDefault(ctx context.Context, stateID string) (*domain.ScoreConfig, error) {
	sc := &domain.ScoreConfig{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,state_id,COALESCE(school_id,''),ca1_max,ca2_max,ca3_max,exam_max,total_max,
		        created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM score_configs WHERE state_id=$1 AND school_id IS NULL`, stateID).
		Scan(&sc.ID, &sc.StateID, &sc.SchoolID, &sc.CA1Max, &sc.CA2Max, &sc.CA3Max,
			&sc.ExamMax, &sc.TotalMax, &sc.CreatedAt, &sc.UpdatedAt, &sc.CreatedBy, &sc.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("score_config", "state default")
	}
	return sc, err
}

// ─────────────────────────────────────────────────────────────────────────────
// REPORT CARD REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type reportCardRepo struct{ db *sql.DB }

func NewReportCardRepository(db *sql.DB) domain.ReportCardRepository { return &reportCardRepo{db: db} }

func (r *reportCardRepo) Upsert(ctx context.Context, rc *domain.ReportCard) error {
	if rc.ID == "" {
		rc.ID = uuid.NewString()
	}
	rc.UpdatedAt = time.Now()
	if rc.CreatedAt.IsZero() {
		rc.CreatedAt = rc.UpdatedAt
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO report_cards
		(id,student_id,school_id,session_id,term_id,level_id,sub_level_id,
		 total_score,average_score,overall_grade,class_position,subject_count,
		 attendance,total_school_days,principal_remark,teacher_remark,
		 published_at,created_at,updated_at,created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NULLIF($15,''),NULLIF($16,''),$17,$18,$19,$20)
		ON CONFLICT (student_id,term_id) DO UPDATE SET
		  total_score=EXCLUDED.total_score, average_score=EXCLUDED.average_score,
		  overall_grade=EXCLUDED.overall_grade, class_position=EXCLUDED.class_position,
		  subject_count=EXCLUDED.subject_count, attendance=EXCLUDED.attendance,
		  total_school_days=EXCLUDED.total_school_days,
		  principal_remark=EXCLUDED.principal_remark, teacher_remark=EXCLUDED.teacher_remark,
		  updated_at=NOW()`,
		rc.ID, rc.StudentID, rc.SchoolID, rc.SessionID, rc.TermID, rc.LevelID, rc.SubLevelID,
		rc.TotalScore, rc.AverageScore, rc.OverallGrade, rc.ClassPosition, rc.SubjectCount,
		rc.Attendance, rc.TotalSchoolDays, rc.PrincipalRemark, rc.TeacherRemark,
		rc.PublishedAt, rc.CreatedAt, rc.UpdatedAt, rc.CreatedBy)
	return err
}

func (r *reportCardRepo) GetByID(ctx context.Context, id string) (*domain.ReportCard, error) {
	return scanReportCard(r.db.QueryRowContext(ctx, reportCardSelect+" WHERE id=$1", id))
}

func (r *reportCardRepo) GetByStudentTerm(ctx context.Context, studentID, termID string) (*domain.ReportCard, error) {
	return scanReportCard(r.db.QueryRowContext(ctx,
		reportCardSelect+" WHERE student_id=$1 AND term_id=$2", studentID, termID))
}

func (r *reportCardRepo) ListByTerm(ctx context.Context, schoolID, termID string, p pagination.Params) ([]*domain.ReportCard, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM report_cards WHERE school_id=$1 AND term_id=$2`,
		schoolID, termID).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}
	rows, err := r.db.QueryContext(ctx,
		reportCardSelect+` WHERE school_id=$1 AND term_id=$2 ORDER BY class_position LIMIT $3 OFFSET $4`,
		schoolID, termID, p.PerPage, p.Offset)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()
	rcs, err := collectReportCards(rows)
	return rcs, total, err
}

func (r *reportCardRepo) ListByStudent(ctx context.Context, studentID string) ([]*domain.ReportCard, error) {
	rows, err := r.db.QueryContext(ctx,
		reportCardSelect+" WHERE student_id=$1 ORDER BY created_at DESC", studentID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectReportCards(rows)
}

func (r *reportCardRepo) Publish(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE report_cards SET published_at=NOW(),updated_at=NOW() WHERE id=$1 AND published_at IS NULL`, id)
	return checkRowsAffected(res, err, "report_card", id)
}

const reportCardSelect = `
	SELECT id,student_id,school_id,session_id,term_id,level_id,sub_level_id,
	       total_score,average_score,COALESCE(overall_grade,''),class_position,subject_count,
	       attendance,total_school_days,COALESCE(principal_remark,''),COALESCE(teacher_remark,''),
	       published_at,created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
	FROM report_cards`

func scanReportCard(s scanner) (*domain.ReportCard, error) {
	rc := &domain.ReportCard{}
	err := s.Scan(
		&rc.ID, &rc.StudentID, &rc.SchoolID, &rc.SessionID, &rc.TermID,
		&rc.LevelID, &rc.SubLevelID,
		&rc.TotalScore, &rc.AverageScore, &rc.OverallGrade, &rc.ClassPosition, &rc.SubjectCount,
		&rc.Attendance, &rc.TotalSchoolDays, &rc.PrincipalRemark, &rc.TeacherRemark,
		&rc.PublishedAt, &rc.CreatedAt, &rc.UpdatedAt, &rc.CreatedBy, &rc.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("report_card", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return rc, nil
}

func collectReportCards(rows *sql.Rows) ([]*domain.ReportCard, error) {
	var out []*domain.ReportCard
	for rows.Next() {
		rc, err := scanReportCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}
