// Package router wires all routes, middleware, and handlers into a single chi.Router.
package router

import (
	"net/http"

	"log/slog"

	chi "github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	swagger "github.com/swaggo/http-swagger/v2"

	"github.com/edossier/api/internal/interfaces/http/handler"
	"github.com/edossier/api/internal/interfaces/http/middleware"
	"github.com/edossier/api/pkg/token"
)

// Deps bundles all handler and middleware dependencies.
type Deps struct {
	Log         *slog.Logger
	TokenMaker  *token.Maker
	RoleChecker middleware.UserRoleChecker

	Auth           *handler.AuthHandler
	User           *handler.UserHandler
	Role           *handler.RoleHandler
	Zone           *handler.ZoneHandler
	State          *handler.StateHandler
	LGA            *handler.LGAHandler
	School         *handler.SchoolHandler
	Academic       *handler.AcademicHandler
	Gender         *handler.GenderHandler
	Level          *handler.LevelHandler
	Subject        *handler.SubjectHandler
	Personnel      *handler.PersonnelHandler
	Student        *handler.StudentHandler
	Result         *handler.ResultHandler
	Report         *handler.ReportHandler
	ZonalSummary   *handler.ZonalSummaryHandler
	Avatar         *handler.AvatarHandler
	StudentAuth    *handler.StudentAuthHandler
	Attendance     *handler.AttendanceHandler
	Prediction     *handler.PredictionHandler
	Recommendation *handler.RecommendationHandler
}

