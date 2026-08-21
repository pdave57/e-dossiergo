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
// STATE REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type stateRepo struct{ db *sql.DB }

func NewStateRepository(db *sql.DB) domain.StateRepository { return &stateRepo{db: db} }

func (r *stateRepo) Create(ctx context.Context, s *domain.State) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO states (id,name,code,country,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		s.ID, s.Name, s.Code, s.Country, s.CreatedAt, s.UpdatedAt, s.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("state code already exists")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *stateRepo) GetByID(ctx context.Context, id string) (*domain.State, error) {
	return scanState(r.db.QueryRowContext(ctx,
		`SELECT id,name,code,country,created_at,updated_at,
		        COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM states WHERE id=$1 AND deleted_at IS NULL`, id))
}

func (r *stateRepo) GetByCode(ctx context.Context, code string) (*domain.State, error) {
	return scanState(r.db.QueryRowContext(ctx,
		`SELECT id,name,code,country,created_at,updated_at,
		        COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM states WHERE code=$1 AND deleted_at IS NULL`, code))
}

func (r *stateRepo) Update(ctx context.Context, s *domain.State) error {
	s.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE states SET name=$1,country=$2,updated_at=$3,updated_by=$4 WHERE id=$5 AND deleted_at IS NULL`,
		s.Name, s.Country, s.UpdatedAt, s.UpdatedBy, s.ID)
	return checkRowsAffected(res, err, "state", s.ID)
}

func (r *stateRepo) List(ctx context.Context) ([]*domain.State, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,code,country,created_at,updated_at,
		        COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM states WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.State
	for rows.Next() {
		s, err := scanState(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func scanState(s scanner) (*domain.State, error) {
	st := &domain.State{}
	err := s.Scan(&st.ID, &st.Name, &st.Code, &st.Country,
		&st.CreatedAt, &st.UpdatedAt, &st.CreatedBy, &st.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("state", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return st, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ZONE REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type zoneRepo struct{ db *sql.DB }

func NewZoneRepository(db *sql.DB) domain.ZoneRepository { return &zoneRepo{db: db} }

func (r *zoneRepo) Create(ctx context.Context, z *domain.Zone) error {
	if z.ID == "" {
		z.ID = uuid.NewString()
	}
	now := time.Now()
	z.CreatedAt, z.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO zones (id,state_id,name,code,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		z.ID, z.StateID, z.Name, z.Code, z.CreatedAt, z.UpdatedAt, z.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("zone code already exists in this state")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *zoneRepo) GetByID(ctx context.Context, id string) (*domain.Zone, error) {
	return scanZone(r.db.QueryRowContext(ctx,
		`SELECT id,state_id,name,code,created_at,updated_at,
		        COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM zones WHERE id=$1 AND deleted_at IS NULL`, id))
}

func (r *zoneRepo) Update(ctx context.Context, z *domain.Zone) error {
	z.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE zones SET name=$1,code=$2,updated_at=$3,updated_by=$4 WHERE id=$5 AND deleted_at IS NULL`,
		z.Name, z.Code, z.UpdatedAt, z.UpdatedBy, z.ID)
	return checkRowsAffected(res, err, "zone", z.ID)
}

func (r *zoneRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE zones SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "zone", id)
}

