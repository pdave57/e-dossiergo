package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/edossier/api/pkg/validator"
)

// ─────────────────────────────────────────────────────────────────────────────
// STATES USE CASES
// ─────────────────────────────────────────────────────────────────────────────

type StateService struct {
	states domain.StateRepository
}

func NewStateService(states domain.StateRepository) *StateService {
	return &StateService{states: states}
}

func (uc *StateService) CreateState(ctx context.Context, req dto.CreateStateRequest, createdBy string) (*domain.State, error) {
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

func (uc *StateService) GetState(ctx context.Context, id string) (*domain.State, error) {
	return uc.states.GetByID(ctx, id)
}

func (uc *StateService) ListStates(ctx context.Context) ([]*domain.State, error) {
	return uc.states.List(ctx)
}

func (uc *StateService) UpdateState(ctx context.Context, id string, req dto.UpdateStateRequest, updatedBy string) (*domain.State, error) {
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
type ZoneService struct {
	zones domain.ZoneRepository
}

func NewZoneService(zones domain.ZoneRepository) *ZoneService {
	return &ZoneService{zones: zones}
}

func (uc *ZoneService) CreateZone(ctx context.Context, stateID string, req dto.CreateZoneRequest, createdBy string) (*domain.Zone, error) {
	v := validator.New().Required(req.Name, "name").Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	z := &domain.Zone{StateID: stateID, Name: req.Name, Code: req.Code, AuditFields: domain.AuditFields{CreatedBy: createdBy}}
	return z, uc.zones.Create(ctx, z)
}

func (uc *ZoneService) ListZones(ctx context.Context, stateID string) ([]*domain.Zone, error) {
	return uc.zones.ListByState(ctx, stateID)
}

func (uc *ZoneService) UpdateZone(ctx context.Context, id string, req dto.UpdateZoneRequest, updatedBy string) (*domain.Zone, error) {
	z, err := uc.zones.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	z.Name = req.Name
	z.Code = req.Code
	z.UpdatedBy = updatedBy
	return z, uc.zones.Update(ctx, z)
}

func (uc *ZoneService) DeleteZone(ctx context.Context, id string) error {
	return uc.zones.Delete(ctx, id)
}

// — LGAs —
type LGAService struct {
	lgas domain.LGARepository
}

func NewLGAService(lgas domain.LGARepository) *LGAService {
	return &LGAService{lgas: lgas}
}

func (uc *LGAService) CreateLGA(ctx context.Context, stateID string, req dto.CreateLGARequest, createdBy string) (*domain.LGA, error) {
	v := validator.New().Required(req.ZoneID, "zone_id").Required(req.Name, "name").Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	l := &domain.LGA{StateID: stateID, ZoneID: req.ZoneID, Name: req.Name, Code: req.Code, AuditFields: domain.AuditFields{CreatedBy: createdBy}}
	return l, uc.lgas.Create(ctx, l)
}

func (uc *LGAService) ListLGAs(ctx context.Context, stateID string) ([]*domain.LGA, error) {
	return uc.lgas.ListByState(ctx, stateID)
}

func (uc *LGAService) ListLGAsByZone(ctx context.Context, zoneID string) ([]*domain.LGA, error) {
	return uc.lgas.ListByZone(ctx, zoneID)
}

func (uc *LGAService) UpdateLGA(ctx context.Context, id string, req dto.UpdateLGARequest, updatedBy string) (*domain.LGA, error) {
	v := validator.New().
		Required(req.StateID, "state_id").
		Required(req.ZoneID, "zone_id").
		Required(req.Name, "name").
		Required(req.Code, "code")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	l, err := uc.lgas.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	l.StateID = req.StateID
	l.ZoneID = req.ZoneID
	l.Name = req.Name
	l.Code = req.Code
	l.UpdatedBy = updatedBy
	return l, uc.lgas.Update(ctx, l)
}

func (uc *LGAService) DeleteLGA(ctx context.Context, id string) error {
	return uc.lgas.Delete(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type SchoolService struct {
	schools    domain.SchoolRepository
	facilities domain.SchoolFacilityRepository
	storage    domain.ImageStorage
}

func NewSchoolService(schools domain.SchoolRepository, facilities domain.SchoolFacilityRepository, storage domain.ImageStorage) *SchoolService {
	return &SchoolService{schools: schools, facilities: facilities, storage: storage}
}

func (uc *SchoolService) Create(ctx context.Context, stateID string, req dto.CreateSchoolRequest, createdBy string) (*domain.School, error) {
	v := validator.New().
		Required(req.ZoneID, "zone_id").
		Required(req.LGAID, "lga_id").
		Required(req.Name, "name").
		Required(req.Code, "code").
		OneOf(string(req.Category), []string{"NURSERY", "PRIMARY", "JUNIOR_SECONDARY", "SENIOR_SECONDARY", "COMBINED", "VOCATIONAL"}, "category").
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
		NumberOfClassrooms: req.NumberOfClassrooms, TotalStudents: req.TotalStudents,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return s, uc.schools.Create(ctx, s)
}

func (uc *SchoolService) GetByID(ctx context.Context, id string) (*domain.School, error) {
	return uc.schools.GetByID(ctx, id)
}

func (uc *SchoolService) List(ctx context.Context, f domain.SchoolFilter, p pagination.Params) ([]*domain.School, int, error) {
	return uc.schools.List(ctx, f, p)
}

func (uc *SchoolService) Update(ctx context.Context, id string, req dto.UpdateSchoolRequest, updatedBy string) (*domain.School, error) {
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
	s.NumberOfClassrooms = req.NumberOfClassrooms
	s.TotalStudents = req.TotalStudents
	if req.Status != "" {
		s.Status = domain.SchoolStatus(req.Status)
	}
	s.UpdatedBy = updatedBy
	return s, uc.schools.Update(ctx, s)
}

func (uc *SchoolService) Delete(ctx context.Context, id string) error {
	return uc.schools.Delete(ctx, id)
}

// — Facilities —

func (uc *SchoolService) AddFacility(ctx context.Context, schoolID string, req dto.CreateFacilityRequest, createdBy string) (*domain.SchoolFacility, error) {
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

func (uc *SchoolService) ListFacilities(ctx context.Context, schoolID string) ([]*domain.SchoolFacility, error) {
	return uc.facilities.ListBySchool(ctx, schoolID)
}

func (uc *SchoolService) UpdateFacility(ctx context.Context, id string, req dto.UpdateFacilityRequest, updatedBy string) (*domain.SchoolFacility, error) {
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

func (uc *SchoolService) DeleteFacility(ctx context.Context, id string) error {
	return uc.facilities.Delete(ctx, id)
}

func (uc *SchoolService) CountTotalSchools(ctx context.Context, stateID string) (int, error) {
	return uc.schools.CountTotalSchools(ctx, stateID)
}

func (uc *SchoolService) UploadLogo(ctx context.Context, schoolID string, file io.Reader, filename string) (string, error) {
	url, publicID, err := uc.storage.Upload(ctx, file, filename, "school_logos")
	if err != nil {
		return "", err
	}

	err = uc.schools.UpdateLogo(ctx, schoolID, url)
	if err != nil {
		uc.storage.Delete(ctx, publicID)
		return "", err
	}

	return url, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC SESSION USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type AcademicService struct {
	sessions domain.AcademicSessionRepository
	terms    domain.TermRepository
}

func NewAcademicService(sessions domain.AcademicSessionRepository, terms domain.TermRepository) *AcademicService {
	return &AcademicService{sessions: sessions, terms: terms}
}

func (uc *AcademicService) CreateSession(ctx context.Context, schoolID string, req dto.CreateSessionRequest, createdBy string) (*domain.AcademicSession, error) {
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
		SchoolID: schoolID, Name: req.Name,
		StartYear: req.StartYear, EndYear: req.EndYear,
		Status: domain.SessionDraft, StartDate: req.StartDate, EndDate: req.EndDate,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return s, uc.sessions.Create(ctx, s)
}

func (uc *AcademicService) GetSession(ctx context.Context, id string) (*domain.AcademicSession, error) {
	return uc.sessions.GetByID(ctx, id)
}

func (uc *AcademicService) GetActiveSession(ctx context.Context, schoolID string) (*domain.AcademicSession, error) {
	return uc.sessions.GetActive(ctx, schoolID)
}

func (uc *AcademicService) ListSessions(ctx context.Context, schoolID string, p pagination.Params) ([]*domain.AcademicSession, int, error) {
	return uc.sessions.List(ctx, schoolID, p)
}

func (uc *AcademicService) UpdateSession(ctx context.Context, id string, req dto.UpdateSessionRequest, updatedBy string) (*domain.AcademicSession, error) {
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

func (uc *AcademicService) ActivateSession(ctx context.Context, id, schoolID string) error {
	return uc.sessions.SetActive(ctx, id, schoolID)
}

func (uc *AcademicService) DeleteSession(ctx context.Context, id string) error {
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

func (uc *AcademicService) CreateTerm(ctx context.Context, sessionID string, req dto.CreateTermRequest, createdBy string) (*domain.Term, error) {
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

func (uc *AcademicService) ListTerms(ctx context.Context, sessionID string) ([]*domain.Term, error) {
	return uc.terms.ListBySession(ctx, sessionID)
}

func (uc *AcademicService) UpdateTerm(ctx context.Context, id string, req dto.UpdateTermRequest, updatedBy string) (*domain.Term, error) {
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

func (uc *AcademicService) ActivateTerm(ctx context.Context, id, sessionID string) error {
	return uc.terms.SetActive(ctx, id, sessionID)
}

func (uc *AcademicService) DeleteTerm(ctx context.Context, id string) error {
	return uc.terms.Delete(ctx, id)
}

func (uc *AcademicService) GetTerm(ctx context.Context, id string) (*domain.Term, error) {
	return uc.terms.GetByID(ctx, id)
}

func (uc *AcademicService) ListAllTerms(ctx context.Context) ([]*domain.Term, error) {
	return uc.terms.ListAll(ctx)
}

func (uc *AcademicService) GetActiveTerm(ctx context.Context, sessionID string) (*domain.Term, error) {
	return uc.terms.GetActiveTerm(ctx, sessionID)
}

// CreateTermTopLevel creates a term via the top-level /terms endpoint, where the
// owning session is supplied in the request body rather than the URL path.
func (uc *AcademicService) CreateTermTopLevel(ctx context.Context, req dto.CreateTermRequest, createdBy string) (*domain.Term, error) {
	if strings.TrimSpace(req.SessionID) == "" {
		return nil, apperror.BadRequest("session_id is required")
	}
	return uc.CreateTerm(ctx, req.SessionID, req, createdBy)
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type LevelService struct {
	levels       domain.LevelRepository
	subLevels    domain.SubLevelRepository
	schoolLevels domain.SchoolLevelRepository
}

func NewLevelService(
	levels domain.LevelRepository,
	subLevels domain.SubLevelRepository,
	schoolLevels domain.SchoolLevelRepository,
) *LevelService {
	return &LevelService{levels: levels, subLevels: subLevels, schoolLevels: schoolLevels}
}

func (uc *LevelService) CreateLevel(ctx context.Context, schoolID string, req dto.CreateLevelRequest, createdBy string) (*domain.Level, error) {
	v := validator.New().
		Required(req.Name, "name").
		Required(req.Code, "code").
		OneOf(req.Type, []string{"NURSERY", "PRIMARY", "JSS", "SSS", "VOCATIONAL"}, "type")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	l := &domain.Level{
		SchoolID: schoolID, Name: req.Name, Code: req.Code,
		Type: domain.LevelType(req.Type), Order: req.Order,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return l, uc.levels.Create(ctx, l)
}

func (uc *LevelService) ListLevels(ctx context.Context, schoolID string) ([]*domain.Level, error) {
	return uc.levels.ListBySchool(ctx, schoolID)
}

func (uc *LevelService) GetLevel(ctx context.Context, id string) (*domain.Level, error) {
	return uc.levels.GetByID(ctx, id)
}

func (uc *LevelService) UpdateLevel(ctx context.Context, id string, req dto.UpdateLevelRequest, updatedBy string) (*domain.Level, error) {
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

func (uc *LevelService) DeleteLevel(ctx context.Context, id string) error {
	return uc.levels.Delete(ctx, id)
}

// — Sub-Levels —

func (uc *LevelService) CreateSubLevel(ctx context.Context, schoolID string, req dto.CreateSubLevelRequest, createdBy string) (*domain.SubLevel, error) {
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

func (uc *LevelService) ListSubLevels(ctx context.Context, schoolID, levelID string) ([]*domain.SubLevel, error) {
	return uc.subLevels.ListByLevel(ctx, schoolID, levelID)
}

func (uc *LevelService) UpdateSubLevel(ctx context.Context, id string, req dto.UpdateSubLevelRequest, updatedBy string) (*domain.SubLevel, error) {
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

func (uc *LevelService) DeleteSubLevel(ctx context.Context, id string) error {
	return uc.subLevels.Delete(ctx, id)
}

// — School Levels (which levels a school offers) —

func (uc *LevelService) UpsertSchoolLevel(ctx context.Context, schoolID string, req dto.UpsertSchoolLevelRequest, createdBy string) error {
	return uc.schoolLevels.Upsert(ctx, &domain.SchoolLevel{
		SchoolID: schoolID, LevelID: req.LevelID, SessionID: req.SessionID,
		IsActive: req.IsActive, AuditFields: domain.AuditFields{CreatedBy: createdBy},
	})
}

func (uc *LevelService) ListSchoolLevels(ctx context.Context, schoolID, sessionID string) ([]*domain.SchoolLevel, error) {
	return uc.schoolLevels.ListBySchool(ctx, schoolID, sessionID)
}

func (uc *LevelService) RemoveSchoolLevel(ctx context.Context, schoolID, levelID, sessionID string) error {
	return uc.schoolLevels.Delete(ctx, schoolID, levelID, sessionID)
}

// ─────────────────────────────────────────────────────────────────────────────
// SUBJECT USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type SubjectService struct {
	subjects       domain.SubjectRepository
	schoolSubjects domain.SchoolSubjectRepository
}

func NewSubjectService(subjects domain.SubjectRepository, schoolSubjects domain.SchoolSubjectRepository) *SubjectService {
	return &SubjectService{subjects: subjects, schoolSubjects: schoolSubjects}
}

func (uc *SubjectService) Create(ctx context.Context, stateID string, req dto.CreateSubjectRequest, createdBy string) (*domain.Subject, error) {
	v := validator.New().
		Required(req.Name, "name").
		Required(req.Code, "code").
		OneOf(req.Category, []string{"CORE", "ELECTIVE", "PRACTICAL"}, "category").
		OneOf(req.LevelType, []string{"NURSERY", "PRIMARY", "JSS", "SSS", "VOCATIONAL"}, "level_type")
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

func (uc *SubjectService) GetByID(ctx context.Context, id string) (*domain.Subject, error) {
	return uc.subjects.GetByID(ctx, id)
}

func (uc *SubjectService) ListByState(ctx context.Context, stateID string) ([]*domain.Subject, error) {
	return uc.subjects.ListByState(ctx, stateID)
}

func (uc *SubjectService) Update(ctx context.Context, id string, req dto.UpdateSubjectRequest, updatedBy string) (*domain.Subject, error) {
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

func (uc *SubjectService) Delete(ctx context.Context, id string) error {
	return uc.subjects.Delete(ctx, id)
}

// — School Subjects —

func (uc *SubjectService) AssignToSchool(ctx context.Context, schoolID string, req dto.CreateSchoolSubjectRequest, createdBy string) (*domain.SchoolSubject, error) {
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

func (uc *SubjectService) ListSchoolSubjects(ctx context.Context, schoolID, sessionID, levelID string) ([]*domain.SchoolSubject, error) {
	return uc.schoolSubjects.ListBySchool(ctx, schoolID, sessionID, levelID)
}

func (uc *SubjectService) UpdateSchoolSubject(ctx context.Context, id string, req dto.UpdateSchoolSubjectRequest, updatedBy string) (*domain.SchoolSubject, error) {
	ss, err := uc.schoolSubjects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ss.TeacherID = req.TeacherID
	ss.IsActive = req.IsActive
	ss.UpdatedBy = updatedBy
	return ss, uc.schoolSubjects.Update(ctx, ss)
}

func (uc *SubjectService) RemoveSchoolSubject(ctx context.Context, id string) error {
	return uc.schoolSubjects.Delete(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type PersonnelService struct {
	personnel domain.PersonnelRepository
	transfers domain.PersonnelTransferRepository
	schools   domain.SchoolRepository
}

func NewPersonnelService(
	personnel domain.PersonnelRepository,
	transfers domain.PersonnelTransferRepository,
	schools domain.SchoolRepository,
) *PersonnelService {
	return &PersonnelService{personnel: personnel, transfers: transfers, schools: schools}
}

func (uc *PersonnelService) Create(ctx context.Context, stateID string, req dto.CreatePersonnelRequest, createdBy string) (*domain.Personnel, error) {
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

func (uc *PersonnelService) GetByID(ctx context.Context, id string) (*domain.Personnel, error) {
	return uc.personnel.GetByID(ctx, id)
}

func (uc *PersonnelService) List(ctx context.Context, f domain.PersonnelFilter, p pagination.Params) ([]*domain.Personnel, int, error) {
	return uc.personnel.List(ctx, f, p)
}

func (uc *PersonnelService) Update(ctx context.Context, id string, req dto.UpdatePersonnelRequest, updatedBy string) (*domain.Personnel, error) {
	p, err := uc.personnel.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.SchoolID != "" {
		p.SchoolID = req.SchoolID
	}
	p.StaffID = req.StaffID
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

func (uc *PersonnelService) Delete(ctx context.Context, id string) error {
	return uc.personnel.Delete(ctx, id)
}

// Transfer moves a staff member to another school and updates their school_id.
func (uc *PersonnelService) Transfer(ctx context.Context, personnelID string, req dto.TransferPersonnelRequest, createdBy string) (*domain.PersonnelTransfer, error) {
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

func (uc *PersonnelService) ListTransfers(ctx context.Context, personnelID string) ([]*domain.PersonnelTransfer, error) {
	return uc.transfers.ListByPersonnel(ctx, personnelID)
}

func (uc *PersonnelService) ListSchoolTransfers(ctx context.Context, schoolID string) ([]*domain.PersonnelTransfer, error) {
	return uc.transfers.ListBySchool(ctx, schoolID)
}

func (uc *PersonnelService) CountTotalPersonnel(ctx context.Context, stateID string) (int, error) {
	return uc.personnel.CountTotalPersonnel(ctx, stateID)
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT USE CASE
// ─────────────────────────────────────────────────────────────────────────────

type StudentService struct {
	students     domain.StudentRepository
	enrollments  domain.EnrollmentRepository
	subLevels    domain.SubLevelRepository
	progressions domain.LevelProgressionRepository
	levels       domain.LevelRepository
	schools      domain.SchoolRepository
	states       domain.StateRepository
	lgas         domain.LGARepository
	scores       domain.ScoreSheetRepository
	terms        domain.TermRepository
}

func NewStudentService(
	students domain.StudentRepository,
	enrollments domain.EnrollmentRepository,
	subLevels domain.SubLevelRepository,
	progressions domain.LevelProgressionRepository,
	levels domain.LevelRepository,
	schools domain.SchoolRepository,
	states domain.StateRepository,
	lgas domain.LGARepository,
	scores domain.ScoreSheetRepository,
	terms domain.TermRepository,
) *StudentService {
	return &StudentService{
		students: students, enrollments: enrollments,
		subLevels: subLevels, progressions: progressions, levels: levels,
		schools: schools, states: states, lgas: lgas,
		scores: scores, terms: terms,
	}
}

func (uc *StudentService) Register(ctx context.Context, stateID string, req dto.CreateStudentRequest, createdBy string) (*domain.Student, error) {
	v := validator.New().
		Required(req.SchoolID, "school_id").
		Required(req.FirstName, "first_name").
		Required(req.LastName, "last_name").
		OneOf(req.Gender, []string{"MALE", "FEMALE", "OTHER"}, "gender").
		Check(!req.DateOfBirth.IsZero(), "date_of_birth", "is required").
		Required(req.GuardianName, "guardian_name").
		Required(req.GuardianPhone, "guardian_phone")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	// Resolve school code used to build the enrollment number.
	school, err := uc.schools.GetByID(ctx, req.SchoolID)
	if err != nil {
		return nil, apperror.NotFound("school", req.SchoolID)
	}

	// Use provided enrollment year or default to current year.
	year := req.EnrollmentYear
	if year == 0 {
		year = time.Now().Year()
	}

	// Enrollment number format: SCHOOLCODE-YY-SERIAL
	// e.g. GSSJAL-26-0001, where SERIAL is incremented per (school, year).
	yy := year % 100
	prefix := fmt.Sprintf("%s-%02d-", school.Code, yy)
	serialNo, err := uc.students.GetNextSerialByPrefix(ctx, prefix)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	enrollmentNo := fmt.Sprintf("%s%04d", prefix, serialNo)

	s := &domain.Student{
		StateID: stateID, EnrollmentYear: year, EnrollmentNo: enrollmentNo, SerialNo: serialNo,
		FirstName: req.FirstName, MiddleName: req.MiddleName, LastName: req.LastName,
		Gender: domain.Gender(req.Gender), DateOfBirth: req.DateOfBirth,
		StateOfOrigin: req.StateOfOrigin, LGAID: req.LGAID, Religion: req.Religion,
		Address:      req.Address,
		GuardianName: req.GuardianName, GuardianPhone: req.GuardianPhone,
		GuardianRelation: req.GuardianRelation, Status: domain.StudentStatusActive,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return s, uc.students.Create(ctx, s)
}

func (uc *StudentService) GetNextSerial(ctx context.Context, schoolID string, year int) (int, error) {
	school, err := uc.schools.GetByID(ctx, schoolID)
	if err != nil {
		return 0, apperror.NotFound("school", schoolID)
	}
	if year == 0 {
		year = time.Now().Year()
	}
	prefix := fmt.Sprintf("%s-%02d-", school.Code, year%100)
	return uc.students.GetNextSerialByPrefix(ctx, prefix)
}

func (uc *StudentService) GetByID(ctx context.Context, id string) (*domain.Student, error) {
	return uc.students.GetByID(ctx, id)
}

func (uc *StudentService) List(ctx context.Context, f domain.StudentFilter, p pagination.Params) ([]*domain.Student, int, error) {
	return uc.students.List(ctx, f, p)
}

func (uc *StudentService) Update(ctx context.Context, id string, req dto.UpdateStudentRequest, updatedBy string) (*domain.Student, error) {
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

func (uc *StudentService) Delete(ctx context.Context, id string) error {
	return uc.students.Delete(ctx, id)
}

// Enroll registers a student into a school/session/level.
func (uc *StudentService) Enroll(ctx context.Context, req dto.EnrollStudentRequest, createdBy string) (*domain.Enrollment, error) {
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
		Status:      domain.EnrollmentStatusActive,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return e, uc.enrollments.Create(ctx, e)
}

func (uc *StudentService) GetEnrollment(ctx context.Context, id string) (*domain.Enrollment, error) {
	return uc.enrollments.GetByID(ctx, id)
}

func (uc *StudentService) ListEnrollments(ctx context.Context, f domain.EnrollmentFilter, p pagination.Params) ([]*domain.Enrollment, int, error) {
	return uc.enrollments.List(ctx, f, p)
}

func (uc *StudentService) UpdateEnrollment(ctx context.Context, id string, req dto.UpdateEnrollmentRequest, updatedBy string) (*domain.Enrollment, error) {
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
func (uc *StudentService) RecordProgression(ctx context.Context, schoolID string, req dto.RecordProgressionRequest, createdBy string) (*domain.LevelProgression, error) {
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
		Decision:  domain.ProgressionDecision(req.Decision),
		DecidedBy: createdBy, DecisionDate: req.DecisionDate,
		Remarks:     req.Remarks,
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

func (uc *StudentService) ListProgressions(ctx context.Context, studentID string) ([]*domain.LevelProgression, error) {
	return uc.progressions.ListByStudent(ctx, studentID)
}

func (uc *StudentService) ListSessionProgressions(ctx context.Context, schoolID, sessionID string) ([]*domain.LevelProgression, error) {
	return uc.progressions.ListBySession(ctx, schoolID, sessionID)
}

func (uc *StudentService) CountTotalStudents(ctx context.Context, stateID string) (int, error) {
	return uc.students.CountTotalStudents(ctx, stateID)
}

// unused time import guard
var _ = time.Now
