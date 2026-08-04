// internal/application/service/avatar_service.go
package service

import (
	"context"
	"io"
	"github.com/edossier/api/internal/domain"
)

type AvatarService struct {
	personnelRepo domain.PersonnelRepository
	studentRepo   domain.StudentRepository
	storage       domain.ImageStorage
}

func NewAvatarService(pRepo domain.PersonnelRepository, sRepo domain.StudentRepository, storage domain.ImageStorage) *AvatarService {
	return &AvatarService{personnelRepo: pRepo, studentRepo: sRepo, storage: storage}
}

func (s *AvatarService) UploadPersonnelAvatar(ctx context.Context, personnelID string, file io.Reader, filename string) (string, error) {
	url, publicID, err := s.storage.Upload(ctx, file, filename, "personnel_avatars")
	if err != nil {
		return "", err
	}

	err = s.personnelRepo.UpdateAvatar(ctx, personnelID, url)
	if err != nil {
		s.storage.Delete(ctx, publicID)
		return "", err
	}

	return url, nil
}

func (s *AvatarService) UploadStudentAvatar(ctx context.Context, studentID string, file io.Reader, filename string) (string, error) {
	url, _, err := s.storage.Upload(ctx, file, filename, "student_avatars")
	if err != nil {
		return "", err
	}

	err = s.studentRepo.UpdateAvatar(ctx, studentID, url)
	if err != nil {
		return "", err
	}

	return url, nil
}