func (r *zoneRepo) ListByState(ctx context.Context, stateID string) ([]*domain.Zone, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,state_id,name,code,created_at,updated_at,
		        COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM zones WHERE state_id=$1 AND deleted_at IS NULL ORDER BY name`, stateID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.Zone
	for rows.Next() {
		z, err := scanZone(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, z)
	}
	return out, rows.Err()
}

func scanZone(s scanner) (*domain.Zone, error) {
	z := &domain.Zone{}
	err := s.Scan(&z.ID, &z.StateID, &z.Name, &z.Code,
		&z.CreatedAt, &z.UpdatedAt, &z.CreatedBy, &z.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("zone", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return z, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// LGA REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type lgaRepo struct{ db *sql.DB }

func NewLGARepository(db *sql.DB) domain.LGARepository { return &lgaRepo{db: db} }

func (r *lgaRepo) Create(ctx context.Context, l *domain.LGA) error {
	if l.ID == "" {
		l.ID = uuid.NewString()
	}
	now := time.Now()
	l.CreatedAt, l.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO lgas (id,state_id,zone_id,name,code,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		l.ID, l.StateID, l.ZoneID, l.Name, l.Code, l.CreatedAt, l.UpdatedAt, l.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("LGA code already exists in this state")
		}
		if isFKViolation(err) {
			return apperror.BadRequest("invalid zone or state reference")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *lgaRepo) GetByID(ctx context.Context, id string) (*domain.LGA, error) {
	return scanLGA(r.db.QueryRowContext(ctx,
		`SELECT lgas.id,state_id,zone_id,name,code,zones.name,created_at,updated_at,
		        COALESCE(lgas.created_by,''),COALESCE(lgas.updated_by,'')
		 FROM lgas LEFT JOIN zones ON zones.id=lgas.zone_id WHERE lgas.id=$1 AND lgas.deleted_at IS NULL`, id))
}

func (r *lgaRepo) Update(ctx context.Context, l *domain.LGA) error {
	l.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE lgas SET state_id=$1,zone_id=$2,name=$3,code=$4,updated_at=$5,updated_by=$6
		 WHERE id=$7 AND deleted_at IS NULL`,
		l.StateID, l.ZoneID, l.Name, l.Code, l.UpdatedAt, l.UpdatedBy, l.ID)
	if err != nil {
		fmt.Printf("LGA_UPDATE_RAW_ERR type=%T err=%+v\n", err, err)
		if isUniqueViolation(err) {
			return apperror.Conflict("LGA code already exists in this state")
		}
		if isFKViolation(err) {
			return apperror.BadRequest("invalid zone or state reference")
		}
		return apperror.Internal(err)
	}
	return checkRowsAffected(res, nil, "lga", l.ID)
}

func (r *lgaRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE lgas SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "lga", id)
}

func (r *lgaRepo) ListByState(ctx context.Context, stateID string) ([]*domain.LGA, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT lgas.id,lgas.state_id,lgas.zone_id,lgas.name,lgas.code,zones.name,lgas.created_at,lgas.updated_at,
		        COALESCE(lgas.created_by,''),COALESCE(lgas.updated_by,'')
		 FROM lgas LEFT JOIN zones ON zones.id=lgas.zone_id WHERE lgas.state_id=$1 AND lgas.deleted_at IS NULL ORDER BY lgas.name`, stateID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectLGAs(rows)
}

func (r *lgaRepo) ListByZone(ctx context.Context, zoneID string) ([]*domain.LGA, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT lgas.id,state_id,zone_id,name,code,zones.name,created_at,updated_at,
		        COALESCE(lgas.created_by,''),COALESCE(lgas.updated_by,'')
		 FROM lgas LEFT JOIN zones ON zones.id=lgas.zone_id WHERE zone_id=$1 AND lgas.deleted_at IS NULL ORDER BY lgas.name`, zoneID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	return collectLGAs(rows)
}

