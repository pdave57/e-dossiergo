// Package domain — repository interfaces.
// These are the ports in the hexagonal / clean architecture sense.
// Infrastructure layer must implement all of these.
package domain

import (
	"context"
	"time"

	"github.com/edossier/api/pkg/pagination"
)

// ─────────────────────────────────────────────────────────────────────────────
// FILTER STRUCTS  (passed to list methods)
// ─────────────────────────────────────────────────────────────────────────────

type UserFilter struct {
	StateID  string
	SchoolID string
	Status   string
	Search   string
}

type SchoolFilter struct {
	StateID   string
	ZoneID    string
	LGAID     string
	Category  string
	Ownership string
	Status    string
	Search    string
}

type PersonnelFilter struct {
	StateID  string
	SchoolID string
	Role     string
	Status   string
	Search   string
}

type StudentFilter struct {
	StateID  string
	SchoolID string
	LGAID    string
	Status   string
	Search   string
}

type EnrollmentFilter struct {
	SchoolID   string
	SessionID  string
	LevelID    string
	SubLevelID string
	Status     string
}

type ScoreSheetFilter struct {
	SchoolID   string
	SessionID  string
	TermID     string
	LevelID    string
	SubLevelID string
	SubjectID  string
}

// ─────────────────────────────────────────────────────────────────────────────
// AUTH / USER REPOSITORIES
// ─────────────────────────────────────────────────────────────────────────────

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter UserFilter, p pagination.Params) ([]*User, int, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id string) (*Role, error)
	GetByCode(ctx context.Context, code string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, stateID string) ([]*Role, error)

	// Permissions
	AddPermission(ctx context.Context, rp *RolePermission) error
	RemovePermission(ctx context.Context, roleID, permissionID string) error
	GetPermissions(ctx context.Context, roleID string) ([]*Permission, error)
	GetPermissionsForUser(ctx context.Context, userID string) ([]*Permission, error)
}

type PermissionRepository interface {
	Create(ctx context.Context, perm *Permission) error
	GetByID(ctx context.Context, id string) (*Permission, error)
	List(ctx context.Context) ([]*Permission, error)
	Delete(ctx context.Context, id string) error
}

