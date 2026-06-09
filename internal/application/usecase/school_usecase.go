package usecase

import (
	"context"
	"time"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/edossier/api/pkg/validator"
)

// ─────────────────────────────────────────────────────────────────────────────
// GEO USE CASES
// ─────────────────────────────────────────────────────────────────────────────

type GeoUseCase struct {
	states domain.StateRepository
	zones  domain.ZoneRepository
	lgas   domain.LGARepository
}

func NewGeoUseCase(states domain.StateRepository, zones domain.ZoneRepository, lgas domain.LGARepository) *GeoUseCase {
	return &GeoUseCase{states: states, zones: zones, lgas: lgas}
}

// — States —

func (uc *GeoUseCase) CreateState(ctx context.Context, req dto.CreateStateRequest, createdBy string) (*domain.State, error) {
	v := validator.New().Required(req.Name, "name").Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	s := &domain.State{Name: req.Name, Code: req.Code, Country: req.Country, AuditFields: domain.AuditFields{CreatedBy: createdBy}}
	if s.Country == "" {
		s.Country = "Nigeria"
	}
	return s, uc.states.Create(ctx, s)
}

func (uc *GeoUseCase) GetState(ctx context.Context, id string) (*domain.State, error) {
	return uc.states.GetByID(ctx, id)
}

func (uc *GeoUseCase) ListStates(ctx context.Context) ([]*domain.State, error) {
	return uc.states.List(ctx)
}

func (uc *GeoUseCase) UpdateState(ctx context.Context, id string, req dto.UpdateStateRequest, updatedBy string) (*domain.State, error) {
	s, err := uc.states.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.Name = req.Name
	if req.Country != "" {
		s.Country = req.Country
	}
	s.UpdatedBy = updatedBy
	return s, uc.states.Update(ctx, s)
}

// — Zones —

func (uc *GeoUseCase) CreateZone(ctx context.Context, stateID string, req dto.CreateZoneRequest, createdBy string) (*domain.Zone, error) {
	v := validator.New().Required(req.Name, "name").Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	z := &domain.Zone{StateID: stateID, Name: req.Name, Code: req.Code, AuditFields: domain.AuditFields{CreatedBy: createdBy}}
	return z, uc.zones.Create(ctx, z)
}

func (uc *GeoUseCase) ListZones(ctx context.Context, stateID string) ([]*domain.Zone, error) {
	return uc.zones.ListByState(ctx, stateID)
}

func (uc *GeoUseCase) UpdateZone(ctx context.Context, id string, req dto.UpdateZoneRequest, updatedBy string) (*domain.Zone, error) {
	z, err := uc.zones.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	z.Name = req.Name
	z.Code = req.Code
	z.UpdatedBy = updatedBy
	return z, uc.zones.Update(ctx, z)
}

func (uc *GeoUseCase) DeleteZone(ctx context.Context, id string) error {
	return uc.zones.Delete(ctx, id)
}

// — LGAs —

func (uc *GeoUseCase) CreateLGA(ctx context.Context, stateID string, req dto.CreateLGARequest, createdBy string) (*domain.LGA, error) {
	v := validator.New().Required(req.ZoneID, "zone_id").Required(req.Name, "name").Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	l := &domain.LGA{StateID: stateID, ZoneID: req.ZoneID, Name: req.Name, Code: req.Code, AuditFields: domain.AuditFields{CreatedBy: createdBy}}
	return l, uc.lgas.Create(ctx, l)
}

func (uc *GeoUseCase) ListLGAs(ctx context.Context, stateID string) ([]*domain.LGA, error) {
	return uc.lgas.ListByState(ctx, stateID)
}

func (uc *GeoUseCase) ListLGAsByZone(ctx context.Context, zoneID string) ([]*domain.LGA, error) {
	return uc.lgas.ListByZone(ctx, zoneID)
}

func (uc *GeoUseCase) UpdateLGA(ctx context.Context, id string, req dto.UpdateLGARequest, updatedBy string) (*domain.LGA, error) {
	l, err := uc.lgas.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	l.ZoneID = req.ZoneID
	l.Name = req.Name
	l.Code = req.Code
	l.UpdatedBy = updatedBy
	return l, uc.lgas.Update(ctx, l)
}