func collectLGAs(rows *sql.Rows) ([]*domain.LGA, error) {
	var out []*domain.LGA
	for rows.Next() {
		l, err := scanLGA(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func scanLGA(s scanner) (*domain.LGA, error) {
	l := &domain.LGA{}
	err := s.Scan(&l.ID, &l.StateID, &l.ZoneID, &l.Name, &l.Code, &l.ZoneName,
		&l.CreatedAt, &l.UpdatedAt, &l.CreatedBy, &l.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("lga", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return l, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type schoolRepo struct{ db *sql.DB }

func NewSchoolRepository(db *sql.DB) domain.SchoolRepository { return &schoolRepo{db: db} }

func (r *schoolRepo) Create(ctx context.Context, s *domain.School) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO schools
		 (id,state_id,zone_id,lga_id,name,code,category,ownership,status,address,
		  head_teacher,founded,number_of_classrooms,total_students,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		s.ID, s.StateID, s.ZoneID, s.LGAID, s.Name, s.Code,
		s.Category, s.Ownership, s.Status, s.Address,
		nullableString(s.HeadTeacher),
		s.Founded, s.NumberOfClassrooms, s.TotalStudents, s.CreatedAt, s.UpdatedAt, s.CreatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("school code already exists")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (r *schoolRepo) GetByID(ctx context.Context, id string) (*domain.School, error) {
	return scanSchool(r.db.QueryRowContext(ctx, schoolSelectSQL+" WHERE s.id=$1 AND s.deleted_at IS NULL", id))
}

func (r *schoolRepo) GetByCode(ctx context.Context, code string) (*domain.School, error) {
	return scanSchool(r.db.QueryRowContext(ctx, schoolSelectSQL+" WHERE s.code=$1 AND s.deleted_at IS NULL", code))
}

func (r *schoolRepo) Update(ctx context.Context, s *domain.School) error {
	s.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE schools SET zone_id=$1,lga_id=$2,name=$3,category=$4,ownership=$5,status=$6,
		  address=$7,head_teacher=$8,founded=$9,number_of_classrooms=$10,total_students=$11,updated_at=$12,updated_by=$13
		 WHERE id=$14 AND deleted_at IS NULL`,
		s.ZoneID, s.LGAID, s.Name, s.Category, s.Ownership, s.Status,
		s.Address, nullableString(s.HeadTeacher), s.Founded, s.NumberOfClassrooms, s.TotalStudents, s.UpdatedAt, s.UpdatedBy, s.ID)
	return checkRowsAffected(res, err, "school", s.ID)
}

func (r *schoolRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE schools SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "school", id)
}

func (r *schoolRepo) List(ctx context.Context, f domain.SchoolFilter, p pagination.Params) ([]*domain.School, int, error) {
	where, args := []string{"s.deleted_at IS NULL"}, []any{}
	idx := 1
	if f.StateID != "" {
		where = append(where, fmt.Sprintf("s.state_id=$%d", idx))
		args = append(args, f.StateID)
		idx++
	}
	if f.ZoneID != "" {
		where = append(where, fmt.Sprintf("s.zone_id=$%d", idx))
		args = append(args, f.ZoneID)
		idx++
	}
	if f.LGAID != "" {
		where = append(where, fmt.Sprintf("s.lga_id=$%d", idx))
		args = append(args, f.LGAID)
		idx++
	}
	if f.Category != "" {
		where = append(where, fmt.Sprintf("s.category=$%d", idx))
		args = append(args, f.Category)
		idx++
	}
	if f.Ownership != "" {
		where = append(where, fmt.Sprintf("s.ownership=$%d", idx))
		args = append(args, f.Ownership)
		idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("s.status=$%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(s.name ILIKE $%d OR s.code ILIKE $%d)", idx, idx))
		args = append(args, "%"+f.Search+"%")
		idx++
	}

	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schools s "+clause, args...).Scan(&total); err != nil {
		return nil, 0, apperror.Internal(err)
	}

	args = append(args, p.PerPage, p.Offset)
	q := fmt.Sprintf("%s %s ORDER BY s.name LIMIT $%d OFFSET $%d",
		schoolSelectSQL, clause, idx, idx+1)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	defer rows.Close()

	var out []*domain.School
	for rows.Next() {
		s, err := scanSchool(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func (r *schoolRepo) CountTotalSchools(ctx context.Context, stateID string) (int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schools WHERE state_id=$1 AND deleted_at IS NULL`,
		stateID).Scan(&total); err != nil {
		return 0, apperror.Internal(err)
	}
	return total, nil
}

func (r *schoolRepo) UpdateLogo(ctx context.Context, id, logoURL string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE schools SET logo_url=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
		logoURL, id)
	return checkRowsAffected(res, err, "school", id)
}

const schoolSelectSQL = `
	SELECT s.id,s.state_id,s.zone_id,s.lga_id,s.name,s.code,s.category,s.ownership,s.status,
	       COALESCE(s.address,''),COALESCE(s.head_teacher,''),s.founded,
	       COALESCE(s.number_of_classrooms,0),COALESCE(s.total_students,0),COALESCE(s.logo_url,''),
	       s.created_at,s.updated_at,
	       COALESCE(s.created_by,''),COALESCE(s.updated_by,'')
	FROM schools s`

func scanSchool(s scanner) (*domain.School, error) {
	sc := &domain.School{}
	err := s.Scan(
		&sc.ID, &sc.StateID, &sc.ZoneID, &sc.LGAID, &sc.Name, &sc.Code,
		&sc.Category, &sc.Ownership, &sc.Status,
		&sc.Address, &sc.HeadTeacher, &sc.Founded,
		&sc.NumberOfClassrooms, &sc.TotalStudents, &sc.LogoURL,
		&sc.CreatedAt, &sc.UpdatedAt, &sc.CreatedBy, &sc.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("school", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return sc, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL FACILITY REPOSITORY
// ─────────────────────────────────────────────────────────────────────────────

type schoolFacilityRepo struct{ db *sql.DB }

func NewSchoolFacilityRepository(db *sql.DB) domain.SchoolFacilityRepository {
	return &schoolFacilityRepo{db: db}
}

func (r *schoolFacilityRepo) Create(ctx context.Context, f *domain.SchoolFacility) error {
	if f.ID == "" {
		f.ID = uuid.NewString()
	}
	now := time.Now()
	f.CreatedAt, f.UpdatedAt = now, now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO school_facilities (id,school_id,type,name,quantity,condition,notes,created_at,updated_at,created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		f.ID, f.SchoolID, f.Type, f.Name, f.Quantity, f.Condition,
		nullableString(f.Notes), f.CreatedAt, f.UpdatedAt, f.CreatedBy)
	return apperror.Internal(err)
}

func (r *schoolFacilityRepo) GetByID(ctx context.Context, id string) (*domain.SchoolFacility, error) {
	return scanFacility(r.db.QueryRowContext(ctx,
		`SELECT id,school_id,type,name,quantity,condition,COALESCE(notes,''),
		        created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM school_facilities WHERE id=$1 AND deleted_at IS NULL`, id))
}

func (r *schoolFacilityRepo) Update(ctx context.Context, f *domain.SchoolFacility) error {
	f.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE school_facilities SET type=$1,name=$2,quantity=$3,condition=$4,
		  notes=$5,updated_at=$6,updated_by=$7 WHERE id=$8 AND deleted_at IS NULL`,
		f.Type, f.Name, f.Quantity, f.Condition, nullableString(f.Notes),
		f.UpdatedAt, f.UpdatedBy, f.ID)
	return checkRowsAffected(res, err, "facility", f.ID)
}

func (r *schoolFacilityRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE school_facilities SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return checkRowsAffected(res, err, "facility", id)
}

func (r *schoolFacilityRepo) ListBySchool(ctx context.Context, schoolID string) ([]*domain.SchoolFacility, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,school_id,type,name,quantity,condition,COALESCE(notes,''),
		        created_at,updated_at,COALESCE(created_by,''),COALESCE(updated_by,'')
		 FROM school_facilities WHERE school_id=$1 AND deleted_at IS NULL ORDER BY type,name`, schoolID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer rows.Close()
	var out []*domain.SchoolFacility
	for rows.Next() {
		f, err := scanFacility(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func scanFacility(s scanner) (*domain.SchoolFacility, error) {
	f := &domain.SchoolFacility{}
	err := s.Scan(&f.ID, &f.SchoolID, &f.Type, &f.Name, &f.Quantity, &f.Condition,
		&f.Notes, &f.CreatedAt, &f.UpdatedAt, &f.CreatedBy, &f.UpdatedBy)
	if err == sql.ErrNoRows {
		return nil, apperror.NotFound("facility", "")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return f, nil
}