type UserRoleRepository interface {
	Assign(ctx context.Context, ur *UserRole) error
	Revoke(ctx context.Context, userID, roleID string) error
	GetRolesForUser(ctx context.Context, userID string) ([]*Role, error)
	HasRole(ctx context.Context, userID, roleCode string) (bool, error)
	HasPermission(ctx context.Context, userID, resource, action string) (bool, error)
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, rt *RefreshToken) error
	GetByUserID(ctx context.Context, userID string) (*RefreshToken, error)
	Revoke(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

// ─────────────────────────────────────────────────────────────────────────────
// GEO
// ─────────────────────────────────────────────────────────────────────────────

type StateRepository interface {
	Create(ctx context.Context, s *State) error
	GetByID(ctx context.Context, id string) (*State, error)
	GetByCode(ctx context.Context, code string) (*State, error)
	Update(ctx context.Context, s *State) error
	List(ctx context.Context) ([]*State, error)
}

type ZoneRepository interface {
	Create(ctx context.Context, z *Zone) error
	GetByID(ctx context.Context, id string) (*Zone, error)
	Update(ctx context.Context, z *Zone) error
	Delete(ctx context.Context, id string) error
	ListByState(ctx context.Context, stateID string) ([]*Zone, error)
}

type LGARepository interface {
	Create(ctx context.Context, l *LGA) error
	GetByID(ctx context.Context, id string) (*LGA, error)
	Update(ctx context.Context, l *LGA) error
	Delete(ctx context.Context, id string) error
	ListByState(ctx context.Context, stateID string) ([]*LGA, error)
	ListByZone(ctx context.Context, zoneID string) ([]*LGA, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL
// ─────────────────────────────────────────────────────────────────────────────

type SchoolRepository interface {
	Create(ctx context.Context, s *School) error
	GetByID(ctx context.Context, id string) (*School, error)
	GetByCode(ctx context.Context, code string) (*School, error)
	Update(ctx context.Context, s *School) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter SchoolFilter, p pagination.Params) ([]*School, int, error)
	CountTotalSchools(ctx context.Context, stateID string) (int, error)
	UpdateLogo(ctx context.Context, id, logoURL string) error
}

type SchoolFacilityRepository interface {
	Create(ctx context.Context, f *SchoolFacility) error
	GetByID(ctx context.Context, id string) (*SchoolFacility, error)
	Update(ctx context.Context, f *SchoolFacility) error
	Delete(ctx context.Context, id string) error
	ListBySchool(ctx context.Context, schoolID string) ([]*SchoolFacility, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC SETUP
// ─────────────────────────────────────────────────────────────────────────────

type AcademicSessionRepository interface {
	Create(ctx context.Context, s *AcademicSession) error
	GetByID(ctx context.Context, id string) (*AcademicSession, error)
	GetActive(ctx context.Context, schoolID string) (*AcademicSession, error)
	Update(ctx context.Context, s *AcademicSession) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, schoolID string, p pagination.Params) ([]*AcademicSession, int, error)
	SetActive(ctx context.Context, id, schoolID string) error
}

type TermRepository interface {
	Create(ctx context.Context, t *Term) error
	GetByID(ctx context.Context, id string) (*Term, error)
	GetActiveTerm(ctx context.Context, sessionID string) (*Term, error)
	Update(ctx context.Context, t *Term) error
	Delete(ctx context.Context, id string) error
	ListBySession(ctx context.Context, sessionID string) ([]*Term, error)
	ListAll(ctx context.Context) ([]*Term, error)
	SetActive(ctx context.Context, id, sessionID string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL
// ─────────────────────────────────────────────────────────────────────────────

type LevelRepository interface {
	Create(ctx context.Context, l *Level) error
	GetByID(ctx context.Context, id string) (*Level, error)
	Update(ctx context.Context, l *Level) error
	Delete(ctx context.Context, id string) error
	ListBySchool(ctx context.Context, schoolID string) ([]*Level, error)
	GetNextLevel(ctx context.Context, currentLevelID string) (*Level, error)
}

type SubLevelRepository interface {
	Create(ctx context.Context, s *SubLevel) error
	GetByID(ctx context.Context, id string) (*SubLevel, error)
	Update(ctx context.Context, s *SubLevel) error
	Delete(ctx context.Context, id string) error
	ListByLevel(ctx context.Context, schoolID, levelID string) ([]*SubLevel, error)
	CountEnrolled(ctx context.Context, subLevelID, sessionID string) (int, error)
}

type SchoolLevelRepository interface {
	Upsert(ctx context.Context, sl *SchoolLevel) error
	ListBySchool(ctx context.Context, schoolID, sessionID string) ([]*SchoolLevel, error)
	Delete(ctx context.Context, schoolID, levelID, sessionID string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// SUBJECT
// ─────────────────────────────────────────────────────────────────────────────

type SubjectRepository interface {
	Create(ctx context.Context, s *Subject) error
	GetByID(ctx context.Context, id string) (*Subject, error)
	GetByCode(ctx context.Context, code string) (*Subject, error)
	Update(ctx context.Context, s *Subject) error
	Delete(ctx context.Context, id string) error
	ListByState(ctx context.Context, stateID string) ([]*Subject, error)
}

type SchoolSubjectRepository interface {
	Create(ctx context.Context, ss *SchoolSubject) error
	GetByID(ctx context.Context, id string) (*SchoolSubject, error)
	Update(ctx context.Context, ss *SchoolSubject) error
	Delete(ctx context.Context, id string) error
	ListBySchool(ctx context.Context, schoolID, sessionID, levelID string) ([]*SchoolSubject, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL
// ─────────────────────────────────────────────────────────────────────────────

type PersonnelRepository interface {
	Create(ctx context.Context, p *Personnel) error
	GetByID(ctx context.Context, id string) (*Personnel, error)
	GetByStaffID(ctx context.Context, staffID string) (*Personnel, error)
	Update(ctx context.Context, p *Personnel) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter PersonnelFilter, p pagination.Params) ([]*Personnel, int, error)
	CountTotalPersonnel(ctx context.Context, stateID string) (int, error)
	UpdateAvatar(ctx context.Context, id, avatarURL string) error
}

type PersonnelTransferRepository interface {
	Create(ctx context.Context, t *PersonnelTransfer) error
	GetByID(ctx context.Context, id string) (*PersonnelTransfer, error)
	ListByPersonnel(ctx context.Context, personnelID string) ([]*PersonnelTransfer, error)
	ListBySchool(ctx context.Context, schoolID string) ([]*PersonnelTransfer, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT & ENROLLMENT
// ─────────────────────────────────────────────────────────────────────────────

type StudentRepository interface {
	Create(ctx context.Context, s *Student) error
	GetByID(ctx context.Context, id string) (*Student, error)
	GetByEnrollmentNo(ctx context.Context, admNo string) (*Student, error)
	Update(ctx context.Context, s *Student) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter StudentFilter, p pagination.Params) ([]*Student, int, error)
	CountByEnrollmentPrefix(ctx context.Context, prefix string) (int, error)
	CountByGender(ctx context.Context, stateID string) (male, female, other int, err error)
	CountTotalStudents(ctx context.Context, stateID string) (int, error)
	UpdateAvatar(ctx context.Context, id, avatarURL string) error
	GetNextSerialByPrefix(ctx context.Context, prefix string) (int, error)
}

type EnrollmentRepository interface {
	Create(ctx context.Context, e *Enrollment) error
	GetByID(ctx context.Context, id string) (*Enrollment, error)
	GetActiveByStudent(ctx context.Context, studentID, sessionID string) (*Enrollment, error)
	Update(ctx context.Context, e *Enrollment) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter EnrollmentFilter, p pagination.Params) ([]*Enrollment, int, error)
	ExistsForSession(ctx context.Context, studentID, sessionID string) (bool, error)
}

type LevelProgressionRepository interface {
	Create(ctx context.Context, lp *LevelProgression) error
	GetByID(ctx context.Context, id string) (*LevelProgression, error)
	ListByStudent(ctx context.Context, studentID string) ([]*LevelProgression, error)
	ListBySession(ctx context.Context, schoolID, sessionID string) ([]*LevelProgression, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// EXAMINATION & RESULTS
// ─────────────────────────────────────────────────────────────────────────────

type ScoreConfigRepository interface {
	Upsert(ctx context.Context, sc *ScoreConfig) error
	GetBySchool(ctx context.Context, schoolID string) (*ScoreConfig, error)
	GetStateDefault(ctx context.Context, stateID string) (*ScoreConfig, error)
}

type GradeConfigRepository interface {
	Upsert(ctx context.Context, gc *GradeConfig) error
	Delete(ctx context.Context, id string) error
	ListBySchool(ctx context.Context, schoolID string) ([]*GradeConfig, error)
	ListBySchoolAndLevel(ctx context.Context, schoolID, levelID string) ([]*GradeConfig, error)
	ListStateDefault(ctx context.Context, stateID string) ([]*GradeConfig, error)
	EvaluateGrade(ctx context.Context, score float64, schoolID, stateID string) (*GradeConfig, error)
}

type ScoreSheetRepository interface {
	Upsert(ctx context.Context, ss *ScoreSheet) error
	GetByID(ctx context.Context, id string) (*ScoreSheet, error)
	GetByStudentSubjectTerm(ctx context.Context, studentID, subjectID, termID string) (*ScoreSheet, error)
	List(ctx context.Context, filter ScoreSheetFilter, p pagination.Params) ([]*ScoreSheet, int, error)
	ListByStudent(ctx context.Context, studentID, sessionID string) ([]*ScoreSheet, error)
	ComputePositions(ctx context.Context, termID, subLevelID, subjectID string) error
	ComputePositionsBulk(ctx context.Context, termID, subLevelID string) error
	GetClassSubjectStats(ctx context.Context, termID, subLevelID string) ([]ClassSubjectStat, error)
}

type ReportCardRepository interface {
	Upsert(ctx context.Context, rc *ReportCard) error
	GetByID(ctx context.Context, id string) (*ReportCard, error)
	GetByStudentTerm(ctx context.Context, studentID, termID string) (*ReportCard, error)
	ListByTerm(ctx context.Context, schoolID, termID string, p pagination.Params) ([]*ReportCard, int, error)
	ListByStudent(ctx context.Context, studentID string) ([]*ReportCard, error)
	Publish(ctx context.Context, id string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// ATTENDANCE
// ─────────────────────────────────────────────────────────────────────────────

type PersonnelAttendanceRepository interface {
	Create(ctx context.Context, a *PersonnelAttendance) error
	GetByID(ctx context.Context, id string) (*PersonnelAttendance, error)
	GetByPersonnelAndDate(ctx context.Context, personnelID string, date time.Time) (*PersonnelAttendance, error)
	Update(ctx context.Context, a *PersonnelAttendance) error
	Delete(ctx context.Context, id string) error
	ListBySchoolAndDate(ctx context.Context, schoolID string, date time.Time) ([]*PersonnelAttendance, error)
	ListByPersonnelAndRange(ctx context.Context, personnelID string, from, to time.Time) ([]*PersonnelAttendance, error)
}

type StudentAttendanceRepository interface {
	Create(ctx context.Context, a *StudentAttendance) error
	GetByID(ctx context.Context, id string) (*StudentAttendance, error)
	GetByStudentAndDate(ctx context.Context, studentID string, date time.Time) (*StudentAttendance, error)
	Update(ctx context.Context, a *StudentAttendance) error
	Delete(ctx context.Context, id string) error
	ListBySchoolAndDate(ctx context.Context, schoolID string, date time.Time) ([]*StudentAttendance, error)
	ListByStudentAndRange(ctx context.Context, studentID string, from, to time.Time) ([]*StudentAttendance, error)
	ListBySchoolAndRange(ctx context.Context, schoolID string, from, to time.Time) ([]*StudentAttendance, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// REPORTS & ANALYTICS
// ─────────────────────────────────────────────────────────────────────────────

type ReportRepository interface {
	GetDashboardStats(ctx context.Context, stateID, schoolID string) (*DashboardStats, error)
	GetTotalTeachingPersonnel(ctx context.Context) (int, error)
	
}

type ZonalReportRepository interface {
	GetZoneSummaryReport(ctx context.Context, sessionID string, stateID string) ([]ZoneSummaryReport, error)
}


// ─────────────────────────────────────────────────────────────────────────────
// RECOMMENDATION / ML
// ─────────────────────────────────────────────────────────────────────────────

type RecommendationRepository interface {
	ListSchoolsWithAggregates(ctx context.Context) ([]SchoolRecommendationRow, error)
}

// SchoolRecommendationRow is a pre-aggregated projection used by the ML recommender.
type SchoolRecommendationRow struct {
	ID                   string
	Name                 string
	ZoneName             string
	LGAName              string
	Category             string
	TotalTeachers        int
	QualifiedTeachers    int
	TotalStudents        int
	TotalClassrooms      int
	FunctionalClassrooms int
	HasLibrary           bool
	HasLaboratory        bool
	HasToilet            bool
	HasElectricity       bool
	HasWater             bool
	HasInternet          bool
	SubjectsOffered      int
	ExpectedSubjects     int
	BooksPerStudent      float64
	AvgPassRate          *float64
}

// ─────────────────────────────────────────────────────────────────────────────
// REPOSITORY INTERFACE  (port — infrastructure implements this)
// ─────────────────────────────────────────────────────────────────────────────
 
// PredictionRepository is the data-access contract for the prediction engine.
// All methods are read-only aggregate queries — no writes.
type PredictionRepository interface {
	// GetFacilitySignal returns the aggregated facility profile for a school.
	GetFacilitySignal(schoolID string) (*FacilitySignal, error)
 
	// GetPersonnelSignal returns the aggregated staff profile for a school.
	GetPersonnelSignal(schoolID string) (*PersonnelSignal, error)
 
	// GetSchoolHistoricalSignal returns aggregate performance for a school.
	GetSchoolHistoricalSignal(schoolID string) (*HistoricalSignal, error)
 
	// GetStudentHistoricalSignal returns per-student performance signal.
	GetStudentHistoricalSignal(studentID, schoolID string) (*HistoricalSignal, error)
 
	// GetEnrolledStudents returns lightweight student rows for a school+session.
	GetEnrolledStudents(schoolID, sessionID string) ([]StudentRow, error)
 
	// GetSchoolName returns just the school name for labelling.
	GetSchoolName(schoolID string) (string, error)
 
	// GetEnrollmentCount returns the total active enrolled students for ratio calc.
	GetEnrollmentCount(schoolID string) (int, error)
}
 
// StudentRow is a minimal projection used only by the prediction engine.
type StudentRow struct {
	ID         string
	FirstName  string
	LastName   string
}

// GetFullName returns the student's full name as "First Last".
func (sr *StudentRow) GetFullName() string {
	return sr.FirstName + " " + sr.LastName
}