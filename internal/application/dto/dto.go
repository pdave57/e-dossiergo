// Package dto defines all request and response transfer objects.
// These live in the application layer and cross the boundary between
// the HTTP interface and use-case layer.
package dto

import (
	"time"

	"github.com/edossier/api/internal/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// AUTH
// ─────────────────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	StateID   string `json:"state_id"   validate:"required"`
	SchoolID  string `json:"school_id"`
	Email     string `json:"email"      validate:"required"`
	Password  string `json:"password"   validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"  validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID          string     `json:"id"`
	StateID     string     `json:"state_id"`
	SchoolID    string     `json:"school_id,omitempty"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Status      string     `json:"status"`
	Roles       []string   `json:"roles,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// USER MANAGEMENT
// ─────────────────────────────────────────────────────────────────────────────

type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"  validate:"required"`
	SchoolID  string `json:"school_id"`
	Status    string `json:"status"`
}

type AssignRoleRequest struct {
	RoleID   string `json:"role_id"   validate:"required"`
	SchoolID string `json:"school_id"`
}

type RevokeRoleRequest struct {
	RoleID string `json:"role_id" validate:"required"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ROLE / PERMISSION
// ─────────────────────────────────────────────────────────────────────────────

type CreateRoleRequest struct {
	Name        string `json:"name"        validate:"required"`
	Code        string `json:"code"        validate:"required"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"        validate:"required"`
	Description string `json:"description"`
}

