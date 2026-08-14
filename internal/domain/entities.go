// Package domain contains all enterprise business entities and rules.
// It has zero dependencies on any framework, database, or external package.
package domain

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// COMMON EMBEDDED TYPES
// ─────────────────────────────────────────────────────────────────────────────

// AuditFields tracks creation and mutation metadata on every entity.
type AuditFields struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	UpdatedBy string     `json:"updated_by,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// USER / AUTH / RBAC
// ─────────────────────────────────────────────────────────────────────────────

// UserStatus represents the lifecycle state of a user account.
type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusInactive  UserStatus = "INACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

// User is the authentication principal.
type User struct {
	ID           string     `json:"id"`
	StateID      string     `json:"state_id"`
	SchoolID     string     `json:"school_id,omitempty"` // empty = state-level user
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Status       UserStatus `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	AuditFields
}

// FullName returns the user's display name.
func (u *User) FullName() string { return u.FirstName + " " + u.LastName }

// IsActive returns whether the user can authenticate.
func (u *User) IsActive() bool { return u.Status == UserStatusActive }

// Role is a named collection of permissions (RBAC).
type Role struct {
	ID          string `json:"id"`
	StateID     string `json:"state_id"`
	Name        string `json:"name"` // "PRINCIPAL", "TEACHER", "STATE_ADMIN"
	Code        string `json:"code"` // unique slug
	Description string `json:"description,omitempty"`
	IsSystem    bool   `json:"is_system"` // system roles cannot be deleted
	AuditFields
}

// Permission is an atomic capability.
type Permission struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"` // "students", "results", "personnel"
	Action      string `json:"action"`   // "create", "read", "update", "delete", "publish"
	Description string `json:"description,omitempty"`
}

// RolePermission links a role to a permission.
type RolePermission struct {
	RoleID       string    `json:"role_id"`
	PermissionID string    `json:"permission_id"`
	GrantedAt    time.Time `json:"granted_at"`
	GrantedBy    string    `json:"granted_by"`
}

// UserRole links a user to a role, optionally scoped to a school.
type UserRole struct {
	UserID     string    `json:"user_id"`
	RoleID     string    `json:"role_id"`
	SchoolID   string    `json:"school_id,omitempty"`
	AssignedAt time.Time `json:"assigned_at"`
	AssignedBy string    `json:"assigned_by"`
}

// RefreshToken stores hashed refresh tokens for rotation.
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ZONES / ADMINISTRATIVE
// ─────────────────────────────────────────────────────────────────────────────

// State is the top-level administrative scope.
type State struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Country string `json:"country"`
	AuditFields
}

