package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/token"
)

type StudentAuthService struct {
	students      domain.StudentRepository
	schools       domain.SchoolRepository
	refreshTokens domain.RefreshTokenRepository
	tokenMaker    *token.Maker
}

func NewStudentAuthService(
	students domain.StudentRepository,
	schools domain.SchoolRepository,
	refreshTokens domain.RefreshTokenRepository,
	tokenMaker *token.Maker,
) *StudentAuthService {
	return &StudentAuthService{
		students:      students,
		schools:       schools,
		refreshTokens: refreshTokens,
		tokenMaker:    tokenMaker,
	}
}

func (uc *StudentAuthService) Login(ctx context.Context, req dto.StudentLoginRequest) (*dto.StudentLoginResponse, error) {
	if req.SchoolCode == "" || req.EnrollmentNo == "" {
		return nil, apperror.Validation(map[string]any{
			"school_code":   "is required",
			"enrollment_no": "is required",
		})
	}

	school, err := uc.schools.GetByCode(ctx, req.SchoolCode)
	if err != nil {
		return nil, apperror.Unauthorized("invalid school code or enrollment number")
	}

	student, err := uc.students.GetByEnrollmentNo(ctx, req.EnrollmentNo)
	if err != nil {
		return nil, apperror.Unauthorized("invalid school code or enrollment number")
	}

	if student.SchoolID != school.ID {
		return nil, apperror.Unauthorized("invalid school code or enrollment number")
	}

	if student.Status != domain.StudentStatusActive {
		return nil, apperror.Unauthorized("student account is not active")
	}

	accessTok, exp, err := uc.tokenMaker.CreateAccessToken(student.ID, student.StateID, student.SchoolID, []string{"STUDENT"})
	if err != nil {
		return nil, apperror.Internal(err)
	}

	refreshTok, refreshExp, err := uc.tokenMaker.CreateRefreshToken(student.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(refreshTok)))
	_ = uc.refreshTokens.Save(ctx, &domain.RefreshToken{
		UserID:    student.ID,
		TokenHash: hash,
		ExpiresAt: refreshExp,
	})

	return &dto.StudentLoginResponse{
		AccessToken:  accessTok,
		RefreshToken: refreshTok,
		ExpiresAt:    exp,
		Student: dto.StudentInfo{
			ID:           student.ID,
			StateID:      student.StateID,
			SchoolID:     student.SchoolID,
			EnrollmentNo: student.EnrollmentNo,
			FirstName:    student.FirstName,
			MiddleName:   student.MiddleName,
			LastName:     student.LastName,
			Gender:       string(student.Gender),
			Status:       string(student.Status),
		},
	}, nil
}
