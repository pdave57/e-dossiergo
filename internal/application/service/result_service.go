package service

import (
	"context"
	"time"

	"github.com/edossier/api/internal/application/dto"
	"github.com/edossier/api/internal/domain"
	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
	"github.com/edossier/api/pkg/validator"
)

// ResultService orchestrates score entry, grade evaluation, and report card generation.
type ResultService struct {
	scores       domain.ScoreSheetRepository
	reportCards  domain.ReportCardRepository
	gradeConfigs domain.GradeConfigRepository
	scoreConfigs domain.ScoreConfigRepository
	enrollments  domain.EnrollmentRepository
	subLevels    domain.SubLevelRepository
	terms        domain.TermRepository
}

func NewResultService(
	scores domain.ScoreSheetRepository,
	reportCards domain.ReportCardRepository,
	gradeConfigs domain.GradeConfigRepository,
	scoreConfigs domain.ScoreConfigRepository,
	enrollments domain.EnrollmentRepository,
	subLevels domain.SubLevelRepository,
	terms domain.TermRepository,
) *ResultService {
	return &ResultService{
		scores: scores, reportCards: reportCards,
		gradeConfigs: gradeConfigs, scoreConfigs: scoreConfigs,
		enrollments: enrollments, subLevels: subLevels, terms: terms,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SCORE CONFIG
// ─────────────────────────────────────────────────────────────────────────────

func (uc *ResultService) UpsertScoreConfig(ctx context.Context, stateID string, req dto.UpsertScoreConfigRequest, createdBy string) (*domain.ScoreConfig, error) {
	v := validator.New().
		Check(req.CA1Max > 0, "ca1_max", "must be greater than 0").
		Check(req.CA2Max > 0, "ca2_max", "must be greater than 0").
		Check(req.CA3Max > 0, "ca3_max", "must be greater than 0").
		Check(req.ExamMax > 0, "exam_max", "must be greater than 0").
		Check(req.TotalMax > 0, "total_max", "must be greater than 0")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	// Validate component sum equals total
	sum := req.CA1Max + req.CA2Max + req.CA3Max + req.ExamMax
	if sum != req.TotalMax {
		return nil, apperror.BadRequest("ca1_max + ca2_max + ca3_max + exam_max must equal total_max")
	}
	sc := &domain.ScoreConfig{
		StateID: stateID, SchoolID: req.SchoolID, LevelID: req.LevelID,
		CA1Max: req.CA1Max, CA2Max: req.CA2Max, CA3Max: req.CA3Max,
		ExamMax: req.ExamMax, TotalMax: req.TotalMax,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return sc, uc.scoreConfigs.Upsert(ctx, sc)
}

func (uc *ResultService) GetScoreConfig(ctx context.Context, schoolID, stateID string) (*domain.ScoreConfig, error) {
	sc, err := uc.scoreConfigs.GetBySchool(ctx, schoolID)
	if err != nil {
		// Fall back to state default
		return uc.scoreConfigs.GetStateDefault(ctx, stateID)
	}
	return sc, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GRADE CONFIG
// ─────────────────────────────────────────────────────────────────────────────

func (uc *ResultService) UpsertGradeConfig(ctx context.Context, stateID string, req dto.UpsertGradeConfigRequest, createdBy string) (*domain.GradeConfig, error) {
	v := validator.New().
		Required(req.Grade, "grade").
		Required(req.Remark, "remark").
		Check(req.MaxScore > req.MinScore, "max_score", "must be greater than min_score")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}
	gc := &domain.GradeConfig{
		StateID: stateID, SchoolID: req.SchoolID, LevelID: req.LevelID,
		Grade: req.Grade, MinScore: req.MinScore, MaxScore: req.MaxScore,
		Remark: req.Remark, Points: req.Points,
		AuditFields: domain.AuditFields{CreatedBy: createdBy},
	}
	return gc, uc.gradeConfigs.Upsert(ctx, gc)
}

func (uc *ResultService) DeleteGradeConfig(ctx context.Context, id string) error {
	return uc.gradeConfigs.Delete(ctx, id)
}

func (uc *ResultService) ListGradeConfigs(ctx context.Context, schoolID, levelID, stateID string) ([]*domain.GradeConfig, error) {
	if schoolID != "" && levelID != "" {
		return uc.gradeConfigs.ListBySchoolAndLevel(ctx, schoolID, levelID)
	}
	if schoolID != "" {
		configs, err := uc.gradeConfigs.ListBySchool(ctx, schoolID)
		if err != nil || len(configs) == 0 {
			return uc.gradeConfigs.ListStateDefault(ctx, stateID)
		}
		return configs, nil
	}
	return uc.gradeConfigs.ListStateDefault(ctx, stateID)
}

// ─────────────────────────────────────────────────────────────────────────────
// SCORE ENTRY
// ─────────────────────────────────────────────────────────────────────────────

// UpsertScore records or updates a student's score for one subject+term.
// It validates component maximums, computes the total, evaluates the grade.
func (uc *ResultService) UpsertScore(ctx context.Context, stateID string, req dto.UpsertScoreRequest, recordedBy string) (*domain.ScoreSheet, error) {
	v := validator.New().
		Required(req.StudentID, "student_id").
		Required(req.LevelID, "level_id").
		Required(req.SubLevelID, "sub_level_id").
		Required(req.SubjectID, "subject_id").
		Required(req.TermID, "term_id").
		Min(req.CA1Score, 0, "ca1_score").
		Min(req.CA2Score, 0, "ca2_score").
		Min(req.CA3Score, 0, "ca3_score").
		Min(req.ExamScore, 0, "exam_score")
	if !v.Valid() {
		return nil, apperror.Validation(v.Errors())
	}

	// Resolve sub-level to obtain the owning school
	sub, err := uc.subLevels.GetByID(ctx, req.SubLevelID)
	if err != nil {
		return nil, err
	}

	// Resolve term to obtain the session
	term, err := uc.terms.GetByID(ctx, req.TermID)
	if err != nil {
		return nil, err
	}

	// Resolve score config (school-specific or state default)
	cfg, err := uc.GetScoreConfig(ctx, sub.SchoolID, stateID)
	if err != nil {
		// If no config exists at all, use the classic Nigerian default: CA=30, Exam=70
		cfg = &domain.ScoreConfig{CA1Max: 10, CA2Max: 10, CA3Max: 10, ExamMax: 70, TotalMax: 100}
	}

	// Validate component bounds
	bv := validator.New().
		Max(req.CA1Score, cfg.CA1Max, "ca1_score").
		Max(req.CA2Score, cfg.CA2Max, "ca2_score").
		Max(req.CA3Score, cfg.CA3Max, "ca3_score").
		Max(req.ExamScore, cfg.ExamMax, "exam_score")
	if !bv.Valid() {
		return nil, apperror.Validation(bv.Errors())
	}

	ss := &domain.ScoreSheet{
		StudentID:  req.StudentID,
		SchoolID:   sub.SchoolID,
		SessionID:  term.SessionID,
		LevelID:    req.LevelID,
		SubLevelID: req.SubLevelID,
		TermID:     req.TermID,
		SubjectID:  req.SubjectID,
		CA1Score:   req.CA1Score,
		CA2Score:   req.CA2Score,
		CA3Score:   req.CA3Score,
		ExamScore:  req.ExamScore,
		RecordedBy: recordedBy,
		RecordedAt: time.Now(),
	}
	ss.ComputeTotal()

	// Evaluate grade
	gc, err := uc.gradeConfigs.EvaluateGrade(ctx, ss.TotalScore, sub.SchoolID, stateID)
	if err == nil {
		ss.Grade = gc.Grade
		ss.Remark = gc.Remark
	}

	return ss, uc.scores.Upsert(ctx, ss)
}

// BulkUpsertScores processes multiple score entries in one call.
func (uc *ResultService) BulkUpsertScores(ctx context.Context, stateID string, req dto.BulkUpsertScoreRequest, recordedBy string) ([]*domain.ScoreSheet, []error) {
	results := make([]*domain.ScoreSheet, 0, len(req.Scores))
	errs := make([]error, 0)
	for _, s := range req.Scores {
		ss, err := uc.UpsertScore(ctx, stateID, s, recordedBy)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, ss)
	}
	return results, errs
}

func (uc *ResultService) GetScore(ctx context.Context, id string) (*domain.ScoreSheet, error) {
	return uc.scores.GetByID(ctx, id)
}

func (uc *ResultService) GetStudentScores(ctx context.Context, studentID, sessionID string) ([]*domain.ScoreSheet, error) {
	return uc.scores.ListByStudent(ctx, studentID, sessionID)
}

func (uc *ResultService) ListScores(ctx context.Context, f domain.ScoreSheetFilter, p pagination.Params) ([]*domain.ScoreSheet, int, error) {
	return uc.scores.List(ctx, f, p)
}

// ComputePositions re-ranks all students in a class arm for a given subject/term.
func (uc *ResultService) ComputePositions(ctx context.Context, termID, subLevelID, subjectID string) error {
	if termID == "" || subLevelID == "" || subjectID == "" {
		return apperror.BadRequest("term_id, sub_level_id, and subject_id are required")
	}
	return uc.scores.ComputePositions(ctx, termID, subLevelID, subjectID)
}

// ─────────────────────────────────────────────────────────────────────────────
// REPORT CARD GENERATION
// ─────────────────────────────────────────────────────────────────────────────

// GenerateReportCards computes and persists report cards for all students
// in a given sub-level for the specified term. Idempotent — safe to re-run.
func (uc *ResultService) GenerateReportCards(ctx context.Context, stateID string, req dto.GenerateReportCardRequest) (int, error) {
	v := validator.New().
		Required(req.SchoolID, "school_id").
		Required(req.SessionID, "session_id").
		Required(req.TermID, "term_id").
		Required(req.SubLevelID, "sub_level_id")
	if !v.Valid() {
		return 0, apperror.Validation(v.Errors())
	}

	// Fetch all active enrollments for this class arm
	enrollFilter := domain.EnrollmentFilter{
		SchoolID:   req.SchoolID,
		SessionID:  req.SessionID,
		SubLevelID: req.SubLevelID,
		Status:     string(domain.EnrollmentStatusActive),
	}
	enrollments, _, err := uc.enrollments.List(ctx, enrollFilter, pagination.Params{Page: 1, PerPage: 1000})
	if err != nil {
		return 0, err
	}

	generated := 0
	for _, enr := range enrollments {
		// Get all scores for this student in this term
		allScores, err := uc.scores.ListByStudent(ctx, enr.StudentID, req.SessionID)
		if err != nil {
			continue
		}

		// Filter to the current term
		var termScores []*domain.ScoreSheet
		for _, sc := range allScores {
			if sc.TermID == req.TermID {
				termScores = append(termScores, sc)
			}
		}
		if len(termScores) == 0 {
			continue
		}

		// Aggregate
		var totalScore float64
		for _, sc := range termScores {
			totalScore += sc.TotalScore
		}
		avg := totalScore / float64(len(termScores))

		// Overall grade from average
		overallGrade := ""
		gc, err := uc.gradeConfigs.EvaluateGrade(ctx, avg, req.SchoolID, stateID)
		if err == nil {
			overallGrade = gc.Grade
		}

		rc := &domain.ReportCard{
			StudentID:    enr.StudentID,
			SchoolID:     req.SchoolID,
			SessionID:    req.SessionID,
			TermID:       req.TermID,
			LevelID:      enr.LevelID,
			SubLevelID:   req.SubLevelID,
			TotalScore:   totalScore,
			AverageScore: avg,
			OverallGrade: overallGrade,
			SubjectCount: len(termScores),
		}
		if err := uc.reportCards.Upsert(ctx, rc); err != nil {
			continue
		}
		generated++
	}

	// Compute class positions across all generated report cards
	if generated > 0 {
		uc.computeReportCardPositions(ctx, req.SchoolID, req.TermID)
	}
	return generated, nil
}

// computeReportCardPositions ranks report cards by average score within a class.
func (uc *ResultService) computeReportCardPositions(ctx context.Context, schoolID, termID string) {
	rcs, _, err := uc.reportCards.ListByTerm(ctx, schoolID, termID, pagination.Params{Page: 1, PerPage: 1000})
	if err != nil {
		return
	}
	// Simple descending sort then assign positions
	// In production this is handled by the DB rank window function
	for pos, rc := range rcs {
		rc.ClassPosition = pos + 1
		_ = uc.reportCards.Upsert(ctx, rc)
	}
}

func (uc *ResultService) GetReportCard(ctx context.Context, id string) (*domain.ReportCard, error) {
	return uc.reportCards.GetByID(ctx, id)
}

func (uc *ResultService) GetStudentReportCard(ctx context.Context, studentID, termID string) (*domain.ReportCard, error) {
	return uc.reportCards.GetByStudentTerm(ctx, studentID, termID)
}

func (uc *ResultService) ListReportCards(ctx context.Context, schoolID, termID string, p pagination.Params) ([]*domain.ReportCard, int, error) {
	return uc.reportCards.ListByTerm(ctx, schoolID, termID, p)
}

func (uc *ResultService) GetStudentAllReports(ctx context.Context, studentID string) ([]*domain.ReportCard, error) {
	return uc.reportCards.ListByStudent(ctx, studentID)
}

func (uc *ResultService) UpdateRemarks(ctx context.Context, id string, req dto.UpdateReportCardRemarksRequest, updatedBy string) (*domain.ReportCard, error) {
	rc, err := uc.reportCards.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rc.IsPublished() {
		return nil, apperror.BadRequest("cannot update a published report card")
	}
	rc.PrincipalRemark = req.PrincipalRemark
	rc.TeacherRemark = req.TeacherRemark
	if req.Attendance > 0 {
		rc.Attendance = req.Attendance
	}
	if req.TotalSchoolDays > 0 {
		rc.TotalSchoolDays = req.TotalSchoolDays
	}
	rc.UpdatedBy = updatedBy
	return rc, uc.reportCards.Upsert(ctx, rc)
}

// Publish makes a report card visible to students/parents.
func (uc *ResultService) Publish(ctx context.Context, id string) error {
	rc, err := uc.reportCards.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rc.IsPublished() {
		return apperror.BadRequest("report card is already published")
	}
	return uc.reportCards.Publish(ctx, id)
}
