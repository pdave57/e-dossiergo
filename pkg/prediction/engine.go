package prediction
// Package prediction implements the e-Dossier student performance prediction engine.
//
// MODEL DESIGN
// ─────────────
// The engine uses a weighted composite score built from three independent signal groups:
//
//   CompositeScore = (FacilityScore × 0.25) + (PersonnelScore × 0.35) + (HistoricalScore × 0.40)
//
// Weights rationale:
//   • Historical performance is the strongest predictor of future performance (0.40).
//   • Teaching quality (personnel) has a stronger direct effect than infrastructure (0.35).
//   • Facilities provide the enabling environment, not the direct cause (0.25).
//
// When no historical data exists (new school/student), weights auto-redistribute:
//   CompositeScore = (FacilityScore × 0.40) + (PersonnelScore × 0.60)
// and Confidence drops to reflect lower certainty.
//
// All intermediate scores are on a 0–100 scale.
// Output confidence is 0–1: 1.0 requires >= 3 terms of historical data.

import (
	"fmt"
	"math"
	"time"

	"github.com/edossier/api/internal/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// MODEL WEIGHTS  (all float64, must sum to 1.0 within each group)
// ─────────────────────────────────────────────────────────────────────────────

const (
	weightFacility   = 0.25
	weightPersonnel  = 0.35
	weightHistorical = 0.40

	// When no historical data → redistribute to facility + personnel only.
	weightFacilityNoHistory  = 0.40
	weightPersonnelNoHistory = 0.60

	// Minimum pass mark used in risk classification.
	passMark = 40.0

	// Score thresholds for risk levels.
	riskLowThreshold    = 60.0
	riskMediumThreshold = 40.0

	// School rating thresholds.
	ratingExcellentThreshold = 80.0
	ratingGoodThreshold      = 65.0
	ratingAverageThreshold   = 50.0

	// Facility condition weights (relative, summed per school then normalised).
	facilityWeightGood    = 1.0
	facilityWeightFair    = 0.6
	facilityWeightPoor    = 0.3
	facilityWeightDefunct = 0.0

	// Key facility bonus points (added to raw facility score).
	bonusLibrary    = 8.0
	bonusLab        = 8.0
	bonusICT        = 7.0
	bonusSportField = 4.0

	// Qualification score values for PersonnelSignal.
	qualScorePostgrad   = 15.0
	qualScoreQualified  = 10.0
	qualScoreUnqualified = 5.0

	// Ideal staff-to-student ratio (1 teacher per 30 students).
	idealRatio = 1.0 / 30.0
)

// Engine computes student and school performance predictions.
type Engine struct{}

// New returns a prediction Engine.
func New() *Engine { return &Engine{} }

// ─────────────────────────────────────────────────────────────────────────────
// FACILITY SCORE
// ─────────────────────────────────────────────────────────────────────────────

// FacilityScore computes a 0–100 score from a FacilitySignal.
// Formula: weighted condition score (0–70 base) + key-type bonuses (0–27) + small diversity bonus (0–3).
func (e *Engine) FacilityScore(sig *domain.FacilitySignal) (float64, []domain.Factor) {
	if sig == nil || sig.TotalFacilities == 0 {
		return 0, []domain.Factor{{
			Name:   "Facilities",
			Score:  0,
			Weight: weightFacility,
			Detail: "No facility records found for this school.",
		}}
	}

	// Weighted condition score — how many facilities are in good/fair/poor state.
	conditionWeighted := float64(sig.GoodCount)*facilityWeightGood +
		float64(sig.FairCount)*facilityWeightFair +
		float64(sig.PoorCount)*facilityWeightPoor +
		float64(sig.DefunctCount)*facilityWeightDefunct

	// Normalise to 0–70 against total facilities (all-good = 70).
	baseScore := (conditionWeighted / float64(sig.TotalFacilities)) * 70.0

	// Key facility bonuses.
	var bonus float64
	var details []string
	if sig.HasLibrary {
		bonus += bonusLibrary
		details = append(details, "library (+8)")
	}
	if sig.HasLab {
		bonus += bonusLab
		details = append(details, "laboratory (+8)")
	}
	if sig.HasICT {
		bonus += bonusICT
		details = append(details, "ICT centre (+7)")
	}
	if sig.HasSportField {
		bonus += bonusSportField
		details = append(details, "sport field (+4)")
	}

	// Diversity bonus: up to 3 points for having many different facility types.
	diversityBonus := math.Min(3.0, float64(sig.TotalFacilities)/5.0)

	total := clamp(baseScore+bonus+diversityBonus, 0, 100)

	detail := fmt.Sprintf(
		"%d facilities (%d good, %d fair, %d poor, %d defunct)",
		sig.TotalFacilities, sig.GoodCount, sig.FairCount, sig.PoorCount, sig.DefunctCount,
	)
	if len(details) > 0 {
		detail += "; key assets: " + joinStrings(details, ", ")
	}

	factors := []domain.Factor{
		{Name: "Facility Condition", Score: clamp(baseScore/70*100, 0, 100), Weight: 0.70, Detail: detail},
		{Name: "Key Facilities",     Score: clamp(bonus/27*100, 0, 100), Weight: 0.27, Detail: "Library, lab, ICT, sport"},
		{Name: "Facility Diversity", Score: clamp(diversityBonus/3*100, 0, 100), Weight: 0.03, Detail: "Variety of facility types"},
	}

	return total, factors
}

// ─────────────────────────────────────────────────────────────────────────────
// PERSONNEL SCORE
// ─────────────────────────────────────────────────────────────────────────────

// PersonnelScore computes a 0–100 score from a PersonnelSignal.
// Formula: qualification score (0–50) + ratio score (0–30) + activity score (0–20).
func (e *Engine) PersonnelScore(sig *domain.PersonnelSignal, enrolledStudents int) (float64, []domain.Factor) {
	if sig == nil || sig.TotalStaff == 0 {
		return 0, []domain.Factor{{
			Name:   "Personnel",
			Score:  0,
			Weight: weightPersonnel,
			Detail: "No personnel records found for this school.",
		}}
	}

	// 1. Qualification score: sum qual points, normalise to 0–50.
	qualPoints := float64(sig.PostgradCount)*qualScorePostgrad +
		float64(sig.QualifiedCount-sig.PostgradCount)*qualScoreQualified +
		float64(sig.ActiveStaff-sig.QualifiedCount)*qualScoreUnqualified
	maxQualPoints := float64(sig.ActiveStaff) * qualScorePostgrad // if all postgrad
	qualScore := 0.0
	if maxQualPoints > 0 {
		qualScore = clamp((qualPoints/maxQualPoints)*50.0, 0, 50)
	}

	// 2. Staff-to-student ratio score: 0–30.
	// Score 30 at ideal ratio (1:30), degrades towards 0 above 1:60 or below 1:15.
	ratioScore := 0.0
	if enrolledStudents > 0 && sig.ActiveStaff > 0 {
		actualRatio := float64(sig.TeacherCount) / float64(enrolledStudents)
		// Score peaks at ideal (1/30), falls off on either side.
		ratioScore = clamp(30.0*(1.0-math.Abs(actualRatio-idealRatio)/idealRatio), 0, 30)
	}

	// 3. Activity score: proportion of total staff that are active — 0–20.
	activityScore := 0.0
	if sig.TotalStaff > 0 {
		activityScore = (float64(sig.ActiveStaff) / float64(sig.TotalStaff)) * 20.0
	}

	total := clamp(qualScore+ratioScore+activityScore, 0, 100)

	factors := []domain.Factor{
		{
			Name:   "Staff Qualifications",
			Score:  clamp(qualScore/50*100, 0, 100),
			Weight: 0.50,
			Detail: fmt.Sprintf("%d postgrad, %d qualified of %d active staff", sig.PostgradCount, sig.QualifiedCount, sig.ActiveStaff),
		},
		{
			Name:   "Staff-to-Student Ratio",
			Score:  clamp(ratioScore/30*100, 0, 100),
			Weight: 0.30,
			Detail: fmt.Sprintf("%d teachers for %d students (ideal 1:30)", sig.TeacherCount, enrolledStudents),
		},
		{
			Name:   "Staff Activity",
			Score:  clamp(activityScore/20*100, 0, 100),
			Weight: 0.20,
			Detail: fmt.Sprintf("%d of %d staff active", sig.ActiveStaff, sig.TotalStaff),
		},
	}

	return total, factors
}

// ─────────────────────────────────────────────────────────────────────────────
// HISTORICAL SCORE
// ─────────────────────────────────────────────────────────────────────────────

// HistoricalScore converts a HistoricalSignal into a 0–100 score.
// Returns (score, confidence, factors).
// Confidence is low when TermsRecorded < 3.
func (e *Engine) HistoricalScore(sig *domain.HistoricalSignal) (score, confidence float64, factors []domain.Factor) {
	if sig == nil || sig.TermsRecorded == 0 {
		return 0, 0, []domain.Factor{{
			Name:   "Historical Performance",
			Score:  0,
			Weight: weightHistorical,
			Detail: "No historical score data available.",
		}}
	}

	// Average score component (0–50): linear scale where 100 raw → 50 pts.
	avgComponent := clamp(sig.AvgScore/100.0*50.0, 0, 50)

	// Pass rate component (0–30): 100% pass rate → 30 pts.
	passComponent := clamp(sig.PassRate*30.0, 0, 30)

	// Distinction rate component (0–20): 100% distinction rate → 20 pts.
	distComponent := clamp(sig.DistinctionRate*20.0, 0, 20)

	total := clamp(avgComponent+passComponent+distComponent, 0, 100)

	// Confidence: grows with terms of data; 3+ terms = full confidence.
	confidence = clamp(float64(sig.TermsRecorded)/3.0, 0, 1)

	factors = []domain.Factor{
		{
			Name:   "Average Score",
			Score:  clamp(avgComponent/50*100, 0, 100),
			Weight: 0.50,
			Detail: fmt.Sprintf("Historical average: %.1f%%", sig.AvgScore),
		},
		{
			Name:   "Pass Rate",
			Score:  clamp(sig.PassRate*100, 0, 100),
			Weight: 0.30,
			Detail: fmt.Sprintf("%.0f%% of results above pass mark (%.0f)", sig.PassRate*100, passMark),
		},
		{
			Name:   "Distinction Rate",
			Score:  clamp(sig.DistinctionRate*100, 0, 100),
			Weight: 0.20,
			Detail: fmt.Sprintf("%.0f%% of results above distinction mark (70)", sig.DistinctionRate*100),
		},
	}

	return total, confidence, factors
}

// ─────────────────────────────────────────────────────────────────────────────
// COMPOSITE SCORE → SCHOOL PREDICTION
// ─────────────────────────────────────────────────────────────────────────────

// SchoolPrediction builds a SchoolPrediction from the three signals.
func (e *Engine) SchoolPrediction(
	schoolID, schoolName string,
	facility *domain.FacilitySignal,
	personnel *domain.PersonnelSignal,
	historical *domain.HistoricalSignal,
	enrolledStudents int,
) domain.SchoolPrediction {
	facScore, facFactors := e.FacilityScore(facility)
	perScore, perFactors := e.PersonnelScore(personnel, enrolledStudents)
	histScore, confidence, histFactors := e.HistoricalScore(historical)

	var composite float64
	if confidence == 0 {
		// No history — use facility + personnel only.
		composite = facScore*weightFacilityNoHistory + perScore*weightPersonnelNoHistory
	} else {
		composite = facScore*weightFacility + perScore*weightPersonnel + histScore*weightHistorical
	}
	composite = clamp(composite, 0, 100)

	// Predicted pass rate: blend historical pass rate with composite score proxy.
	predictedPassRate := 0.0
	if historical != nil && historical.TermsRecorded > 0 {
		predictedPassRate = historical.PassRate*0.6 + (composite/100.0)*0.4
	} else {
		predictedPassRate = composite / 100.0
	}

	// Collect all factor cards.
	allFactors := append(append(facFactors, perFactors...), histFactors...)

	return domain.SchoolPrediction{
		SchoolID:          schoolID,
		SchoolName:        schoolName,
		Rating:            schoolRating(composite),
		CompositeScore:    round2(composite),
		FacilityScore:     round2(facScore),
		PersonnelScore:    round2(perScore),
		HistoricalScore:   round2(histScore),
		PredictedPassRate: round2(predictedPassRate * 100),
		Factors:           allFactors,
		GeneratedAt:       time.Now(),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// STUDENT PREDICTION
// ─────────────────────────────────────────────────────────────────────────────

// StudentPrediction builds a StudentPrediction by blending the school-level
// signals with the student's own historical signal.
//
// The student's own history overrides the school history component when available.
// School facility and personnel scores remain the same for every student in that school.
func (e *Engine) StudentPrediction(
	student domain.StudentRow,
	schoolID string,
	schoolComposite float64,
	facility *domain.FacilitySignal,
	personnel *domain.PersonnelSignal,
	studentHistorical *domain.HistoricalSignal,
	enrolledStudents int,
) domain.StudentPrediction {
	facScore, facFactors := e.FacilityScore(facility)
	perScore, perFactors := e.PersonnelScore(personnel, enrolledStudents)
	histScore, confidence, histFactors := e.HistoricalScore(studentHistorical)

	var composite float64
	if confidence == 0 {
		// No student history — fall back to school composite as baseline,
		// but still weight facility/personnel from the student's school context.
		composite = facScore*weightFacilityNoHistory + perScore*weightPersonnelNoHistory
		// Blend with school composite so students inherit school baseline.
		composite = composite*0.5 + schoolComposite*0.5
	} else {
		composite = facScore*weightFacility + perScore*weightPersonnel + histScore*weightHistorical
	}
	composite = clamp(composite, 0, 100)

	scoreRange := predictedRange(composite)
	allFactors := append(append(facFactors, perFactors...), histFactors...)

	return domain.StudentPrediction{
		StudentID:      student.ID,
		StudentName:    student.FirstName + " " + student.LastName,
		SchoolID:       schoolID,
		RiskLevel:      riskLevel(composite),
		PredictedRange: scoreRange,
		HistoricalAvg:  histAvg(studentHistorical),
		FacilityScore:  round2(facScore),
		PersonnelScore: round2(perScore),
		CompositeScore: round2(composite),
		Confidence:     round2(confidence),
		Factors:        allFactors,
		GeneratedAt:    time.Now(),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────────────────────────────────────

func riskLevel(composite float64) domain.RiskLevel {
	switch {
	case composite >= riskLowThreshold:
		return domain.RiskLow
	case composite >= riskMediumThreshold:
		return domain.RiskMedium
	default:
		return domain.RiskHigh
	}
}

func schoolRating(composite float64) domain.SchoolRating {
	switch {
	case composite >= ratingExcellentThreshold:
		return domain.RatingExcellent
	case composite >= ratingGoodThreshold:
		return domain.RatingGood
	case composite >= ratingAverageThreshold:
		return domain.RatingAverage
	default:
		return domain.RatingPoor
	}
}

func predictedRange(composite float64) domain.ScoreRange {
	// Map composite (0–100) to a predicted raw score range (0–100).
	// We add ±8 around the central estimate for the bracket.
	center := composite
	lo := clamp(center-8, 0, 100)
	hi := clamp(center+8, 0, 100)
	return domain.ScoreRange{
		Min:   round2(lo),
		Max:   round2(hi),
		Label: fmt.Sprintf("%.0f–%.0f%%", lo, hi),
	}
}

func histAvg(sig *domain.HistoricalSignal) float64 {
	if sig == nil {
		return 0
	}
	return round2(sig.AvgScore)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}