// New builds and returns the fully-configured HTTP router.
//
//	@Summary		Test router
//	@Description	Test description
//	@Tags			test
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/test-router [get]
func New(d Deps) http.Handler {
	r := chi.NewRouter()

	// ── Global middleware stack ───────────────────────────────────────────────
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(d.Log))
	r.Use(middleware.Recovery(d.Log))
	r.Use(middleware.CORS([]string{"*"})) // restrict in production

	// ── Health check (unauthenticated) ────────────────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"e-dossier"}`)) //nolint:errcheck
	})

	// ── Swagger UI ───────────────────────────────────────────────────────────
	r.Get("/swagger/*", swagger.Handler(swagger.URL("/swagger/doc.json")))

	// ── API v1 ────────────────────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {

		// ── PUBLIC REPORTS ────────────────────────────────────────────────────
		r.Route("/reports/public", func(r chi.Router) {
			r.Get("/teaching-personnel", d.Report.GetPublicTeachingPersonnel)
		})
		r.Route("/reports/gender", func(r chi.Router) {
			r.Get("/total", d.Gender.CountByGender)
		})

		r.Route("/reports/students", func(r chi.Router) {
			r.Get("/total", d.Student.CountTotalStudents)
		})

		r.Route("/reports/personnel", func(r chi.Router) {
			r.Get("/total", d.Personnel.CountTotalPersonnel)
		})

		r.Route("/reports/schools", func(r chi.Router) {
			r.Get("/total", d.School.CountTotalSchools)
		})
		r.Route("/reports/dashboard", func(r chi.Router) {
			r.With(middleware.Authenticate(d.TokenMaker)).Get("/stats", d.Report.GetDashboardStats)
		})

		r.Route("/reports/zonal", func(r chi.Router) {
			r.With(middleware.Authenticate(d.TokenMaker)).Get("/summary", d.ZonalSummary.GetZoneSummaryReport)
		})

		// ── AUTH (public) ─────────────────────────────────────────────────────
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", d.Auth.Register)
			r.Post("/login", d.Auth.Login)
			r.Post("/refresh", d.Auth.Refresh)
			r.Post("/student-login", d.StudentAuth.Login)
		})

		// ── GEO: STATES (public reads) ──────────────────────────────────────
		r.Get("/states", d.State.ListStates)
		r.Get("/states/{id}", d.State.GetState)

		// Zones and LGAs — public list reads only
		r.Get("/states/{stateId}/zones", d.Zone.ListZones)
		r.Get("/states/{stateId}/lgas", d.LGA.ListLGAs)

		// ── SCHOOLS (public read) ───────────────────────────────────────────
		r.Route("/schools", func(r chi.Router) {
			r.Get("/", d.School.List)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "create")).Post("/", d.School.Create)
			r.Get("/{id}", d.School.GetByID)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Put("/{id}", d.School.Update)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Put("/{id}/logo", d.School.UploadLogo)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "delete")).Delete("/{id}", d.School.Delete)

			// Facilities
			r.Get("/{id}/facilities", d.School.ListFacilities)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "create")).Post("/{id}/facilities", d.School.AddFacility)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Put("/{id}/facilities/{facilityId}", d.School.UpdateFacility)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "delete")).Delete("/{id}/facilities/{facilityId}", d.School.DeleteFacility)

			// School-level assignments
			r.Get("/{schoolId}/levels", d.Level.ListSchoolLevels)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Post("/{schoolId}/levels", d.Level.UpsertSchoolLevel)

			// Sub-levels (class arms)
			r.Get("/{schoolId}/sub-levels", func(w http.ResponseWriter, r *http.Request) {
				// delegates to level handler with schoolId from path
				d.Level.ListSubLevels(w, r)
			})
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Post("/{schoolId}/sub-levels", d.Level.CreateSubLevel)

			// School subjects
			r.Get("/{schoolId}/subjects", d.Subject.ListSchoolSubjects)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Post("/{schoolId}/subjects", d.Subject.AssignToSchool)
		})

		r.Route("/sub-levels", func(r chi.Router) {
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Post("/", d.Level.CreateSubLevelGlobal)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Put("/{id}", d.Level.UpdateSubLevel)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Delete("/{id}", d.Level.DeleteSubLevel)
		})

		r.Route("/school-subjects", func(r chi.Router) {
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Put("/{id}", d.Subject.UpdateSchoolSubject)
			r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Delete("/{id}", d.Subject.RemoveSchoolSubject)
		})

		// ── AUTHENTICATED routes ──────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(d.TokenMaker))

			// Current user
			r.Get("/auth/me", d.Auth.Me)
			r.Post("/auth/logout", d.Auth.Logout)
			r.Post("/auth/change-password", d.Auth.ChangePassword)

			// ── GEO: STATES / ZONES / LGAS (authenticated writes) ────────────
			r.With(authorize(d, "zones", "create")).Post("/states", d.State.CreateState)
			r.With(authorize(d, "zones", "update")).Put("/states/{id}", d.State.UpdateState)

			r.With(authorize(d, "zones", "create")).Post("/states/{stateId}/zones", d.Zone.CreateZone)
			r.With(authorize(d, "zones", "update")).Put("/zones/{id}", d.Zone.UpdateZone)
			r.With(authorize(d, "zones", "delete")).Delete("/zones/{id}", d.Zone.DeleteZone)

			r.With(authorize(d, "lgas", "create")).Post("/states/{stateId}/lgas", d.LGA.CreateLGA)
			r.With(authorize(d, "lgas", "update")).Put("/lgas/{id}", d.LGA.UpdateLGA)
			r.With(authorize(d, "lgas", "delete")).Delete("/lgas/{id}", d.LGA.DeleteLGA)

			// ── USERS ─────────────────────────────────────────────────────────
			r.Route("/users", func(r chi.Router) {
				r.With(authorize(d, "users", "read")).Get("/", d.User.List)
				r.With(authorize(d, "users", "read")).Get("/{id}", d.User.GetByID)
				r.With(authorize(d, "users", "update")).Put("/{id}", d.User.Update)
				r.With(authorize(d, "users", "delete")).Delete("/{id}", d.User.Delete)

				// role assignment
				r.With(authorize(d, "roles", "update")).Post("/{id}/roles", d.User.AssignRole)
				r.With(authorize(d, "roles", "update")).Delete("/{id}/roles/{roleId}", d.User.RevokeRole)
				r.With(authorize(d, "roles", "read")).Get("/{id}/roles", d.User.GetRoles)
			})

			// ── ROLES & PERMISSIONS ───────────────────────────────────────────
			r.Route("/roles", func(r chi.Router) {
				r.With(authorize(d, "roles", "read")).Get("/", d.Role.List)
				r.With(authorize(d, "roles", "create")).Post("/", d.Role.Create)
				r.With(authorize(d, "roles", "read")).Get("/{id}", d.Role.GetByID)
				r.With(authorize(d, "roles", "update")).Put("/{id}", d.Role.Update)
				r.With(authorize(d, "roles", "delete")).Delete("/{id}", d.Role.Delete)
				r.With(authorize(d, "roles", "update")).Post("/{id}/permissions", d.Role.AddPermission)
				r.With(authorize(d, "roles", "update")).Delete("/{id}/permissions/{permId}", d.Role.RemovePermission)
			})
			r.With(authorize(d, "roles", "read")).Get("/permissions", d.Role.ListPermissions)

			// ── ACADEMIC SESSIONS ─────────────────────────────────────────────
			r.Route("/sessions", func(r chi.Router) {
				r.With(authorize(d, "sessions", "read")).Get("/", d.Academic.ListSessions)
				r.With(authorize(d, "sessions", "read")).Get("/active", d.Academic.GetActiveSession)
				r.With(authorize(d, "sessions", "create")).Post("/", d.Academic.CreateSession)
				r.With(authorize(d, "sessions", "read")).Get("/{id}", d.Academic.GetSession)
				r.With(authorize(d, "sessions", "update")).Put("/{id}", d.Academic.UpdateSession)
				r.With(authorize(d, "sessions", "update")).Post("/{id}/activate", d.Academic.ActivateSession)
				r.With(authorize(d, "sessions", "update")).Delete("/{id}", d.Academic.DeleteSession)

				// Terms nested under sessions
				r.With(authorize(d, "sessions", "read")).Get("/{sessionId}/terms", d.Academic.ListTerms)
				r.With(authorize(d, "sessions", "create")).Post("/{sessionId}/terms", d.Academic.CreateTerm)
				r.With(authorize(d, "sessions", "update")).Put("/{sessionId}/terms/{id}", d.Academic.UpdateTerm)
				r.With(authorize(d, "sessions", "update")).Post("/{sessionId}/terms/{id}/activate", d.Academic.ActivateTerm)
				r.With(authorize(d, "sessions", "update")).Delete("/{sessionId}/terms/{id}", d.Academic.DeleteTerm)
			})

			// ── TERMS (top-level; owning session supplied in request body) ──
			r.Route("/terms", func(r chi.Router) {
				r.With(authorize(d, "sessions", "read")).Get("/", d.Academic.ListAllTerms)
				r.With(authorize(d, "sessions", "create")).Post("/", d.Academic.CreateTermTopLevel)
				r.With(authorize(d, "sessions", "read")).Get("/{id}", d.Academic.GetTerm)
				r.With(authorize(d, "sessions", "update")).Put("/{id}", d.Academic.UpdateTerm)
				r.With(authorize(d, "sessions", "update")).Delete("/{id}", d.Academic.DeleteTerm)
				r.With(authorize(d, "sessions", "read")).Get("/active", d.Academic.GetActiveTerm)
			})

			// ── LEVELS (state-wide class definitions) ─────────────────────────
			r.Route("/levels", func(r chi.Router) {
				r.Get("/", d.Level.List)
				r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "create")).Post("/", d.Level.Create)
				r.Get("/{id}", d.Level.GetByID)
				r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Put("/{id}", d.Level.Update)
				r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "delete")).Delete("/{id}", d.Level.Delete)

				// Sub-levels listed by level
				r.Get("/{levelId}/sub-levels", d.Level.ListSubLevels)
			})

			// ── SUBJECTS (state-wide) ─────────────────────────────────────────
			r.Route("/subjects", func(r chi.Router) {
				r.Get("/", d.Subject.List)
				r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "create")).Post("/", d.Subject.Create)
				r.Get("/{id}", d.Subject.GetByID)
				r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "update")).Put("/{id}", d.Subject.Update)
				r.With(middleware.Authenticate(d.TokenMaker), authorize(d, "schools", "delete")).Delete("/{id}", d.Subject.Delete)
			})

			// ── PERSONNEL ─────────────────────────────────────────────────────
			r.Route("/personnel", func(r chi.Router) {
				r.With(authorize(d, "personnel", "read")).Get("/", d.Personnel.List)
				r.With(authorize(d, "personnel", "create")).Post("/", d.Personnel.Create)

				r.With(authorize(d, "personnel", "read")).Get("/{id}", d.Personnel.GetByID)
				r.With(authorize(d, "personnel", "update")).Put("/{id}", d.Personnel.Update)
				r.With(authorize(d, "personnel", "delete")).Delete("/{id}", d.Personnel.Delete)
				r.With(authorize(d, "personnel", "update")).Post("/{id}/transfer", d.Personnel.Transfer)
				r.With(authorize(d, "personnel", "read")).Get("/{id}/transfers", d.Personnel.ListTransfers)
			})

			// ── STUDENTS ──────────────────────────────────────────────────────
			r.Route("/students", func(r chi.Router) {

				r.With(authorize(d, "students", "read")).Get("/", d.Student.List)
				r.With(authorize(d, "students", "create")).Post("/", d.Student.Create)
				r.With(authorize(d, "students", "read")).Get("/{id}", d.Student.GetByID)
				r.With(authorize(d, "students", "update")).Put("/{id}", d.Student.Update)
				r.With(authorize(d, "students", "delete")).Delete("/{id}", d.Student.Delete)
				r.With(authorize(d, "students", "read")).Get("/next-serial", d.Student.GetNextSerial)

				// Progressions (level advancement)
				r.With(authorize(d, "enrollments", "update")).Post("/{id}/progressions", d.Student.RecordProgression)
				r.With(authorize(d, "enrollments", "read")).Get("/{id}/progressions", d.Student.ListProgressions)

				// Report cards for a student
				r.With(authorize(d, "results", "read")).Get("/{studentId}/report-cards", d.Result.GetStudentAllReports)
				r.With(authorize(d, "results", "read")).Get("/{studentId}/report-card", d.Result.GetStudentReportCard)

				// Scores for a student
				r.With(authorize(d, "results", "read")).Get("/{studentId}/scores", d.Result.GetStudentScores)
			})

			// ── ENROLLMENTS ───────────────────────────────────────────────────
			r.Route("/enrollments", func(r chi.Router) {
				r.With(authorize(d, "enrollments", "read")).Get("/", d.Student.ListEnrollments)
				r.With(authorize(d, "enrollments", "create")).Post("/", d.Student.Enroll)
				r.With(authorize(d, "enrollments", "update")).Put("/{enrollmentId}", d.Student.UpdateEnrollment)
			})

			// ── RESULTS ───────────────────────────────────────────────────────
			r.Route("/results", func(r chi.Router) {
				// Score entry
				r.With(authorize(d, "results", "create")).Post("/scores", d.Result.UpsertScore)
				r.With(authorize(d, "results", "create")).Post("/scores/bulk", d.Result.BulkUpsertScores)
				r.With(authorize(d, "results", "update")).Post("/scores/compute-positions", d.Result.ComputePositions)

				// Report cards
				r.With(authorize(d, "results", "read")).Get("/report-cards", d.Result.ListReportCards)
				r.With(authorize(d, "results", "create")).Post("/report-cards/generate", d.Result.GenerateReportCards)
				r.With(authorize(d, "results", "read")).Get("/report-cards/{id}", d.Result.GetReportCard)
				r.With(authorize(d, "results", "update")).Put("/report-cards/{id}/remarks", d.Result.UpdateRemarks)
				r.With(authorize(d, "results", "publish")).Post("/report-cards/{id}/publish", d.Result.Publish)

				// Score & grade configuration
				r.With(authorize(d, "results", "update")).Post("/score-config", d.Result.UpsertScoreConfig)
				r.With(authorize(d, "results", "update")).Post("/grade-config", d.Result.UpsertGradeConfig)
				r.With(authorize(d, "results", "read")).Get("/grade-config", d.Result.ListGradeConfigs)
			})

			// ── AVATARS ───────────────────────────────────────────────────────
			r.Route("/avatar", func(r chi.Router) {
				r.With(authorize(d, "avatar", "update")).Put("/personnel/{id}", d.Avatar.UploadPersonnelAvatar)
				r.With(authorize(d, "avatar", "update")).Put("/students/{id}", d.Avatar.UploadStudentAvatar)
			})

			// ── ATTENDANCE ───────────────────────────────────────────────────
			r.Route("/attendance", func(r chi.Router) {
				r.Route("/personnel", func(r chi.Router) {
					r.With(authorize(d, "attendance", "create")).Post("/", d.Attendance.RecordPersonnelAttendance)
					r.With(authorize(d, "attendance", "read")).Get("/{id}", d.Attendance.GetPersonnelAttendance)
					r.With(authorize(d, "attendance", "update")).Put("/{id}", d.Attendance.UpdatePersonnelAttendance)
					r.With(authorize(d, "attendance", "delete")).Delete("/{id}", d.Attendance.DeletePersonnelAttendance)
					r.With(authorize(d, "attendance", "read")).Get("/school", d.Attendance.ListPersonnelAttendanceBySchoolAndDate)
					r.With(authorize(d, "attendance", "read")).Get("/{id}/range", d.Attendance.ListPersonnelAttendanceByPersonnelAndRange)
				})
				r.Route("/students", func(r chi.Router) {
					r.With(authorize(d, "attendance", "create")).Post("/", d.Attendance.RecordStudentAttendance)
					r.With(authorize(d, "attendance", "create")).Post("/bulk", d.Attendance.BulkRecordStudentAttendance)
					r.With(authorize(d, "attendance", "read")).Get("/{id}", d.Attendance.GetStudentAttendance)
					r.With(authorize(d, "attendance", "update")).Put("/{id}", d.Attendance.UpdateStudentAttendance)
					r.With(authorize(d, "attendance", "delete")).Delete("/{id}", d.Attendance.DeleteStudentAttendance)
					r.With(authorize(d, "attendance", "read")).Get("/school", d.Attendance.ListStudentAttendanceBySchoolAndDate)
					r.With(authorize(d, "attendance", "read")).Get("/student/{id}/range", d.Attendance.ListStudentAttendanceByStudentAndRange)
					r.With(authorize(d, "attendance", "read")).Get("/school/range", d.Attendance.ListStudentAttendanceBySchoolAndRange)
				})
			})

			r.Route("/predictions", func(r chi.Router) {
				r.With(authorize(d, "results", "read")).Get("/schools/{schoolId}", d.Prediction.SchoolReport)
				r.With(authorize(d, "results", "read")).Get("/schools/{schoolId}/full", d.Prediction.FullReport)
				r.With(authorize(d, "results", "read")).Get("/schools/{schoolId}/students/{studentId}", d.Prediction.StudentReport)
			})

			// ── RECOMMENDATIONS ───────────────────────────────────────────────
			r.With(authorize(d, "reports", "read")).Get("/recommendations", d.Recommendation.GetRecommendations)

		})
	})

	return r
}

// authorize is a helper that returns a middleware applying resource+action RBAC.
func authorize(d Deps, resource, action string) func(http.Handler) http.Handler {
	return middleware.Authorize(d.RoleChecker, resource, action)
}
