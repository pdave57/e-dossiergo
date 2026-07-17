package service

import (
	"context"
	"time"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/validator"
)

// AttendanceService handles daily attendance for personnel and students.
type AttendanceService struct {
	personnelAttendances domain.PersonnelAttendanceRepository
	studentAttendances   domain.StudentAttendanceRepository
	personnelRepo        domain.PersonnelRepository
	studentRepo          domain.StudentRepository
	schoolRepo           domain.SchoolRepository
}

func NewAttendanceService(
	personnelAttendances domain.PersonnelAttendanceRepository,
	studentAttendances domain.StudentAttendanceRepository,
	personnelRepo domain.PersonnelRepository,
	studentRepo domain.StudentRepository,
	schoolRepo domain.SchoolRepository,
) *AttendanceService {
	return &AttendanceService{
		personnelAttendances: personnelAttendances,
		studentAttendances:   studentAttendances,
		personnelRepo:        personnelRepo,
		studentRepo:          studentRepo,
		schoolRepo:           schoolRepo,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL ATTENDANCE
// ─────────────────────────────────────────────────────────────────────────────

func (uc *AttendanceService) RecordPersonnelAttendance(ctx context.Context, req dto.PersonnelAttendanceRequest, createdBy string) (*dto.PersonnelAttendanceResponse, error) {
	v := validator.New().
		Required(req.PersonnelID, "personnel_id").
		Required(req.SchoolID, "school_id").
		Required(req.RecordedBy, "recorded_by").
		Check(!req.Date.IsZero(), "date", "is required").
		OneOf(req.Status, []string{"PRESENT", "ABSENT", "LATE", "EXCUSED"}, "status")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	if _, err := uc.personnelRepo.GetByID(ctx, req.PersonnelID); err != nil {
		return nil, apperror.NotFound("personnel", req.PersonnelID)
	}
	if _, err := uc.schoolRepo.GetByID(ctx, req.SchoolID); err != nil {
		return nil, apperror.NotFound("school", req.SchoolID)
	}

	attDate := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, time.UTC)
	a := &domain.PersonnelAttendance{
		PersonnelID:    req.PersonnelID,
		SchoolID:       req.SchoolID,
		AttendanceDate: attDate,
		Status:         domain.AttendanceStatus(req.Status),
		Remarks:        req.Remarks,
		RecordedBy:     req.RecordedBy,
		AuditFields:    domain.AuditFields{CreatedBy: createdBy, UpdatedBy: createdBy},
	}
	if err := uc.personnelAttendances.Create(ctx, a); err != nil {
		return nil, err
	}
	return uc.toPersonnelAttendanceResponse(ctx, a), nil
}

func (uc *AttendanceService) GetPersonnelAttendance(ctx context.Context, id string) (*dto.PersonnelAttendanceResponse, error) {
	a, err := uc.personnelAttendances.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return uc.toPersonnelAttendanceResponse(ctx, a), nil
}

func (uc *AttendanceService) UpdatePersonnelAttendance(ctx context.Context, id string, req dto.UpdatePersonnelAttendanceRequest, updatedBy string) (*dto.PersonnelAttendanceResponse, error) {
	v := validator.New().
		OneOf(req.Status, []string{"PRESENT", "ABSENT", "LATE", "EXCUSED"}, "status")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	a, err := uc.personnelAttendances.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Status = domain.AttendanceStatus(req.Status)
	a.Remarks = req.Remarks
	a.RecordedBy = updatedBy
	a.UpdatedBy = updatedBy
	if err := uc.personnelAttendances.Update(ctx, a); err != nil {
		return nil, err
	}
	return uc.toPersonnelAttendanceResponse(ctx, a), nil
}

func (uc *AttendanceService) DeletePersonnelAttendance(ctx context.Context, id string) error {
	return uc.personnelAttendances.Delete(ctx, id)
}

func (uc *AttendanceService) ListPersonnelAttendanceBySchoolAndDate(ctx context.Context, schoolID string, date time.Time) ([]*dto.PersonnelAttendanceResponse, error) {
	records, err := uc.personnelAttendances.ListBySchoolAndDate(ctx, schoolID, date)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.PersonnelAttendanceResponse, len(records))
	for i, a := range records {
		out[i] = uc.toPersonnelAttendanceResponse(ctx, a)
	}
	return out, nil
}

