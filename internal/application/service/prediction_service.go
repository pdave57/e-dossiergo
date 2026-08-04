// Package Service — PredictionService orchestrates the prediction engine.
package service

import (
	"fmt"

	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/prediction"
)

// PredictionService generates performance prediction reports.
type PredictionService struct {
	repo   domain.PredictionRepository
	engine *prediction.Engine
}

// NewPredictionService constructs a PredictionService.
func NewPredictionService(repo domain.PredictionRepository) *PredictionService {
	return &PredictionService{
		repo:   repo,
		engine: prediction.New(),
	}
}

// GenerateReport produces a full PredictionReport for one school.
// sessionID is optional — if empty, all active enrollments are used.
func (uc *PredictionService) GenerateReport(schoolID, sessionID string) (*domain.PredictionReport, error) {
	// ── 1. Fetch all signals in parallel-friendly sequence ────────────────────

	schoolName, err := uc.repo.GetSchoolName(schoolID)
	if err != nil {
		return nil, fmt.Errorf("prediction: school lookup: %w", err)
	}

	facilitySig, err := uc.repo.GetFacilitySignal(schoolID)
	if err != nil {
		return nil, fmt.Errorf("prediction: facility signal: %w", err)
	}

	personnelSig, err := uc.repo.GetPersonnelSignal(schoolID)
	if err != nil {
		return nil, fmt.Errorf("prediction: personnel signal: %w", err)
	}

	schoolHistSig, err := uc.repo.GetSchoolHistoricalSignal(schoolID)
	if err != nil {
		return nil, fmt.Errorf("prediction: school history: %w", err)
	}

	enrolledCount, err := uc.repo.GetEnrollmentCount(schoolID)
	if err != nil {
		return nil, fmt.Errorf("prediction: enrollment count: %w", err)
	}

	students, err := uc.repo.GetEnrolledStudents(schoolID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("prediction: enrolled students: %w", err)
	}

	// ── 2. School-level prediction ────────────────────────────────────────────

	schoolPred := uc.engine.SchoolPrediction(
		schoolID, schoolName,
		facilitySig, personnelSig, schoolHistSig,
		enrolledCount,
	)

	// ── 3. Per-student predictions ────────────────────────────────────────────

	studentPreds := make([]domain.StudentPrediction, 0, len(students))
	var highRisk, medRisk, lowRisk int

	for _, student := range students {
		studentHist, err := uc.repo.GetStudentHistoricalSignal(student.ID, schoolID)
		if err != nil {
			// Non-fatal: use zero signal for students with no score data yet.
			studentHist = &domain.HistoricalSignal{StudentID: student.ID, SchoolID: schoolID}
		}

		sp := uc.engine.StudentPrediction(
			student,
			schoolID,
			schoolPred.CompositeScore,
			facilitySig,
			personnelSig,
			studentHist,
			enrolledCount,
		)
		studentPreds = append(studentPreds, sp)

		switch sp.RiskLevel {
		case domain.RiskHigh:
			highRisk++
		case domain.RiskMedium:
			medRisk++
		default:
			lowRisk++
		}
	}

	// Back-fill risk counts into school prediction.
	schoolPred.HighRiskCount = highRisk
	schoolPred.MediumRiskCount = medRisk
	schoolPred.LowRiskCount = lowRisk

	return &domain.PredictionReport{
		School:   schoolPred,
		Students: studentPreds,
	}, nil
}

// DTO mirror for the HTTP handler.

// SchoolOnlyReport generates a school-level-only prediction (faster, no per-student loop).
func (uc *PredictionService) SchoolOnlyReport(schoolID string) (*domain.SchoolPrediction, error) {
	schoolName, err := uc.repo.GetSchoolName(schoolID)
	if err != nil {
		return nil, err
	}
	facilitySig, err := uc.repo.GetFacilitySignal(schoolID)
	if err != nil {
		return nil, err
	}
	personnelSig, err := uc.repo.GetPersonnelSignal(schoolID)
	if err != nil {
		return nil, err
	}
	schoolHistSig, err := uc.repo.GetSchoolHistoricalSignal(schoolID)
	if err != nil {
		return nil, err
	}
	enrolledCount, err := uc.repo.GetEnrollmentCount(schoolID)
	if err != nil {
		return nil, err
	}

	pred := uc.engine.SchoolPrediction(
		schoolID, schoolName,
		facilitySig, personnelSig, schoolHistSig,
		enrolledCount,
	)
	return &pred, nil
}