// Zone groups LGAs within a state for educational administration.
type Zone struct {
	ID      string `json:"id"`
	StateID string `json:"state_id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	AuditFields
}

// LGA is a Local Government Area.
type LGA struct {
	ID       string `json:"id"`
	StateID  string `json:"state_id"`
	ZoneID   string `json:"zone_id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	ZoneName string `json:"zone_name"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL
// ─────────────────────────────────────────────────────────────────────────────

type SchoolCategory string

const (
	SchoolCategoryPrimary    SchoolCategory = "PRIMARY"
	SchoolCategoryJuniorSec  SchoolCategory = "JUNIOR_SECONDARY"
	SchoolCategorySeniorSec  SchoolCategory = "SENIOR_SECONDARY"
	SchoolCategoryCombined   SchoolCategory = "COMBINED"
	SchoolCategoryVocational SchoolCategory = "VOCATIONAL"
)

type SchoolOwnership string

const (
	SchoolOwnershipGovernment SchoolOwnership = "GOVERNMENT"
	SchoolOwnershipPrivate    SchoolOwnership = "PRIVATE"
	SchoolOwnershipMission    SchoolOwnership = "MISSION"
	SchoolOwnershipCommunity  SchoolOwnership = "COMMUNITY"
)

type SchoolStatus string

const (
	SchoolStatusActive    SchoolStatus = "ACTIVE"
	SchoolStatusInactive  SchoolStatus = "INACTIVE"
	SchoolStatusSuspended SchoolStatus = "SUSPENDED"
)

// School belongs to an LGA (one LGA contains many Schools).
// Hierarchy: State → Zone → LGA → School
type School struct {
	ID                 string          `json:"id"`
	StateID            string          `json:"state_id"`
	ZoneID             string          `json:"zone_id"`
	LGAID              string          `json:"lga_id"` // owning Local Government Area
	Name               string          `json:"name"`
	Code               string          `json:"code"`
	Category           SchoolCategory  `json:"category"`
	Ownership          SchoolOwnership `json:"ownership"`
	Status             SchoolStatus    `json:"status"`
	NumberOfClassrooms int             `json:"number_of_classrooms"`
	TotalStudents      int             `json:"total_students"`
	Address            string          `json:"address"`
	HeadTeacher        string          `json:"head_teacher,omitempty"`
	Founded            *int            `json:"founded,omitempty"`
	LogoURL            string          `json:"logo_url,omitempty"`
	AuditFields
}

type FacilityType string

const (
	FacilityLibrary    FacilityType = "LIBRARY"
	FacilityLab        FacilityType = "LABORATORY"
	FacilitySportField FacilityType = "SPORT_FIELD"
	FacilityICT        FacilityType = "ICT_CENTER"
	FacilityToilet     FacilityType = "TOILET"
	FacilityBorehole   FacilityType = "BOREHOLE"
	FacilityGenerator  FacilityType = "GENERATOR"
	FacilityCanteen    FacilityType = "CANTEEN"
)

type FacilityCondition string

const (
	FacilityGood    FacilityCondition = "GOOD"
	FacilityFair    FacilityCondition = "FAIR"
	FacilityPoor    FacilityCondition = "POOR"
	FacilityDefunct FacilityCondition = "DEFUNCT"
)

// SchoolFacility records infrastructure available at a school.
type SchoolFacility struct {
	ID        string            `json:"id"`
	SchoolID  string            `json:"school_id"`
	Type      FacilityType      `json:"type"`
	Name      string            `json:"name"`
	Quantity  int               `json:"quantity"`
	Condition FacilityCondition `json:"condition"`
	Notes     string            `json:"notes,omitempty"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC SESSION & TERMS
// ─────────────────────────────────────────────────────────────────────────────

type SessionStatus string

const (
	SessionDraft  SessionStatus = "DRAFT"
	SessionActive SessionStatus = "ACTIVE"
	SessionClosed SessionStatus = "CLOSED"
)

// AcademicSession represents a school year e.g. 2024/2025.
type AcademicSession struct {
	ID        string        `json:"id"`
	SchoolID  string        `json:"school_id"`
	Name      string        `json:"name"`
	StartYear int           `json:"start_year"`
	EndYear   int           `json:"end_year"`
	Status    SessionStatus `json:"status"`
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
	AuditFields
}

// Term represents one of three academic terms within a session.
type Term struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Number    int       `json:"term_number"` // 1 | 2 | 3
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsActive  bool      `json:"is_active"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL (CLASS) SETUP
// ─────────────────────────────────────────────────────────────────────────────

type LevelType string

const (
	LevelTypeNursery    LevelType = "NURSERY"
	LevelTypePrimary    LevelType = "PRIMARY"
	LevelTypeJSS        LevelType = "JSS"
	LevelTypeSSS        LevelType = "SSS"
	LevelTypeVocational LevelType = "VOCATIONAL"
)

// levelTypes holds all valid LevelType constants for enumeration.
var levelTypes = []LevelType{
	LevelTypeNursery,
	LevelTypePrimary,
	LevelTypeJSS,
	LevelTypeSSS,
	LevelTypeVocational,
}

// Values returns all valid LevelType constants, so callers (e.g. the HTTP
// layer) can expose them from the backend instead of hardcoding on the client.
func (t LevelType) Values() []LevelType { return levelTypes }

// Level is a school-defined class (JSS1, JSS2, SS1, …).
// A School owns its Levels; use SchoolLevel to activate a Level for a session.
type Level struct {
	ID       string    `json:"id"`
	SchoolID string    `json:"school_id"` // owning school
	Name     string    `json:"name"`
	Code     string    `json:"code"`
	Type     LevelType `json:"type"`
	Order    int       `json:"order"` // used for progression sequencing
	AuditFields
}

// SubLevel is a school-defined class arm (JSS1A, JSS1B, …).
type SubLevel struct {
	ID       string `json:"id"`
	SchoolID string `json:"school_id"`
	LevelID  string `json:"level_id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Capacity int    `json:"capacity"`
	AuditFields
}

// SchoolLevel declares which levels a school offers in a given session.
type SchoolLevel struct {
	ID        string `json:"id"`
	SchoolID  string `json:"school_id"`
	LevelID   string `json:"level_id"`
	SessionID string `json:"session_id"`
	IsActive  bool   `json:"is_active"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// SUBJECT
// ─────────────────────────────────────────────────────────────────────────────

type SubjectCategory string

const (
	SubjectPractical SubjectCategory = "PRACTICAL"
)

// Subject is a state-wide subject definition.
type Subject struct {
	ID        string          `json:"id"`
	StateID   string          `json:"state_id"`
	Name      string          `json:"name"`
	Code      string          `json:"code"`
	Category  SubjectCategory `json:"category"`
	LevelType LevelType       `json:"level_type"`
	AuditFields
}

// SchoolSubject is a subject offered by a school for a specific level and session.
type SchoolSubject struct {
	ID        string `json:"id"`
	SchoolID  string `json:"school_id"`
	SubjectID string `json:"subject_id"`
	LevelID   string `json:"level_id"`
	SessionID string `json:"session_id"`
	TeacherID string `json:"teacher_id,omitempty"`
	IsActive  bool   `json:"is_active"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL (STAFF)
// ─────────────────────────────────────────────────────────────────────────────

type StaffRole string

const (
	StaffRoleTeacher       StaffRole = "TEACHER"
	StaffRoleHeadTeacher   StaffRole = "HEAD_TEACHER"
	StaffRolePrincipal     StaffRole = "PRINCIPAL"
	StaffRoleVicePrincipal StaffRole = "VICE_PRINCIPAL"
	StaffRoleAdmin         StaffRole = "ADMIN_OFFICER"
	StaffRoleCounselor     StaffRole = "COUNSELOR"
	StaffRoleLibrarian     StaffRole = "LIBRARIAN"
	StaffRoleLabTechnician StaffRole = "LAB_TECHNICIAN"
	StaffRoleOther         StaffRole = "OTHER"
)

type StaffStatus string

const (
	StaffStatusActive      StaffStatus = "ACTIVE"
	StaffStatusInactive    StaffStatus = "INACTIVE"
	StaffStatusSuspended   StaffStatus = "SUSPENDED"
	StaffStatusRetired     StaffStatus = "RETIRED"
	StaffStatusTransferred StaffStatus = "TRANSFERRED"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

// Personnel represents a staff member registered in the state system.
type Personnel struct {
	ID               string      `json:"id"`
	StateID          string      `json:"state_id"`
	SchoolID         string      `json:"school_id"`
	StaffID          string      `json:"staff_id"` // TSC / TRCN number
	FirstName        string      `json:"first_name"`
	MiddleName       string      `json:"middle_name,omitempty"`
	LastName         string      `json:"last_name"`
	Gender           Gender      `json:"gender"`
	DateOfBirth      *time.Time  `json:"date_of_birth,omitempty"`
	Email            string      `json:"email,omitempty"`
	Phone            string      `json:"phone,omitempty"`
	Address          string      `json:"address,omitempty"`
	Role             StaffRole   `json:"role"`
	Status           StaffStatus `json:"status"`
	Qualification    string      `json:"qualification,omitempty"`
	Specialization   string      `json:"specialization,omitempty"`
	DateOfEmployment *time.Time  `json:"date_of_employment,omitempty"`
	LGAID            string      `json:"lga_id,omitempty"`
	AvatarURL        string      `json:"avatar_url,omitempty"`
	AuditFields
}

// FullName returns the personnel's display name.
func (p *Personnel) FullName() string {
	if p.MiddleName != "" {
		return p.FirstName + " " + p.MiddleName + " " + p.LastName
	}
	return p.FirstName + " " + p.LastName
}

// PersonnelTransfer records a staff movement between schools.
type PersonnelTransfer struct {
	ID           string    `json:"id"`
	PersonnelID  string    `json:"personnel_id"`
	FromSchoolID string    `json:"from_school_id"`
	ToSchoolID   string    `json:"to_school_id"`
	TransferDate time.Time `json:"transfer_date"`
	Reason       string    `json:"reason,omitempty"`
	ApprovedBy   string    `json:"approved_by,omitempty"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT & ENROLLMENT
// ─────────────────────────────────────────────────────────────────────────────

type StudentStatus string

const (
	StudentStatusActive      StudentStatus = "ACTIVE"
	StudentStatusInactive    StudentStatus = "INACTIVE"
	StudentStatusGraduated   StudentStatus = "GRADUATED"
	StudentStatusTransferred StudentStatus = "TRANSFERRED"
	StudentStatusDropped     StudentStatus = "DROPPED_OUT"
	StudentStatusDeceased    StudentStatus = "DECEASED"
)

// Student is a learner registered in the state system.
type Student struct {
	ID               string        `json:"id"`
	StateID          string        `json:"state_id"`
	EnrollmentYear   int           `json:"enrollment_year"`
	EnrollmentNo     string        `json:"enrollment_no"`
	SerialNo         int           `json:"serial_no"`
	FirstName        string        `json:"first_name"`
	MiddleName       string        `json:"middle_name,omitempty"`
	LastName         string        `json:"last_name"`
	Gender           Gender        `json:"gender"`
	DateOfBirth      time.Time     `json:"date_of_birth"`
	StateOfOrigin    string        `json:"state_of_origin,omitempty"`
	LGAID            string        `json:"lga_id,omitempty"`
	Religion         string        `json:"religion,omitempty"`
	Address          string        `json:"address,omitempty"`
	SchoolID         string        `json:"school_id,omitempty"`
	GuardianName     string        `json:"guardian_name"`
	GuardianPhone    string        `json:"guardian_phone"`
	GuardianRelation string        `json:"guardian_relation,omitempty"`
	Status           StudentStatus `json:"status"`
	AvatarURL        string        `json:"avatar_url,omitempty"`
	AuditFields
}

// FullName returns the student's display name.
func (s *Student) FullName() string {
	if s.MiddleName != "" {
		return s.FirstName + " " + s.MiddleName + " " + s.LastName
	}
	return s.FirstName + " " + s.LastName
}

type EnrollmentStatus string

const (
	EnrollmentStatusActive      EnrollmentStatus = "ACTIVE"
	EnrollmentStatusCompleted   EnrollmentStatus = "COMPLETED"
	EnrollmentStatusRepeated    EnrollmentStatus = "REPEATED"
	EnrollmentStatusTransferred EnrollmentStatus = "TRANSFERRED"
	EnrollmentStatusWithdrawn   EnrollmentStatus = "WITHDRAWN"
)

// Enrollment links a Student to a School, Level, SubLevel and Session.
type Enrollment struct {
	ID         string           `json:"id"`
	StudentID  string           `json:"student_id"`
	SchoolID   string           `json:"school_id"`
	SessionID  string           `json:"session_id"`
	LevelID    string           `json:"level_id"`
	SubLevelID string           `json:"sub_level_id"`
	Status     EnrollmentStatus `json:"status"`
	EnrolledAt time.Time        `json:"enrolled_at"`
	AuditFields
}

type ProgressionDecision string

const (
	ProgressionPromoted  ProgressionDecision = "PROMOTED"
	ProgressionRepeated  ProgressionDecision = "REPEATED"
	ProgressionGraduated ProgressionDecision = "GRADUATED"
)

// LevelProgression records a student's transition between levels at end of session.
type LevelProgression struct {
	ID            string              `json:"id"`
	StudentID     string              `json:"student_id"`
	SchoolID      string              `json:"school_id"`
	FromSessionID string              `json:"from_session_id"`
	ToSessionID   string              `json:"to_session_id,omitempty"`
	FromLevelID   string              `json:"from_level_id"`
	ToLevelID     string              `json:"to_level_id,omitempty"`
	Decision      ProgressionDecision `json:"decision"`
	DecidedBy     string              `json:"decided_by"`
	DecisionDate  time.Time           `json:"decision_date"`
	Remarks       string              `json:"remarks,omitempty"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// EXAMINATION & RESULTS
// ─────────────────────────────────────────────────────────────────────────────

// ScoreConfig defines maximum marks per component (state-wide, per-school, or per-level).
type ScoreConfig struct {
	ID       string  `json:"id"`
	StateID  string  `json:"state_id"`
	SchoolID string  `json:"school_id,omitempty"`
	LevelID  string  `json:"level_id,omitempty"`
	CA1Max   float64 `json:"ca1_max"`
	CA2Max   float64 `json:"ca2_max"`
	CA3Max   float64 `json:"ca3_max"`
	ExamMax  float64 `json:"exam_max"`
	TotalMax float64 `json:"total_max"`
	AuditFields
}

// GradeConfig defines grading scale boundaries.
type GradeConfig struct {
	ID       string  `json:"id"`
	StateID  string  `json:"state_id"`
	SchoolID string  `json:"school_id,omitempty"`
	LevelID  string  `json:"level_id,omitempty"`
	Grade    string  `json:"grade"`
	MinScore float64 `json:"min_score"`
	MaxScore float64 `json:"max_score"`
	Remark   string  `json:"remark"`
	Points   float64 `json:"points"`
	AuditFields
}

// ScoreSheet holds raw scores for one student per subject per term.
type ScoreSheet struct {
	ID         string    `json:"id"`
	StudentID  string    `json:"student_id"`
	LevelID    string    `json:"level_id"`
	SubLevelID string    `json:"sub_level_id"`
	SchoolID   string    `json:"school_id"`
	SessionID  string    `json:"session_id"`
	TermID     string    `json:"term_id"`
	SubjectID  string    `json:"subject_id"`
	CA1Score   float64   `json:"ca1_score"`
	CA2Score   float64   `json:"ca2_score"`
	CA3Score   float64   `json:"ca3_score"`
	ExamScore  float64   `json:"exam_score"`
	TotalScore float64   `json:"total_score"` // computed by domain
	Grade      string    `json:"grade"`
	Remark     string    `json:"remark"`
	Position   int       `json:"position"`
	RecordedBy string    `json:"recorded_by"`
	RecordedAt time.Time `json:"recorded_at"`
	AuditFields
}

// ComputeTotal sums all component scores — the only domain business rule for scoring.
func (s *ScoreSheet) ComputeTotal() {
	s.TotalScore = s.CA1Score + s.CA2Score + s.CA3Score + s.ExamScore
}

// ReportCard is the compiled end-of-term result per student.
type ReportCard struct {
	ID              string     `json:"id"`
	StudentID       string     `json:"student_id"`
	SchoolID        string     `json:"school_id"`
	SessionID       string     `json:"session_id"`
	TermID          string     `json:"term_id"`
	LevelID         string     `json:"level_id"`
	SubLevelID      string     `json:"sub_level_id"`
	TotalScore      float64    `json:"total_score"`
	AverageScore    float64    `json:"average_score"`
	OverallGrade    string     `json:"overall_grade"`
	ClassPosition   int        `json:"class_position"`
	SubjectCount    int        `json:"subject_count"`
	Attendance      int        `json:"attendance"`
	TotalSchoolDays int        `json:"total_school_days"`
	PrincipalRemark string     `json:"principal_remark,omitempty"`
	TeacherRemark   string     `json:"teacher_remark,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	AuditFields
}

// IsPublished returns whether the report card is visible to students/parents.
func (r *ReportCard) IsPublished() bool { return r.PublishedAt != nil }

// ─────────────────────────────────────────────────────────────────────────────
// ATTENDANCE
// ─────────────────────────────────────────────────────────────────────────────

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "PRESENT"
	AttendanceStatusAbsent  AttendanceStatus = "ABSENT"
	AttendanceStatusLate    AttendanceStatus = "LATE"
	AttendanceStatusExcused AttendanceStatus = "EXCUSED"
)

type PersonnelAttendance struct {
	ID             string           `json:"id"`
	PersonnelID    string           `json:"personnel_id"`
	SchoolID       string           `json:"school_id"`
	AttendanceDate time.Time        `json:"attendance_date"`
	Status         AttendanceStatus `json:"status"`
	Remarks        string           `json:"remarks,omitempty"`
	RecordedBy     string           `json:"recorded_by"`
	AuditFields
}

type StudentAttendance struct {
	ID             string           `json:"id"`
	StudentID      string           `json:"student_id"`
	SchoolID       string           `json:"school_id"`
	AttendanceDate time.Time        `json:"attendance_date"`
	Status         AttendanceStatus `json:"status"`
	Remarks        string           `json:"remarks,omitempty"`
	RecordedBy     string           `json:"recorded_by"`
	AuditFields
}

// ─────────────────────────────────────────────────────────────────────────────
// DASHBOARD STATISTICS
// ─────────────────────────────────────────────────────────────────────────────

// DashboardStats represents aggregated statistics for a dashboard view.
type DashboardStats struct {
	TotalSchools      int `json:"total_schools"`
	TotalStudents     int `json:"total_students"`
	TeachingPersonnel int `json:"teaching_personnel"`
	ResultCompletion  int `json:"result_completion"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ZONE SUMMARY REPORT
// ─────────────────────────────────────────────────────────────────────────────


type ZoneSummaryReport struct {
	Zone                string   `json:"zone" gorm:"column:zone"`
	SchoolCount         int64    `json:"school" gorm:"column:school"`
	TeachingStaffCount  int64    `json:"teaching_staff" gorm:"column:teaching_staff"`
	StudentCount        int64    `json:"students" gorm:"column:students"`
	StudentTeacherRatio *float64 `json:"students_teachers_ratio,omitempty" gorm:"column:students_teachers_ratio"`
}

// PublicTeachingPersonnel represents the total number of teachers across the entire system.
type PublicTeachingPersonnel struct {
	Total int `json:"total"`
}

// GenderReport represents gender distribution statistics.
type GenderReport struct {
	StateID   string  `json:"state_id"`
	Male      int     `json:"male"`
	Female    int     `json:"female"`
	Other     int     `json:"other"`
	Total     int     `json:"total"`
	MalePct   float64 `json:"male_pct"`
	FemalePct float64 `json:"female_pct"`
	OtherPct  float64 `json:"other_pct"`
}

type TotalSchools struct {
	Total int `json:"total"`
}

// ─────────────────────────────────────────────────────────────────────────────
// INPUT SIGNALS  (raw data the engine reads from the DB)
// ─────────────────────────────────────────────────────────────────────────────

// FacilitySignal is the aggregated facility profile for one school.
type FacilitySignal struct {
	SchoolID        string
	TotalFacilities int
	GoodCount       int // condition = GOOD
	FairCount       int // condition = FAIR
	PoorCount       int // condition = POOR
	DefunctCount    int // condition = DEFUNCT
	HasLibrary      bool
	HasLab          bool
	HasICT          bool
	HasSportField   bool
	FacilityScore   float64 // 0–100, computed by engine
}

// PersonnelSignal is the aggregated staff profile for one school.
type PersonnelSignal struct {
	SchoolID          string
	TotalStaff        int
	ActiveStaff       int
	QualifiedCount    int     // NCE / B.Ed / M.Ed / Ph.D
	PostgradCount     int     // M.Ed or Ph.D
	StaffStudentRatio float64 // active staff / enrolled students
	TeacherCount      int     // role = TEACHER
	PersonnelScore    float64 // 0–100, computed by engine
}

// HistoricalSignal holds aggregated academic performance for a school or student.
type HistoricalSignal struct {
	SchoolID        string
	StudentID       string  // empty = school-level signal
	AvgScore        float64 // mean total_score across all score_sheets
	PassRate        float64 // % of score_sheets with total_score >= 40
	DistinctionRate float64 // % with total_score >= 70
	TermsRecorded   int     // how many term results exist
}

// ─────────────────────────────────────────────────────────────────────────────
// PREDICTION OUTPUTS
// ─────────────────────────────────────────────────────────────────────────────

// RiskLevel classifies a student's likelihood of poor academic performance.
type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"    // predicted avg >= 60
	RiskMedium RiskLevel = "MEDIUM" // predicted avg 40–59
	RiskHigh   RiskLevel = "HIGH"   // predicted avg < 40
)

// SchoolRating classifies overall school performance capacity.
type SchoolRating string

const (
	RatingExcellent SchoolRating = "EXCELLENT" // composite >= 80
	RatingGood      SchoolRating = "GOOD"      // composite 65–79
	RatingAverage   SchoolRating = "AVERAGE"   // composite 50–64
	RatingPoor      SchoolRating = "POOR"      // composite < 50
)

// ScoreRange is a human-readable predicted score bracket.
type ScoreRange struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Label string  `json:"label"` // e.g. "60–75%"
}

// StudentPrediction is the per-student output.
type StudentPrediction struct {
	StudentID      string     `json:"student_id"`
	StudentName    string     `json:"student_name"`
	SchoolID       string     `json:"school_id"`
	RiskLevel      RiskLevel  `json:"risk_level"`
	PredictedRange ScoreRange `json:"predicted_score_range"`
	HistoricalAvg  float64    `json:"historical_avg"`
	FacilityScore  float64    `json:"facility_score"`
	PersonnelScore float64    `json:"personnel_score"`
	CompositeScore float64    `json:"composite_score"` // weighted blend
	Confidence     float64    `json:"confidence"`      // 0–1, based on data richness
	Factors        []Factor   `json:"contributing_factors"`
	GeneratedAt    time.Time  `json:"generated_at"`
}

// SchoolPrediction is the school-level output.
type SchoolPrediction struct {
	SchoolID          string       `json:"school_id"`
	SchoolName        string       `json:"school_name"`
	Rating            SchoolRating `json:"rating"`
	CompositeScore    float64      `json:"composite_score"`
	FacilityScore     float64      `json:"facility_score"`
	PersonnelScore    float64      `json:"personnel_score"`
	HistoricalScore   float64      `json:"historical_score"`
	PredictedPassRate float64      `json:"predicted_pass_rate"`
	HighRiskCount     int          `json:"high_risk_student_count"`
	MediumRiskCount   int          `json:"medium_risk_student_count"`
	LowRiskCount      int          `json:"low_risk_student_count"`
	Factors           []Factor     `json:"contributing_factors"`
	GeneratedAt       time.Time    `json:"generated_at"`
}

// Factor is one named contributor to the prediction — shown in the UI as insight cards.
type Factor struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`  // 0–100
	Weight float64 `json:"weight"` // 0–1, how much this factor contributes
	Detail string  `json:"detail"` // human-readable explanation
}

// PredictionReport bundles school + all student predictions for one request.
type PredictionReport struct {
	School   SchoolPrediction    `json:"school"`
	Students []StudentPrediction `json:"students"`
}

// IsEmpty returns whether the prediction report contains any data.
func (pr *PredictionReport) IsEmpty() bool { return len(pr.Students) == 0 }