func (uc *AttendanceService) ListPersonnelAttendanceByPersonnelAndRange(ctx context.Context, personnelID string, from, to time.Time) ([]*dto.PersonnelAttendanceResponse, error) {
	records, err := uc.personnelAttendances.ListByPersonnelAndRange(ctx, personnelID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.PersonnelAttendanceResponse, len(records))
	for i, a := range records {
		out[i] = uc.toPersonnelAttendanceResponse(ctx, a)
	}
	return out, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT ATTENDANCE
// ─────────────────────────────────────────────────────────────────────────────

func (uc *AttendanceService) RecordStudentAttendance(ctx context.Context, req dto.StudentAttendanceRequest, createdBy string) (*dto.StudentAttendanceResponse, error) {
	v := validator.New().
		Required(req.StudentID, "student_id").
		Required(req.SchoolID, "school_id").
		Required(req.RecordedBy, "recorded_by").
		Check(!req.Date.IsZero(), "date", "is required").
		OneOf(req.Status, []string{"PRESENT", "ABSENT", "LATE", "EXCUSED"}, "status")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	if _, err := uc.studentRepo.GetByID(ctx, req.StudentID); err != nil {
		return nil, apperror.NotFound("student", req.StudentID)
	}
	if _, err := uc.schoolRepo.GetByID(ctx, req.SchoolID); err != nil {
		return nil, apperror.NotFound("school", req.SchoolID)
	}

	attDate := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, time.UTC)
	a := &domain.StudentAttendance{
		StudentID:      req.StudentID,
		SchoolID:       req.SchoolID,
		AttendanceDate: attDate,
		Status:         domain.AttendanceStatus(req.Status),
		Remarks:        req.Remarks,
		RecordedBy:     req.RecordedBy,
		AuditFields:    domain.AuditFields{CreatedBy: createdBy, UpdatedBy: createdBy},
	}
	if err := uc.studentAttendances.Create(ctx, a); err != nil {
		return nil, err
	}
	return uc.toStudentAttendanceResponse(ctx, a), nil
}

func (uc *AttendanceService) BulkRecordStudentAttendance(ctx context.Context, req dto.BulkStudentAttendanceRequest, createdBy string) ([]*dto.StudentAttendanceResponse, error) {
	v := validator.New().
		Required(req.SchoolID, "school_id").
		Required(req.RecordedBy, "recorded_by").
		Check(!req.Date.IsZero(), "date", "is required").
		Check(len(req.Records) > 0, "records", "must not be empty")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	if _, err := uc.schoolRepo.GetByID(ctx, req.SchoolID); err != nil {
		return nil, apperror.NotFound("school", req.SchoolID)
	}

	attDate := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, time.UTC)
	out := make([]*dto.StudentAttendanceResponse, 0, len(req.Records))
	for _, rec := range req.Records {
		if _, err := uc.studentRepo.GetByID(ctx, rec.StudentID); err != nil {
			continue
		}
		a := &domain.StudentAttendance{
			StudentID:      rec.StudentID,
			SchoolID:       req.SchoolID,
			AttendanceDate: attDate,
			Status:         domain.AttendanceStatus(rec.Status),
			Remarks:        rec.Remarks,
			RecordedBy:     req.RecordedBy,
			AuditFields:    domain.AuditFields{CreatedBy: createdBy, UpdatedBy: createdBy},
		}
		if err := uc.studentAttendances.Create(ctx, a); err != nil {
			continue
		}
		out = append(out, uc.toStudentAttendanceResponse(ctx, a))
	}
	return out, nil
}