type RoleResponse struct {
	ID          string               `json:"id"`
	StateID     string               `json:"state_id"`
	Name        string               `json:"name"`
	Code        string               `json:"code"`
	Description string               `json:"description,omitempty"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
}

type AddPermissionRequest struct {
	PermissionID string `json:"permission_id" validate:"required"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ATTENDANCE
// ─────────────────────────────────────────────────────────────────────────────

type PersonnelAttendanceRequest struct {
	PersonnelID string    `json:"personnel_id" validate:"required"`
	SchoolID    string    `json:"school_id"    validate:"required"`
	Date        time.Time `json:"date"         validate:"required"`
	Status      string    `json:"status"       validate:"required,oneof=PRESENT ABSENT LATE EXCUSED"`
	Remarks     string    `json:"remarks"`
	RecordedBy  string    `json:"recorded_by"  validate:"required"`
}

type UpdatePersonnelAttendanceRequest struct {
	Status  string `json:"status"       validate:"required,oneof=PRESENT ABSENT LATE EXCUSED"`
	Remarks string `json:"remarks"`
}

type PersonnelAttendanceResponse struct {
	ID            string    `json:"id"`
	PersonnelID   string    `json:"personnel_id"`
	SchoolID      string    `json:"school_id"`
	AttendanceDate time.Time `json:"attendance_date"`
	Status        string    `json:"status"`
	Remarks       string    `json:"remarks,omitempty"`
	RecordedBy    string    `json:"recorded_by"`
	PersonnelName string    `json:"personnel_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StudentAttendanceRequest struct {
	StudentID  string    `json:"student_id" validate:"required"`
	SchoolID   string    `json:"school_id"  validate:"required"`
	Date       time.Time `json:"date"       validate:"required"`
	Status     string    `json:"status"     validate:"required,oneof=PRESENT ABSENT LATE EXCUSED"`
	Remarks    string    `json:"remarks"`
	RecordedBy string    `json:"recorded_by" validate:"required"`
}

type UpdateStudentAttendanceRequest struct {
	Status  string `json:"status"     validate:"required,oneof=PRESENT ABSENT LATE EXCUSED"`
	Remarks string `json:"remarks"`
}

type StudentAttendanceResponse struct {
	ID            string    `json:"id"`
	StudentID     string    `json:"student_id"`
	SchoolID      string    `json:"school_id"`
	AttendanceDate time.Time `json:"attendance_date"`
	Status        string    `json:"status"`
	Remarks       string    `json:"remarks,omitempty"`
	RecordedBy    string    `json:"recorded_by"`
	StudentName   string    `json:"student_name,omitempty"`
	EnrollmentNo  string    `json:"enrollment_no,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BulkStudentAttendanceRequest struct {
	SchoolID   string                       `json:"school_id"   validate:"required"`
	Date       time.Time                    `json:"date"        validate:"required"`
	RecordedBy string                       `json:"recorded_by" validate:"required"`
	Records    []StudentAttendanceRequest   `json:"records"     validate:"required"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ZONES / LGAS
// ─────────────────────────────────────────────────────────────────────────────

type CreateStateRequest struct {
	Name    string `json:"name"    validate:"required"`
	Code    string `json:"code"    validate:"required"`
	Country string `json:"country"`
}

type UpdateStateRequest struct {
	Name    string `json:"name"    validate:"required"`
	Country string `json:"country"`
}

type CreateZoneRequest struct {
	Name string `json:"name" validate:"required"`
	Code string `json:"code" validate:"required"`
}

type UpdateZoneRequest struct {
	Name string `json:"name" validate:"required"`
	Code string `json:"code" validate:"required"`
}

type CreateLGARequest struct {
	ZoneID string `json:"zone_id" validate:"required"`
	Name   string `json:"name"    validate:"required"`
	Code   string `json:"code"    validate:"required"`
}

type UpdateLGARequest struct {
	StateID string `json:"state_id" validate:"required"`
	ZoneID  string `json:"zone_id"  validate:"required"`
	Name    string `json:"name"     validate:"required"`
	Code    string `json:"code"     validate:"required"`
}

// ─────────────────────────────────────────────────────────────────────────────
// SCHOOL
// ─────────────────────────────────────────────────────────────────────────────

type CreateSchoolRequest struct {
	StateID            string `json:"state_id" validate:"required"`
	ZoneID             string `json:"zone_id" validate:"required"`
	LGAID              string `json:"lga_id" validate:"required"`
	Name               string `json:"name" validate:"required"`
	Code               string `json:"code" validate:"required"`
	Category           string `json:"category" validate:"required"`
	Ownership          string `json:"ownership"`
	Status             string `json:"status"`
	NumberOfClassrooms int    `json:"number_of_classrooms"`
	TotalStudents      int    `json:"total_students"`
	Address            string `json:"address"`
	HeadTeacher        string `json:"head_teacher,omitempty"`
	Founded            *int   `json:"founded,omitempty"`
}

type UpdateSchoolRequest struct {
	StateID            string `json:"state_id" validate:"required"`
	ZoneID             string `json:"zone_id" validate:"required"`
	LGAID              string `json:"lga_id" validate:"required"`
	Name               string `json:"name" validate:"required"`
	Code               string `json:"code" validate:"required"`
	Category           string `json:"category" validate:"required"`
	Ownership          string `json:"ownership"`
	Status             string `json:"status"`
	NumberOfClassrooms int    `json:"number_of_classrooms"`
	TotalStudents      int    `json:"total_students"`
	Address            string `json:"address"`
	HeadTeacher        string `json:"head_teacher,omitempty"`
	Founded            *int   `json:"founded,omitempty"`
}

type CreateFacilityRequest struct {
	Type      string `json:"type"      validate:"required"`
	Name      string `json:"name"      validate:"required"`
	Quantity  int    `json:"quantity"`
	Condition string `json:"condition" validate:"required"`
	Notes     string `json:"notes"`
}

type UpdateFacilityRequest struct {
	Type      string `json:"type"      validate:"required"`
	Name      string `json:"name"      validate:"required"`
	Quantity  int    `json:"quantity"`
	Condition string `json:"condition" validate:"required"`
	Notes     string `json:"notes"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ACADEMIC SESSION & TERMS
// ─────────────────────────────────────────────────────────────────────────────

type CreateSessionRequest struct {
	Name      string    `json:"name"       validate:"required"`
	StartYear int       `json:"start_year" validate:"required"`
	EndYear   int       `json:"end_year"   validate:"required"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date"   validate:"required"`
}

type UpdateSessionRequest struct {
	Name      string    `json:"name"       validate:"required"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
}

type CreateTermRequest struct {
	SessionID string    `json:"session_id"`
	Number    int       `json:"term_number" validate:"required"`
	Name      string    `json:"name"        validate:"required"`
	StartDate time.Time `json:"start_date"  validate:"required"`
	EndDate   time.Time `json:"end_date"    validate:"required"`
}

type UpdateTermRequest struct {
	Name      string    `json:"name"       validate:"required"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// ─────────────────────────────────────────────────────────────────────────────
// LEVEL
// ─────────────────────────────────────────────────────────────────────────────

type CreateLevelRequest struct {
	Name  string `json:"name"  validate:"required"`
	Code  string `json:"code"  validate:"required"`
	Type  string `json:"type"  validate:"required"`
	Order int    `json:"order"`
}

type UpdateLevelRequest struct {
	Name  string `json:"name"  validate:"required"`
	Code  string `json:"code"  validate:"required"`
	Type  string `json:"type"  validate:"required"`
	Order int    `json:"order"`
}

type CreateSubLevelRequest struct {
	SchoolID string `json:"school_id"`
	LevelID  string `json:"level_id"  validate:"required"`
	Name     string `json:"name"      validate:"required"`
	Code     string `json:"code"      validate:"required"`
	Capacity int    `json:"capacity"`
}

type UpdateSubLevelRequest struct {
	Name     string `json:"name"     validate:"required"`
	Code     string `json:"code"     validate:"required"`
	Capacity int    `json:"capacity"`
}

type UpsertSchoolLevelRequest struct {
	LevelID   string `json:"level_id"   validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
	IsActive  bool   `json:"is_active"`
}

// ─────────────────────────────────────────────────────────────────────────────
// SUBJECT
// ─────────────────────────────────────────────────────────────────────────────

type CreateSubjectRequest struct {
	Name      string `json:"name"       validate:"required"`
	Code      string `json:"code"       validate:"required"`
	Category  string `json:"category"   validate:"required"`
	LevelType string `json:"level_type" validate:"required"`
}

type UpdateSubjectRequest struct {
	Name      string `json:"name"       validate:"required"`
	Code      string `json:"code"       validate:"required"`
	Category  string `json:"category"`
	LevelType string `json:"level_type"`
}

type CreateSchoolSubjectRequest struct {
	SubjectID string `json:"subject_id" validate:"required"`
	LevelID   string `json:"level_id"   validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
	TeacherID string `json:"teacher_id"`
}

type UpdateSchoolSubjectRequest struct {
	TeacherID string `json:"teacher_id"`
	IsActive  bool   `json:"is_active"`
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL
// ─────────────────────────────────────────────────────────────────────────────

type CreatePersonnelRequest struct {
	SchoolID         string     `json:"school_id"          validate:"required"`
	StaffID          string     `json:"staff_id"           validate:"required"`
	FirstName        string     `json:"first_name"         validate:"required"`
	MiddleName       string     `json:"middle_name"`
	LastName         string     `json:"last_name"          validate:"required"`
	Gender           string     `json:"gender"             validate:"required"`
	DateOfBirth      *time.Time `json:"date_of_birth"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	Address          string     `json:"address"`
	Role             string     `json:"role"               validate:"required"`
	Qualification    string     `json:"qualification"`
	Specialization   string     `json:"specialization"`
	DateOfEmployment *time.Time `json:"date_of_employment"`
	LGAID            string     `json:"lga_id"`
}

type UpdatePersonnelRequest struct {
	SchoolID         string     `json:"school_id"`
	StaffID          string     `json:"staff_id"           validate:"required"`
	FirstName        string     `json:"first_name"  validate:"required"`
	MiddleName       string     `json:"middle_name"`
	LastName         string     `json:"last_name"   validate:"required"`
	Gender           string     `json:"gender"      validate:"required"`
	DateOfBirth      *time.Time `json:"date_of_birth"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	Address          string     `json:"address"`
	Role             string     `json:"role"        validate:"required"`
	Status           string     `json:"status"`
	Qualification    string     `json:"qualification"`
	Specialization   string     `json:"specialization"`
	DateOfEmployment *time.Time `json:"date_of_employment"`
}

type TransferPersonnelRequest struct {
	ToSchoolID   string    `json:"to_school_id"  validate:"required"`
	TransferDate time.Time `json:"transfer_date" validate:"required"`
	Reason       string    `json:"reason"`
	ApprovedBy   string    `json:"approved_by"`
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT
// ─────────────────────────────────────────────────────────────────────────────

type CreateStudentRequest struct {
	SchoolID         string    `json:"school_id"         validate:"required"`
	EnrollmentYear   int       `json:"enrollment_year"   validate:"required"`
	FirstName        string    `json:"first_name"        validate:"required"`
	MiddleName       string    `json:"middle_name"`
	LastName         string    `json:"last_name"         validate:"required"`
	Gender           string    `json:"gender"            validate:"required"`
	DateOfBirth      time.Time `json:"date_of_birth"     validate:"required"`
	StateOfOrigin    string    `json:"state_of_origin"`
	LGAID            string    `json:"lga_id"`
	Religion         string    `json:"religion"`
	Address          string    `json:"address"`
	GuardianName     string    `json:"guardian_name"     validate:"required"`
	GuardianPhone    string    `json:"guardian_phone"    validate:"required"`
	GuardianRelation string    `json:"guardian_relation"`
}

type UpdateStudentRequest struct {
	FirstName        string    `json:"first_name"     validate:"required"`
	MiddleName       string    `json:"middle_name"`
	LastName         string    `json:"last_name"      validate:"required"`
	Gender           string    `json:"gender"         validate:"required"`
	DateOfBirth      time.Time `json:"date_of_birth"`
	StateOfOrigin    string    `json:"state_of_origin"`
	LGAID            string    `json:"lga_id"`
	Religion         string    `json:"religion"`
	Address          string    `json:"address"`
	GuardianName     string    `json:"guardian_name"  validate:"required"`
	GuardianPhone    string    `json:"guardian_phone" validate:"required"`
	GuardianRelation string    `json:"guardian_relation"`
	Status           string    `json:"status"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ENROLLMENT
// ─────────────────────────────────────────────────────────────────────────────

type EnrollStudentRequest struct {
	StudentID  string `json:"student_id"   validate:"required"`
	SchoolID   string `json:"school_id"    validate:"required"`
	SessionID  string `json:"session_id"   validate:"required"`
	LevelID    string `json:"level_id"     validate:"required"`
	SubLevelID string `json:"sub_level_id" validate:"required"`
}

type UpdateEnrollmentRequest struct {
	SubLevelID string `json:"sub_level_id"`
	Status     string `json:"status"`
}

type RecordProgressionRequest struct {
	StudentID     string    `json:"student_id"      validate:"required"`
	FromSessionID string    `json:"from_session_id" validate:"required"`
	ToSessionID   string    `json:"to_session_id"`
	FromLevelID   string    `json:"from_level_id"   validate:"required"`
	ToLevelID     string    `json:"to_level_id"`
	Decision      string    `json:"decision"        validate:"required"`
	DecisionDate  time.Time `json:"decision_date"   validate:"required"`
	Remarks       string    `json:"remarks"`
}

// ComputePromotionsRequest asks the backend to evaluate third-term scores and
// auto-advance students whose class-wide average is >= 35.  Students below the
// threshold are still progressed but flagged for principal review.
type ComputePromotionsRequest struct {
	SchoolID   string `json:"school_id"   validate:"required"`
	SessionID  string `json:"session_id"  validate:"required"`
	LevelID    string `json:"level_id"    validate:"required"`
	SubLevelID string `json:"sub_level_id" validate:"required"`
}

// PromotedStudent captures the outcome for one learner during promotion review.
type PromotedStudent struct {
	StudentID  string `json:"student_id"`
	StudentName string `json:"student_name"`
	FromLevelID   string `json:"from_level_id"`
	FromLevelName string `json:"from_level_name"`
	ToLevelID     string `json:"to_level_id,omitempty"`
	ToLevelName   string `json:"to_level_name,omitempty"`
	AverageScore  float64 `json:"average_score"`
	Decision      string `json:"decision"`
	Remarks       string `json:"remarks,omitempty"`
}

type ComputePromotionsResponse struct {
	Total   int                `json:"total"`
	Results []*PromotedStudent `json:"results"`
}

// ─────────────────────────────────────────────────────────────────────────────
// RESULTS
// ─────────────────────────────────────────────────────────────────────────────

type UpsertScoreRequest struct {
	StudentID  string  `json:"student_id"  validate:"required"`
	LevelID    string  `json:"level_id"    validate:"required"`
	SubLevelID string  `json:"sub_level_id" validate:"required"`
	SubjectID  string  `json:"subject_id"  validate:"required"`
	TermID     string  `json:"term_id"     validate:"required"`
	CA1Score   float64 `json:"ca1_score"`
	CA2Score   float64 `json:"ca2_score"`
	CA3Score   float64 `json:"ca3_score"`
	ExamScore  float64 `json:"exam_score"`
}

type BulkUpsertScoreRequest struct {
	Scores []UpsertScoreRequest `json:"scores" validate:"required"`
}

type GenerateReportCardRequest struct {
	SchoolID   string `json:"school_id"   validate:"required"`
	SessionID  string `json:"session_id"  validate:"required"`
	TermID     string `json:"term_id"     validate:"required"`
	SubLevelID string `json:"sub_level_id" validate:"required"`
}

type UpdateReportCardRemarksRequest struct {
	PrincipalRemark string `json:"principal_remark"`
	TeacherRemark   string `json:"teacher_remark"`
	Attendance      int    `json:"attendance"`
	TotalSchoolDays int    `json:"total_school_days"`
}

type UpsertScoreConfigRequest struct {
	SchoolID string  `json:"school_id"` // optional — empty = state default
	LevelID  string  `json:"level_id"`  // optional — empty = school/state default
	CA1Max   float64 `json:"ca1_max"   validate:"required"`
	CA2Max   float64 `json:"ca2_max"   validate:"required"`
	CA3Max   float64 `json:"ca3_max"   validate:"required"`
	ExamMax  float64 `json:"exam_max"  validate:"required"`
	TotalMax float64 `json:"total_max" validate:"required"`
}

type UpsertGradeConfigRequest struct {
	SchoolID string  `json:"school_id"` // optional
	LevelID  string  `json:"level_id"`  // optional
	Grade    string  `json:"grade"      validate:"required"`
	MinScore float64 `json:"min_score"`
	MaxScore float64 `json:"max_score"  validate:"required"`
	Remark   string  `json:"remark"     validate:"required"`
	Points   float64 `json:"points"`
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT AUTH
// ─────────────────────────────────────────────────────────────────────────────

type StudentLoginRequest struct {
	SchoolCode   string `json:"school_code"   validate:"required"`
	EnrollmentNo string `json:"enrollment_no" validate:"required"`
}

type StudentLoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    time.Time   `json:"expires_at"`
	Student      StudentInfo `json:"student"`
}

type StudentInfo struct {
	ID           string `json:"id"`
	StateID      string `json:"state_id"`
	SchoolID     string `json:"school_id,omitempty"`
	EnrollmentNo string `json:"enrollment_no"`
	FirstName    string `json:"first_name"`
	MiddleName   string `json:"middle_name,omitempty"`
	LastName     string `json:"last_name"`
	Gender       string `json:"gender"`
	Status       string `json:"status"`
}

// ─────────────────────────────────────────────────────────────────────────────
// NEWS & ANNOUNCEMENTS
// ─────────────────────────────────────────────────────────────────────────────

// CreateNewsRequest is the payload for publishing a news item or announcement.
// state_id is optional — it falls back to the caller's state from the JWT.
// type is optional and defaults to NEWS.
type CreateNewsRequest struct {
	StateID     string    `json:"state_id"`
	Type        string    `json:"type"`
	Headline    string    `json:"headline"     validate:"required"`
	SubHeadline string    `json:"sub_headline"`
	NewsBody    string    `json:"news_body"`
	NewsDate    time.Time `json:"news_date"    validate:"required"`
	PostedBy    string    `json:"posted_by"    validate:"required"`
}

// UpdateNewsRequest is the payload for editing an existing news item.
// The owning state is immutable and therefore not accepted here.
type UpdateNewsRequest struct {
	Type        string    `json:"type"`
	Headline    string    `json:"headline"     validate:"required"`
	SubHeadline string    `json:"sub_headline"`
	NewsBody    string    `json:"news_body"`
	NewsDate    time.Time `json:"news_date"    validate:"required"`
	PostedBy    string    `json:"posted_by"    validate:"required"`
}

// ─────────────────────────────────────────────────────────────────────────────
// GENERIC LIST RESPONSE WRAPPER
// ─────────────────────────────────────────────────────────────────────────────

type ListResponse[T any] struct {
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ─────────────────────────────────────────────────────────────────────────────
// OUT-OF-SCHOOL CHILDREN (OSC) REPORTS
// ─────────────────────────────────────────────────────────────────────────────

type OSCReportResponse struct {
	StateID string          `json:"state_id"`
	Data    []domain.OSCReportRow `json:"data"`
}
