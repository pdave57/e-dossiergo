package service

import (
	"context"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/internal/infrastructure/ml"
)

type RecommendationService struct {
	repo      domain.RecommendationRepository
	mlClient  *ml.RecommenderClient
}

func NewRecommendationService(repo domain.RecommendationRepository, mlClient *ml.RecommenderClient) *RecommendationService {
	return &RecommendationService{
		repo:     repo,
		mlClient: mlClient,
	}
}

func (uc *RecommendationService) GetRecommendations(ctx context.Context) (*ml.RecommendResponse, error) {
	schools, err := uc.repo.ListSchoolsWithAggregates(ctx)
	if err != nil {
		return nil, err
	}

	inputs := make([]ml.SchoolInput, len(schools))
	for i, s := range schools {
		inputs[i] = ml.SchoolInput{
			ID:                   i + 1,
			Name:                 s.Name,
			Zone:                 s.ZoneName,
			LGA:                  s.LGAName,
			LevelType:            s.Category,
			TotalTeachers:        s.TotalTeachers,
			QualifiedTeachers:    s.QualifiedTeachers,
			TotalStudents:        s.TotalStudents,
			TotalClassrooms:      s.TotalClassrooms,
			FunctionalClassrooms: s.FunctionalClassrooms,
			HasLibrary:           s.HasLibrary,
			HasLaboratory:        s.HasLaboratory,
			HasToilet:            s.HasToilet,
			HasElectricity:       s.HasElectricity,
			HasWater:             s.HasWater,
			HasInternet:          s.HasInternet,
			SubjectsOffered:      s.SubjectsOffered,
			ExpectedSubjects:     s.ExpectedSubjects,
			BooksPerStudent:      s.BooksPerStudent,
			AvgPassRate:          s.AvgPassRate,
		}
	}

	return uc.mlClient.Recommend(ctx, inputs)
}
