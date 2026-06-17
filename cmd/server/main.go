// Command server is the e-Dossier API entry point.
// It wires all layers together using manual dependency injection (no DI framework).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edossier/api/config"
	"github.com/edossier/api/internal/application/service"
	infradb "github.com/edossier/api/internal/infrastructure/db"
	"github.com/edossier/api/internal/infrastructure/repository"
	"github.com/edossier/api/internal/interfaces/http/handler"
	"github.com/edossier/api/internal/interfaces/http/router"
	"github.com/edossier/api/pkg/logger"
	"github.com/edossier/api/pkg/token"
)

func main() {
	// ── Configuration ─────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := infradb.Open(cfg.DatabaseURL, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns, cfg.DBConnMaxLife)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("database connected")

	if err := infradb.Migrate(db); err != nil {
		log.Error("schema migration failed", "error", err)
		os.Exit(1)
	}
	log.Info("schema migration complete")

	// ── Token maker ───────────────────────────────────────────────────────────
	tokenMaker := token.New(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo          := repository.NewUserRepository(db)
	roleRepo          := repository.NewRoleRepository(db)
	permRepo          := repository.NewPermissionRepository(db)
	userRoleRepo      := repository.NewUserRoleRepository(db)
	refreshTokenRepo  := repository.NewRefreshTokenRepository(db)

	stateRepo         := repository.NewStateRepository(db)
	zoneRepo          := repository.NewZoneRepository(db)
	lgaRepo           := repository.NewLGARepository(db)

	schoolRepo        := repository.NewSchoolRepository(db)
	facilityRepo      := repository.NewSchoolFacilityRepository(db)

	sessionRepo       := repository.NewAcademicSessionRepository(db)
	termRepo          := repository.NewTermRepository(db)

	levelRepo         := repository.NewLevelRepository(db)
	subLevelRepo      := repository.NewSubLevelRepository(db)
	schoolLevelRepo   := repository.NewSchoolLevelRepository(db)

	subjectRepo       := repository.NewSubjectRepository(db)
	schoolSubjectRepo := repository.NewSchoolSubjectRepository(db)

	personnelRepo     := repository.NewPersonnelRepository(db)
	transferRepo      := repository.NewPersonnelTransferRepository(db)

	studentRepo       := repository.NewStudentRepository(db)
	enrollmentRepo    := repository.NewEnrollmentRepository(db)
	progressionRepo   := repository.NewLevelProgressionRepository(db)

	scoreSheetRepo    := repository.NewScoreSheetRepository(db)
	reportCardRepo    := repository.NewReportCardRepository(db)
	gradeConfigRepo   := repository.NewGradeConfigRepository(db)
	scoreConfigRepo   := repository.NewScoreConfigRepository(db)

	reportRepo        := repository.NewReportRepository(db)

	// ── Use Cases ─────────────────────────────────────────────────────────────
	authUC := service.NewAuthService(userRepo, userRoleRepo, roleRepo, refreshTokenRepo, tokenMaker)
	userUC := service.NewUserService(userRepo, userRoleRepo, roleRepo)
	roleUC := service.NewRoleService(roleRepo, permRepo)

	zoneUC      := service.NewZoneService(stateRepo, zoneRepo, lgaRepo)
	schoolUC   := service.NewSchoolService(schoolRepo, facilityRepo)
	academicUC := service.NewAcademicService(sessionRepo, termRepo)

	levelUC := service.NewLevelService(levelRepo, subLevelRepo, schoolLevelRepo)
	subjectUC := service.NewSubjectService(subjectRepo, schoolSubjectRepo)

	personnelUC := service.NewPersonnelService(personnelRepo, transferRepo, schoolRepo)

	studentUC := service.NewStudentService(
		studentRepo, enrollmentRepo, subLevelRepo, progressionRepo, levelRepo, schoolRepo,
	)

	resultUC := service.NewResultService(
		scoreSheetRepo, reportCardRepo, gradeConfigRepo, scoreConfigRepo,
		enrollmentRepo, subLevelRepo, termRepo,
	)

	reportUC := service.NewReportService(reportRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler      := handler.NewAuthHandler(authUC)
	userHandler      := handler.NewUserHandler(userUC, userRepo)
	roleHandler      := handler.NewRoleHandler(roleUC)
	zoneHandler      := handler.NewZoneHandler(zoneUC)
	schoolHandler    := handler.NewSchoolHandler(schoolUC)
	academicHandler  := handler.NewAcademicHandler(academicUC)
	levelHandler     := handler.NewLevelHandler(levelUC)
	subjectHandler   := handler.NewSubjectHandler(subjectUC)
	personnelHandler := handler.NewPersonnelHandler(personnelUC)
	studentHandler   := handler.NewStudentHandler(studentUC)
	resultHandler    := handler.NewResultHandler(resultUC)
	reportHandler    := handler.NewReportHandler(reportUC)

	// ── Router ────────────────────────────────────────────────────────────────
	httpHandler := router.New(router.Deps{
		Log:         log,
		TokenMaker:  tokenMaker,
		RoleChecker: userRoleRepo,
		Auth:        authHandler,
		User:        userHandler,
		Role:        roleHandler,
		Zone:        zoneHandler,
		School:      schoolHandler,
		Academic:    academicHandler,
		Level:       levelHandler,
		Subject:     subjectHandler,
		Personnel:   personnelHandler,
		Student:     studentHandler,
		Result:      resultHandler,
		Report:      reportHandler,
	})

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Info("e-Dossier API starting", "port", cfg.Port, "env", cfg.Env)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdown:
		log.Info("shutdown signal received", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		log.Info("server stopped gracefully")
	}
}
