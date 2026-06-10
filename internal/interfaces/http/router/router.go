// Package router wires all routes, middleware, and handlers into a single chi.Router.
package router

import (
	"net/http"

	"log/slog"

	chi "github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/edossier/api/internal/interfaces/http/handler"
	"github.com/edossier/api/internal/interfaces/http/middleware"
	"github.com/edossier/api/pkg/token"
)

// Deps bundles all handler and middleware dependencies.
type Deps struct {
	Log         *slog.Logger
	TokenMaker  *token.Maker
	RoleChecker middleware.UserRoleChecker

	Auth      *handler.AuthHandler
	User      *handler.UserHandler
	Role      *handler.RoleHandler
	Zone      *handler.ZoneHandler
	School    *handler.SchoolHandler
	Academic  *handler.AcademicHandler
	Level     *handler.LevelHandler
	Subject   *handler.SubjectHandler
	Personnel *handler.PersonnelHandler
	Student   *handler.StudentHandler
	Result    *handler.ResultHandler
}

// New builds and returns the fully-configured HTTP router.
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

	// ── API v1 ────────────────────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {

		// ── AUTH (public) ─────────────────────────────────────────────────────
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", d.Auth.Register)
			r.Post("/login", d.Auth.Login)
			r.Post("/refresh", d.Auth.Refresh)
		})

		// ── AUTHENTICATED routes ──────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(d.TokenMaker))

			// Current user
			r.Get("/auth/me", d.Auth.Me)
			r.Post("/auth/logout", d.Auth.Logout)
			r.Post("/auth/change-password", d.Auth.ChangePassword)

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

			// ── ZONE: STATES ───────────────────────────────────────────────────
			r.Route("/states", func(r chi.Router) {
				r.Get("/", d.Zone.ListStates) // open within auth - needed by all
				r.With(authorize(d, "schools", "create")).Post("/", d.Zone.CreateState)
				r.Get("/{id}", d.Zone.GetState)
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.Zone.UpdateState)

				// Zones
				r.Get("/{stateId}/zones", d.Zone.ListZones)
				r.With(authorize(d, "schools", "create")).Post("/{stateId}/zones", d.Zone.CreateZone)

				// LGAs
				r.Get("/{stateId}/lgas", d.Zone.ListLGAs)
				r.With(authorize(d, "schools", "create")).Post("/{stateId}/lgas", d.Zone.CreateLGA)
			})

			r.Route("/zones", func(r chi.Router) {
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.Zone.UpdateZone)
				r.With(authorize(d, "schools", "delete")).Delete("/{id}", d.Zone.DeleteZone)
			})

			r.Route("/lgas", func(r chi.Router) {
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.Zone.UpdateLGA)
				r.With(authorize(d, "schools", "delete")).Delete("/{id}", d.Zone.DeleteLGA)
			})

			// ── SCHOOLS ───────────────────────────────────────────────────────
			r.Route("/schools", func(r chi.Router) {
				r.With(authorize(d, "schools", "read")).Get("/", d.School.List)
				r.With(authorize(d, "schools", "create")).Post("/", d.School.Create)
				r.With(authorize(d, "schools", "read")).Get("/{id}", d.School.GetByID)
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.School.Update)
				r.With(authorize(d, "schools", "delete")).Delete("/{id}", d.School.Delete)

				// Facilities
				r.With(authorize(d, "schools", "read")).Get("/{id}/facilities", d.School.ListFacilities)
				r.With(authorize(d, "schools", "update")).Post("/{id}/facilities", d.School.AddFacility)
				r.With(authorize(d, "schools", "update")).Put("/{id}/facilities/{facilityId}", d.School.UpdateFacility)
				r.With(authorize(d, "schools", "update")).Delete("/{id}/facilities/{facilityId}", d.School.DeleteFacility)

				// School-level assignments
				r.Get("/{schoolId}/levels", d.Level.ListSchoolLevels)
				r.With(authorize(d, "schools", "update")).Post("/{schoolId}/levels", d.Level.UpsertSchoolLevel)

				// Sub-levels (class arms)
				r.With(authorize(d, "schools", "read")).Get("/{schoolId}/sub-levels", func(w http.ResponseWriter, r *http.Request) {
					// delegates to level handler with schoolId from path
					d.Level.ListSubLevels(w, r)
				})
				r.With(authorize(d, "schools", "update")).Post("/{schoolId}/sub-levels", d.Level.CreateSubLevel)

				// School subjects
				r.With(authorize(d, "schools", "read")).Get("/{schoolId}/subjects", d.Subject.ListSchoolSubjects)
				r.With(authorize(d, "schools", "update")).Post("/{schoolId}/subjects", d.Subject.AssignToSchool)
			})

			r.Route("/sub-levels", func(r chi.Router) {
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.Level.UpdateSubLevel)
				r.With(authorize(d, "schools", "update")).Delete("/{id}", d.Level.DeleteSubLevel)
			})

			r.Route("/school-subjects", func(r chi.Router) {
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.Subject.UpdateSchoolSubject)
				r.With(authorize(d, "schools", "update")).Delete("/{id}", d.Subject.RemoveSchoolSubject)
			})

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

			// ── LEVELS (state-wide class definitions) ─────────────────────────
			r.Route("/levels", func(r chi.Router) {
				r.Get("/", d.Level.List)
				r.With(authorize(d, "schools", "create")).Post("/", d.Level.Create)
				r.Get("/{id}", d.Level.GetByID)
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.Level.Update)
				r.With(authorize(d, "schools", "delete")).Delete("/{id}", d.Level.Delete)

				// Sub-levels listed by level
				r.Get("/{levelId}/sub-levels", d.Level.ListSubLevels)
			})

			// ── SUBJECTS (state-wide) ─────────────────────────────────────────
			r.Route("/subjects", func(r chi.Router) {
				r.Get("/", d.Subject.List)
				r.With(authorize(d, "schools", "create")).Post("/", d.Subject.Create)
				r.Get("/{id}", d.Subject.GetByID)
				r.With(authorize(d, "schools", "update")).Put("/{id}", d.Subject.Update)
				r.With(authorize(d, "schools", "delete")).Delete("/{id}", d.Subject.Delete)
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
		})
	})

	return r
}

// authorize is a helper that returns a middleware applying resource+action RBAC.
func authorize(d Deps, resource, action string) func(http.Handler) http.Handler {
	return middleware.Authorize(d.RoleChecker, resource, action)
}