func (uc *AttendanceService) GetStudentAttendance(ctx context.Context, id string) (*dto.StudentAttendanceResponse, error) {
	a, err := uc.studentAttendances.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return uc.toStudentAttendanceResponse(ctx, a), nil
}

func (uc *AttendanceService) UpdateStudentAttendance(ctx context.Context, id string, req dto.UpdateStudentAttendanceRequest, updatedBy string) (*dto.StudentAttendanceResponse, error) {
	v := validator.New().
		OneOf(req.Status, []string{"PRESENT", "ABSENT", "LATE", "EXCUSED"}, "status")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	a, err := uc.studentAttendances.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Status = domain.AttendanceStatus(req.Status)
	a.Remarks = req.Remarks
	a.RecordedBy = updatedBy
	a.UpdatedBy = updatedBy
	if err := uc.studentAttendances.Update(ctx, a); err != nil {
		return nil, err
	}
	return uc.toStudentAttendanceResponse(ctx, a), nil
}

func (uc *AttendanceService) DeleteStudentAttendance(ctx context.Context, id string) error {
	return uc.studentAttendances.Delete(ctx, id)
}

func (uc *AttendanceService) ListStudentAttendanceBySchoolAndDate(ctx context.Context, schoolID string, date time.Time) ([]*dto.StudentAttendanceResponse, error) {
	records, err := uc.studentAttendances.ListBySchoolAndDate(ctx, schoolID, date)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.StudentAttendanceResponse, len(records))
	for i, a := range records {
		out[i] = uc.toStudentAttendanceResponse(ctx, a)
	}
	return out, nil
}

func (uc *AttendanceService) ListStudentAttendanceByStudentAndRange(ctx context.Context, studentID string, from, to time.Time) ([]*dto.StudentAttendanceResponse, error) {
	records, err := uc.studentAttendances.ListByStudentAndRange(ctx, studentID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.StudentAttendanceResponse, len(records))
	for i, a := range records {
		out[i] = uc.toStudentAttendanceResponse(ctx, a)
	}
	return out, nil
}

func (uc *AttendanceService) ListStudentAttendanceBySchoolAndRange(ctx context.Context, schoolID string, from, to time.Time) ([]*dto.StudentAttendanceResponse, error) {
	records, err := uc.studentAttendances.ListBySchoolAndRange(ctx, schoolID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.StudentAttendanceResponse, len(records))
	for i, a := range records {
		out[i] = uc.toStudentAttendanceResponse(ctx, a)
	}
	return out, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────────────────────────────────────

func (uc *AttendanceService) toPersonnelAttendanceResponse(ctx context.Context, a *domain.PersonnelAttendance) *dto.PersonnelAttendanceResponse {
	resp := &dto.PersonnelAttendanceResponse{
		ID:            a.ID,
		PersonnelID:   a.PersonnelID,
		SchoolID:      a.SchoolID,
		AttendanceDate: a.AttendanceDate,
		Status:        string(a.Status),
		Remarks:       a.Remarks,
		RecordedBy:    a.RecordedBy,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
	if p, err := uc.personnelRepo.GetByID(ctx, a.PersonnelID); err == nil {
		resp.PersonnelName = p.FirstName + " " + p.LastName
	}
	return resp
}

func (uc *AttendanceService) toStudentAttendanceResponse(ctx context.Context, a *domain.StudentAttendance) *dto.StudentAttendanceResponse {
	resp := &dto.StudentAttendanceResponse{
		ID:            a.ID,
		StudentID:     a.StudentID,
		SchoolID:      a.SchoolID,
		AttendanceDate: a.AttendanceDate,
		Status:        string(a.Status),
		Remarks:       a.Remarks,
		RecordedBy:    a.RecordedBy,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
	if s, err := uc.studentRepo.GetByID(ctx, a.StudentID); err == nil {
		resp.StudentName = s.FirstName + " " + s.LastName
		resp.EnrollmentNo = s.EnrollmentNo
	}
	return resp
}
