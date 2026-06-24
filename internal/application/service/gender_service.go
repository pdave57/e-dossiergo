package service
import (
	"context"
	"github.com/edossier/api/internal/domain"
)
type GenderService interface {
	CountByGender(ctx context.Context, stateID string) (map[string]int, error)
}

type genderService struct {
	studentRepo domain.StudentRepository
}
func NewGenderService(studentRepo domain.StudentRepository) GenderService {
	return &genderService{studentRepo: studentRepo}
}
func (s *genderService) CountByGender(ctx context.Context, stateID string) (map[string]int, error) {
	male, female, other, err := s.studentRepo.CountByGender(ctx, stateID)
	if err != nil {
		return nil, err
	}
	return map[string]int{
		"male":   male,
		"female": female,
		"other":  other,
	}, nil
}