func (uc *GeoUseCase) DeleteLGA(ctx context.Context, id string) error {
	return uc.lgas.Delete(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type SchoolUseCase struct {
	schools    domain.SchoolRepository
	facilities domain.SchoolFacilityRepository
}

func NewSchoolUseCase(schools domain.SchoolRepository, facilities domain.SchoolFacilityRepository) *SchoolUseCase {
	return &SchoolUseCase{schools: schools, facilities: facilities}
}

func (uc *SchoolUseCase) Create(ctx context.Context, stateID string, req dto.CreateSchoolRequest, createdBy string) (*domain.School, error) {
	v := validator.New().
		Required(req.ZoneID, "zone_id").
		Required(req.LGAID, "lga_id").
		Required(req.Name, "name").
		Required(req.Code, "code").
		OneOf(string(req.Category), []string{"PRIMARY", "JUNIOR_SECONDARY", "SENIOR_SECONDARY", "COMBINED", "VOCATIONAL"}, "category").
		OneOf(string(req.Ownership), []string{"GOVERNMENT", "PRIVATE", "MISSION", "COMMUNITY"}, "ownership")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	s := &domain.School{
		StateID: stateID, ZoneID: req.ZoneID, LGAID: req.LGAID,
		Name: req.Name, Code: req.Code,
		Category: domain.SchoolCategory(req.Category), Ownership: domain.SchoolOwnership(req.Ownership),
		Status: domain.SchoolStatusActive, Address: req.Address,
		HeadTeacher: req.HeadTeacher, Founded: req.Founded,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return s, uc.schools.Create(ctx, s)
}

func (uc *SchoolUseCase) GetByID(ctx context.Context, id string) (*domain.School, error) {
	return uc.schools.GetByID(ctx, id)
}

func (uc *SchoolUseCase) List(ctx context.Context, f domain.SchoolFilter, p pagination.Params) ([]*domain.School, int, error) {
	return uc.schools.List(ctx, f, p)
}

func (uc *SchoolUseCase) Update(ctx context.Context, id string, req dto.UpdateSchoolRequest, updatedBy string) (*domain.School, error) {
	s, err := uc.schools.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.ZoneID = req.ZoneID
	s.LGAID = req.LGAID
	s.Name = req.Name
	s.Category = domain.SchoolCategory(req.Category)
	s.Ownership = domain.SchoolOwnership(req.Ownership)
	s.Address = req.Address	
	s.HeadTeacher = req.HeadTeacher
	s.Founded = req.Founded
	if req.Status != "" {
		s.Status = domain.SchoolStatus(req.Status)
	}
	s.UpdatedBy = updatedBy
	return s, uc.schools.Update(ctx, s)
}

func (uc *SchoolUseCase) Delete(ctx context.Context, id string) error {
	return uc.schools.Delete(ctx, id)
}

// — Facilities —

func (uc *SchoolUseCase) AddFacility(ctx context.Context, schoolID string, req dto.CreateFacilityRequest, createdBy string) (*domain.SchoolFacility, error) {
	v := validator.New().
		Required(req.Type, "type").
		Required(req.Name, "name").
		Required(req.Condition, "condition")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	f := &domain.SchoolFacility{
		SchoolID: schoolID, Type: domain.FacilityType(req.Type), Name: req.Name,
		Quantity: req.Quantity, Condition: domain.FacilityCondition(req.Condition), Notes: req.Notes,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return f, uc.facilities.Create(ctx, f)
}

func (uc *SchoolUseCase) ListFacilities(ctx context.Context, schoolID string) ([]*domain.SchoolFacility, error) {
	return uc.facilities.ListBySchool(ctx, schoolID)
}

func (uc *SchoolUseCase) UpdateFacility(ctx context.Context, id string, req dto.UpdateFacilityRequest, updatedBy string) (*domain.SchoolFacility, error) {
	f, err := uc.facilities.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	f.Type = domain.FacilityType(req.Type)
	f.Name = req.Name
	f.Quantity = req.Quantity
	f.Condition = domain.FacilityCondition(req.Condition)
	f.Notes = req.Notes
	f.UpdatedBy = updatedBy
	return f, uc.facilities.Update(ctx, f)
}

func (uc *SchoolUseCase) DeleteFacility(ctx context.Context, id string) error {
	return uc.facilities.Delete(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC SESSION USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type AcademicUseCase struct {
	sessions domain.AcademicSessionRepository
	terms    domain.TermRepository
}

func NewAcademicUseCase(sessions domain.AcademicSessionRepository, terms domain.TermRepository) *AcademicUseCase {
	return &AcademicUseCase{sessions: sessions, terms: terms}
}

func (uc *AcademicUseCase) CreateSession(ctx context.Context, stateID string, req dto.CreateSessionRequest, createdBy string) (*domain.AcademicSession, error) {
	v := validator.New().
		Required(req.Name, "name").
		Check(req.StartYear > 2000, "start_year", "must be a valid year").
		Check(req.EndYear > req.StartYear, "end_year", "must be after start_year").
		Check(!req.StartDate.IsZero(), "start_date", "is required").
		Check(!req.EndDate.IsZero(), "end_date", "is required").
		Check(req.EndDate.After(req.StartDate), "end_date", "must be after start_date")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	s := &domain.AcademicSession{
		StateID: stateID, Name: req.Name,
		StartYear: req.StartYear, EndYear: req.EndYear,
		Status: domain.SessionDraft, StartDate: req.StartDate, EndDate: req.EndDate,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return s, uc.sessions.Create(ctx, s)
}

func (uc *AcademicUseCase) GetSession(ctx context.Context, id string) (*domain.AcademicSession, error) {
	return uc.sessions.GetByID(ctx, id)
}

func (uc *AcademicUseCase) GetActiveSession(ctx context.Context, stateID string) (*domain.AcademicSession, error) {
	return uc.sessions.GetActive(ctx, stateID)
}

func (uc *AcademicUseCase) ListSessions(ctx context.Context, stateID string, p pagination.Params) ([]*domain.AcademicSession, int, error) {
	return uc.sessions.List(ctx, stateID, p)
}

func (uc *AcademicUseCase) UpdateSession(ctx context.Context, id string, req dto.UpdateSessionRequest, updatedBy string) (*domain.AcademicSession, error) {
	s, err := uc.sessions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.Name = req.Name
	if !req.StartDate.IsZero() {
		s.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		s.EndDate = req.EndDate
	}
	if req.Status != "" {
		s.Status = domain.SessionStatus(req.Status)
	}
	s.UpdatedBy = updatedBy
	return s, uc.sessions.Update(ctx, s)
}

func (uc *AcademicUseCase) ActivateSession(ctx context.Context, id, stateID string) error {
	return uc.sessions.SetActive(ctx, id, stateID)
}

func (uc *AcademicUseCase) DeleteSession(ctx context.Context, id string) error {
	session, err := uc.sessions.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if session.Status == domain.SessionActive {
		return apperror.BadRequest("cannot delete an active session; close it first")
	}
	return uc.sessions.Delete(ctx, id)
}

// — Terms —

func (uc *AcademicUseCase) CreateTerm(ctx context.Context, sessionID string, req dto.CreateTermRequest, createdBy string) (*domain.Term, error) {
	v := validator.New().
		Required(req.Name, "name").
		Check(req.Number >= 1 && req.Number <= 3, "term_number", "must be 1, 2, or 3").
		Check(!req.StartDate.IsZero(), "start_date", "is required").
		Check(!req.EndDate.IsZero(), "end_date", "is required")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	t := &domain.Term{
		SessionID: sessionID, Number: req.Number, Name: req.Name,
		StartDate: req.StartDate, EndDate: req.EndDate, IsActive: false,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return t, uc.terms.Create(ctx, t)
}

func (uc *AcademicUseCase) ListTerms(ctx context.Context, sessionID string) ([]*domain.Term, error) {
	return uc.terms.ListBySession(ctx, sessionID)
}

func (uc *AcademicUseCase) UpdateTerm(ctx context.Context, id string, req dto.UpdateTermRequest, updatedBy string) (*domain.Term, error) {
	t, err := uc.terms.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Name = req.Name
	if !req.StartDate.IsZero() {
		t.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		t.EndDate = req.EndDate
	}
	t.UpdatedBy = updatedBy
	return t, uc.terms.Update(ctx, t)
}

func (uc *AcademicUseCase) ActivateTerm(ctx context.Context, id, sessionID string) error {
	return uc.terms.SetActive(ctx, id, sessionID)
}

func (uc *AcademicUseCase) DeleteTerm(ctx context.Context, id string) error {
	return uc.terms.Delete(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type LevelUseCase struct {
	levels       domain.LevelRepository
	subLevels    domain.SubLevelRepository
	schoolLevels domain.SchoolLevelRepository
}

func NewLevelUseCase(
	levels domain.LevelRepository,
	subLevels domain.SubLevelRepository,
	schoolLevels domain.SchoolLevelRepository,
) *LevelUseCase {
	return &LevelUseCase{levels: levels, subLevels: subLevels, schoolLevels: schoolLevels}
}

func (uc *LevelUseCase) CreateLevel(ctx context.Context, stateID string, req dto.CreateLevelRequest, createdBy string) (*domain.Level, error) {
	v := validator.New().
		Required(req.Name, "name").
		Required(req.Code, "code").
		OneOf(req.Type, []string{"PRIMARY", "JSS", "SSS", "VOCATIONAL"}, "type")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	l := &domain.Level{
		StateID: stateID, Name: req.Name, Code: req.Code,
		Type: domain.LevelType(req.Type), Order: req.Order,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return l, uc.levels.Create(ctx, l)
}

func (uc *LevelUseCase) ListLevels(ctx context.Context, stateID string) ([]*domain.Level, error) {
	return uc.levels.ListByState(ctx, stateID)
}

func (uc *LevelUseCase) GetLevel(ctx context.Context, id string) (*domain.Level, error) {
	return uc.levels.GetByID(ctx, id)
}

func (uc *LevelUseCase) UpdateLevel(ctx context.Context, id string, req dto.UpdateLevelRequest, updatedBy string) (*domain.Level, error) {
	l, err := uc.levels.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	l.Name = req.Name
	l.Code = req.Code
	l.Type = domain.LevelType(req.Type)
	l.Order = req.Order
	l.UpdatedBy = updatedBy
	return l, uc.levels.Update(ctx, l)
}

func (uc *LevelUseCase) DeleteLevel(ctx context.Context, id string) error {
	return uc.levels.Delete(ctx, id)
}

// — Sub-Levels —

func (uc *LevelUseCase) CreateSubLevel(ctx context.Context, schoolID string, req dto.CreateSubLevelRequest, createdBy string) (*domain.SubLevel, error) {
	v := validator.New().
		Required(req.LevelID, "level_id").
		Required(req.Name, "name").
		Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	if req.Capacity == 0 {
		req.Capacity = 40
	}
	sl := &domain.SubLevel{
		SchoolID: schoolID, LevelID: req.LevelID, Name: req.Name,
		Code: req.Code, Capacity: req.Capacity,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return sl, uc.subLevels.Create(ctx, sl)
}

func (uc *LevelUseCase) ListSubLevels(ctx context.Context, schoolID, levelID string) ([]*domain.SubLevel, error) {
	return uc.subLevels.ListByLevel(ctx, schoolID, levelID)
}

func (uc *LevelUseCase) UpdateSubLevel(ctx context.Context, id string, req dto.UpdateSubLevelRequest, updatedBy string) (*domain.SubLevel, error) {
	sl, err := uc.subLevels.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	sl.Name = req.Name
	sl.Code = req.Code
	if req.Capacity > 0 {
		sl.Capacity = req.Capacity
	}
	sl.UpdatedBy = updatedBy
	return sl, uc.subLevels.Update(ctx, sl)
}

func (uc *LevelUseCase) DeleteSubLevel(ctx context.Context, id string) error {
	return uc.subLevels.Delete(ctx, id)
}

// — School Levels (which levels a school offers) —

func (uc *LevelUseCase) UpsertSchoolLevel(ctx context.Context, schoolID string, req dto.UpsertSchoolLevelRequest, createdBy string) error {
	return uc.schoolLevels.Upsert(ctx, &domain.SchoolLevel{
		SchoolID: schoolID, LevelID: req.LevelID, SessionID: req.SessionID,
		IsActive: req.IsActive, AuditFields: domain.AuditFields{CreatedBy: createdBy},
	})
}

func (uc *LevelUseCase) ListSchoolLevels(ctx context.Context, schoolID, sessionID string) ([]*domain.SchoolLevel, error) {
	return uc.schoolLevels.ListBySchool(ctx, schoolID, sessionID)
}

func (uc *LevelUseCase) RemoveSchoolLevel(ctx context.Context, schoolID, levelID, sessionID string) error {
	return uc.schoolLevels.Delete(ctx, schoolID, levelID, sessionID)
}

// ─────────────────────────────────────────────────────────────────────────────
// SUBJECT USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type SubjectUseCase struct {
	subjects       domain.SubjectRepository
	schoolSubjects domain.SchoolSubjectRepository
}

func NewSubjectUseCase(subjects domain.SubjectRepository, schoolSubjects domain.SchoolSubjectRepository) *SubjectUseCase {
	return &SubjectUseCase{subjects: subjects, schoolSubjects: schoolSubjects}
}

func (uc *SubjectUseCase) Create(ctx context.Context, stateID string, req dto.CreateSubjectRequest, createdBy string) (*domain.Subject, error) {
	v := validator.New().
		Required(req.Name, "name").
		Required(req.Code, "code").
		OneOf(req.Category, []string{"CORE", "ELECTIVE", "PRACTICAL"}, "category").
		OneOf(req.LevelType, []string{"PRIMARY", "JSS", "SSS", "VOCATIONAL"}, "level_type")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	s := &domain.Subject{
		StateID: stateID, Name: req.Name, Code: req.Code,
		Category: domain.SubjectCategory(req.Category), LevelType: domain.LevelType(req.LevelType),
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return s, uc.subjects.Create(ctx, s)
}

func (uc *SubjectUseCase) GetByID(ctx context.Context, id string) (*domain.Subject, error) {
	return uc.subjects.GetByID(ctx, id)
}

func (uc *SubjectUseCase) ListByState(ctx context.Context, stateID string) ([]*domain.Subject, error) {
	return uc.subjects.ListByState(ctx, stateID)
}

func (uc *SubjectUseCase) Update(ctx context.Context, id string, req dto.UpdateSubjectRequest, updatedBy string) (*domain.Subject, error) {
	s, err := uc.subjects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.Name = req.Name
	s.Code = req.Code
	if req.Category != "" {
		s.Category = domain.SubjectCategory(req.Category)
	}
	if req.LevelType != "" {
		s.LevelType = domain.LevelType(req.LevelType)
	}
	s.UpdatedBy = updatedBy
	return s, uc.subjects.Update(ctx, s)
}

func (uc *SubjectUseCase) Delete(ctx context.Context, id string) error {
	return uc.subjects.Delete(ctx, id)
}

// — School Subjects —

func (uc *SubjectUseCase) AssignToSchool(ctx context.Context, schoolID string, req dto.CreateSchoolSubjectRequest, createdBy string) (*domain.SchoolSubject, error) {
	v := validator.New().
		Required(req.SubjectID, "subject_id").
		Required(req.LevelID, "level_id").
		Required(req.SessionID, "session_id")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	ss := &domain.SchoolSubject{
		SchoolID: schoolID, SubjectID: req.SubjectID, LevelID: req.LevelID,
		SessionID: req.SessionID, TeacherID: req.TeacherID, IsActive: true,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return ss, uc.schoolSubjects.Create(ctx, ss)
}

func (uc *SubjectUseCase) ListSchoolSubjects(ctx context.Context, schoolID, sessionID, levelID string) ([]*domain.SchoolSubject, error) {
	return uc.schoolSubjects.ListBySchool(ctx, schoolID, sessionID, levelID)
}

func (uc *SubjectUseCase) UpdateSchoolSubject(ctx context.Context, id string, req dto.UpdateSchoolSubjectRequest, updatedBy string) (*domain.SchoolSubject, error) {
	ss, err := uc.schoolSubjects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ss.TeacherID = req.TeacherID
	ss.IsActive = req.IsActive
	ss.UpdatedBy = updatedBy
	return ss, uc.schoolSubjects.Update(ctx, ss)
}

func (uc *SubjectUseCase) RemoveSchoolSubject(ctx context.Context, id string) error {
	return uc.schoolSubjects.Delete(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type PersonnelUseCase struct {
	personnel domain.PersonnelRepository
	transfers domain.PersonnelTransferRepository
	schools   domain.SchoolRepository
}

func NewPersonnelUseCase(
	personnel domain.PersonnelRepository,
	transfers domain.PersonnelTransferRepository,
	schools domain.SchoolRepository,
) *PersonnelUseCase {
	return &PersonnelUseCase{personnel: personnel, transfers: transfers, schools: schools}
}

func (uc *PersonnelUseCase) Create(ctx context.Context, stateID string, req dto.CreatePersonnelRequest, createdBy string) (*domain.Personnel, error) {
	v := validator.New().
		Required(req.SchoolID, "school_id").
		Required(req.StaffID, "staff_id").
		Required(req.FirstName, "first_name").
		Required(req.LastName, "last_name").
		OneOf(req.Gender, []string{"MALE", "FEMALE", "OTHER"}, "gender").
		Required(req.Role, "role")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	p := &domain.Personnel{
		StateID: stateID, SchoolID: req.SchoolID, StaffID: req.StaffID,
		FirstName: req.FirstName, MiddleName: req.MiddleName, LastName: req.LastName,
		Gender: domain.Gender(req.Gender), DateOfBirth: req.DateOfBirth,
		Email: req.Email, Phone: req.Phone, Address: req.Address,
		Role: domain.StaffRole(req.Role), Status: domain.StaffStatusActive,
		Qualification: req.Qualification, Specialization: req.Specialization,
		DateOfEmployment: req.DateOfEmployment, LGAID: req.LGAID,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return p, uc.personnel.Create(ctx, p)
}

func (uc *PersonnelUseCase) GetByID(ctx context.Context, id string) (*domain.Personnel, error) {
	return uc.personnel.GetByID(ctx, id)
}

func (uc *PersonnelUseCase) List(ctx context.Context, f domain.PersonnelFilter, p pagination.Params) ([]*domain.Personnel, int, error) {
	return uc.personnel.List(ctx, f, p)
}

func (uc *PersonnelUseCase) Update(ctx context.Context, id string, req dto.UpdatePersonnelRequest, updatedBy string) (*domain.Personnel, error) {
	p, err := uc.personnel.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.SchoolID != "" {
		p.SchoolID = req.SchoolID
	}
	p.FirstName = req.FirstName
	p.MiddleName = req.MiddleName
	p.LastName = req.LastName
	p.Gender = domain.Gender(req.Gender)
	p.DateOfBirth = req.DateOfBirth
	p.Email = req.Email
	p.Phone = req.Phone
	p.Address = req.Address
	p.Role = domain.StaffRole(req.Role)
	p.Qualification = req.Qualification
	p.Specialization = req.Specialization
	p.DateOfEmployment = req.DateOfEmployment
	if req.Status != "" {
		p.Status = domain.StaffStatus(req.Status)
	}
	p.UpdatedBy = updatedBy
	return p, uc.personnel.Update(ctx, p)
}

func (uc *PersonnelUseCase) Delete(ctx context.Context, id string) error {
	return uc.personnel.Delete(ctx, id)
}

// Transfer moves a staff member to another school and updates their school_id.
func (uc *PersonnelUseCase) Transfer(ctx context.Context, personnelID string, req dto.TransferPersonnelRequest, createdBy string) (*domain.PersonnelTransfer, error) {
	v := validator.New().
		Required(req.ToSchoolID, "to_school_id").
		Check(!req.TransferDate.IsZero(), "transfer_date", "is required")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	p, err := uc.personnel.GetByID(ctx, personnelID)
	if err != nil {
		return nil, err
	}
	if p.SchoolID == req.ToSchoolID {
		return nil, apperror.BadRequest("staff is already at the target school")
	}

	// Verify target school exists
	if _, err := uc.schools.GetByID(ctx, req.ToSchoolID); err != nil {
		return nil, apperror.NotFound("school", req.ToSchoolID)
	}

	transfer := &domain.PersonnelTransfer{
		PersonnelID:  personnelID,
		FromSchoolID: p.SchoolID,
		ToSchoolID:   req.ToSchoolID,
		TransferDate: req.TransferDate,
		Reason:       req.Reason,
		ApprovedBy:   req.ApprovedBy,
		AuditFields:  domain.AuditFields{CreatedBy: createdBy},
	}
	if err := uc.transfers.Create(ctx, transfer); err != nil {
		return nil, err
	}

	// Update staff's current school
	p.SchoolID = req.ToSchoolID
	p.Status = domain.StaffStatusActive
	p.UpdatedBy = createdBy
	if err := uc.personnel.Update(ctx, p); err != nil {
		return nil, err
	}
	return transfer, nil
}

func (uc *PersonnelUseCase) ListTransfers(ctx context.Context, personnelID string) ([]*domain.PersonnelTransfer, error) {
	return uc.transfers.ListByPersonnel(ctx, personnelID)
}

func (uc *PersonnelUseCase) ListSchoolTransfers(ctx context.Context, schoolID string) ([]*domain.PersonnelTransfer, error) {
	return uc.transfers.ListBySchool(ctx, schoolID)
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type StudentUseCase struct {
	students    domain.StudentRepository
	enrollments domain.EnrollmentRepository
	subLevels   domain.SubLevelRepository
	progressions domain.LevelProgressionRepository
	levels      domain.LevelRepository
}

func NewStudentUseCase(
	students domain.StudentRepository,
	enrollments domain.EnrollmentRepository,
	subLevels domain.SubLevelRepository,
	progressions domain.LevelProgressionRepository,
	levels domain.LevelRepository,
) *StudentUseCase {
	return &StudentUseCase{
		students: students, enrollments: enrollments,
		subLevels: subLevels, progressions: progressions, levels: levels,
	}
}

func (uc *StudentUseCase) Register(ctx context.Context, stateID string, req dto.CreateStudentRequest, createdBy string) (*domain.Student, error) {
	v := validator.New().
		Required(req.AdmissionNo, "admission_no").
		Required(req.FirstName, "first_name").
		Required(req.LastName, "last_name").
		OneOf(req.Gender, []string{"MALE", "FEMALE", "OTHER"}, "gender").
		Check(!req.DateOfBirth.IsZero(), "date_of_birth", "is required").
		Required(req.GuardianName, "guardian_name").
		Required(req.GuardianPhone, "guardian_phone")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	s := &domain.Student{
		StateID: stateID, AdmissionNo: req.AdmissionNo,
		FirstName: req.FirstName, MiddleName: req.MiddleName, LastName: req.LastName,
		Gender: domain.Gender(req.Gender), DateOfBirth: req.DateOfBirth,
		StateOfOrigin: req.StateOfOrigin, LGAID: req.LGAID, Religion: req.Religion,
		Phone: req.Phone, Email: req.Email, Address: req.Address,
		GuardianName: req.GuardianName, GuardianPhone: req.GuardianPhone,
		GuardianRelation: req.GuardianRelation, Status: domain.StudentStatusActive,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return s, uc.students.Create(ctx, s)
}

func (uc *StudentUseCase) GetByID(ctx context.Context, id string) (*domain.Student, error) {
	return uc.students.GetByID(ctx, id)
}

func (uc *StudentUseCase) List(ctx context.Context, f domain.StudentFilter, p pagination.Params) ([]*domain.Student, int, error) {
	return uc.students.List(ctx, f, p)
}

func (uc *StudentUseCase) Update(ctx context.Context, id string, req dto.UpdateStudentRequest, updatedBy string) (*domain.Student, error) {
	s, err := uc.students.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.FirstName = req.FirstName
	s.MiddleName = req.MiddleName
	s.LastName = req.LastName
	s.Gender = domain.Gender(req.Gender)
	if !req.DateOfBirth.IsZero() {
		s.DateOfBirth = req.DateOfBirth
	}
	s.StateOfOrigin = req.StateOfOrigin
	s.LGAID = req.LGAID
	s.Religion = req.Religion
	s.Phone = req.Phone
	s.Email = req.Email
	s.Address = req.Address
	s.GuardianName = req.GuardianName
	s.GuardianPhone = req.GuardianPhone
	s.GuardianRelation = req.GuardianRelation
	if req.Status != "" {
		s.Status = domain.StudentStatus(req.Status)
	}
	s.UpdatedBy = updatedBy
	return s, uc.students.Update(ctx, s)
}

func (uc *StudentUseCase) Delete(ctx context.Context, id string) error {
	return uc.students.Delete(ctx, id)
}

// Enroll registers a student into a school/session/level.
func (uc *StudentUseCase) Enroll(ctx context.Context, req dto.EnrollStudentRequest, createdBy string) (*domain.Enrollment, error) {
	v := validator.New().
		Required(req.StudentID, "student_id").
		Required(req.SchoolID, "school_id").
		Required(req.SessionID, "session_id").
		Required(req.LevelID, "level_id").
		Required(req.SubLevelID, "sub_level_id")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	// Guard: student must exist
	if _, err := uc.students.GetByID(ctx, req.StudentID); err != nil {
		return nil, err
	}

	// Guard: prevent double-enrollment in same session
	exists, err := uc.enrollments.ExistsForSession(ctx, req.StudentID, req.SessionID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict("student already enrolled for this session")
	}

	// Guard: sub-level capacity check
	sl, err := uc.subLevels.GetByID(ctx, req.SubLevelID)
	if err != nil {
		return nil, err
	}
	enrolled, err := uc.subLevels.CountEnrolled(ctx, req.SubLevelID, req.SessionID)
	if err != nil {
		return nil, err
	}
	if enrolled >= sl.Capacity {
		return nil, apperror.Conflict("class arm is at full capacity")
	}

	e := &domain.Enrollment{
		StudentID: req.StudentID, SchoolID: req.SchoolID, SessionID: req.SessionID,
		LevelID: req.LevelID, SubLevelID: req.SubLevelID,
		Status: domain.EnrollmentStatusActive,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return e, uc.enrollments.Create(ctx, e)
}

func (uc *StudentUseCase) GetEnrollment(ctx context.Context, id string) (*domain.Enrollment, error) {
	return uc.enrollments.GetByID(ctx, id)
}

func (uc *StudentUseCase) ListEnrollments(ctx context.Context, f domain.EnrollmentFilter, p pagination.Params) ([]*domain.Enrollment, int, error) {
	return uc.enrollments.List(ctx, f, p)
}

func (uc *StudentUseCase) UpdateEnrollment(ctx context.Context, id string, req dto.UpdateEnrollmentRequest, updatedBy string) (*domain.Enrollment, error) {
	e, err := uc.enrollments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.SubLevelID != "" {
		e.SubLevelID = req.SubLevelID
	}
	if req.Status != "" {
		e.Status = domain.EnrollmentStatus(req.Status)
	}
	e.UpdatedBy = updatedBy
	return e, uc.enrollments.Update(ctx, e)
}

// RecordProgression records end-of-session promotion/repeat/graduation.
func (uc *StudentUseCase) RecordProgression(ctx context.Context, schoolID string, req dto.RecordProgressionRequest, createdBy string) (*domain.LevelProgression, error) {
	v := validator.New().
		Required(req.StudentID, "student_id").
		Required(req.FromSessionID, "from_session_id").
		Required(req.FromLevelID, "from_level_id").
		OneOf(req.Decision, []string{"PROMOTED", "REPEATED", "GRADUATED"}, "decision").
		Check(!req.DecisionDate.IsZero(), "decision_date", "is required")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	// Auto-derive next level when promoted
	toLevel := req.ToLevelID
	if req.Decision == string(domain.ProgressionPromoted) && toLevel == "" {
		next, err := uc.levels.GetNextLevel(ctx, req.FromLevelID)
		if err == nil {
			toLevel = next.ID
		}
	}

	lp := &domain.LevelProgression{
		StudentID: req.StudentID, SchoolID: schoolID,
		FromSessionID: req.FromSessionID, ToSessionID: req.ToSessionID,
		FromLevelID: req.FromLevelID, ToLevelID: toLevel,
		Decision: domain.ProgressionDecision(req.Decision),
		DecidedBy: createdBy, DecisionDate: req.DecisionDate,
		Remarks: req.Remarks,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}

	// Update student status for graduates
	if req.Decision == string(domain.ProgressionGraduated) {
		s, err := uc.students.GetByID(ctx, req.StudentID)
		if err == nil {
			s.Status = domain.StudentStatusGraduated
			s.UpdatedBy = createdBy
			_ = uc.students.Update(ctx, s)
		}
	}

	return lp, uc.progressions.Create(ctx, lp)
}

func (uc *StudentUseCase) ListProgressions(ctx context.Context, studentID string) ([]*domain.LevelProgression, error) {
	return uc.progressions.ListByStudent(ctx, studentID)
}

func (uc *StudentUseCase) ListSessionProgressions(ctx context.Context, schoolID, sessionID string) ([]*domain.LevelProgression, error) {
	return uc.progressions.ListBySession(ctx, schoolID, sessionID)
}

// unused time import guard
var _ = time.Now
