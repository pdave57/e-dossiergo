// Command server is the e-Dossier API entry point.
// It wires all layers together using manual dependency injection (no DI framework).
//
//	@title			e-Dossier API
//	@version		1.0
//	@description	e-Dossier School Management System API
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.email	support@edossier.com
//
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.basic	BasicAuth
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" + space + JWT token.
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
	"github.com/edossier/api/internal/infrastructure/ml"
	"github.com/edossier/api/internal/infrastructure/repository"
	"github.com/edossier/api/internal/infrastructure/storage"
	"github.com/edossier/api/internal/interfaces/http/handler"
	"github.com/edossier/api/internal/interfaces/http/router"
	"github.com/edossier/api/pkg/logger"
	"github.com/edossier/api/pkg/token"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Error("failed to connect to gorm database", "error", err)
		os.Exit(1)
	}

	//initialize redis
	// redisClient, err := infradb.NewRedisClient(cfg.RedisURL, cfg.RedisPassword, 0)
	// if err != nil {
	// 	log.Error("failed to connect to redis", "error", err)
	// 	os.Exit(1)
	// }
	// defer redisClient.Close()
	// log.Info("redis connected")

	//Initialize Cloudinary (Infrastructure)
	cloudinaryClient, err := storage.NewCloudinaryStorage(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		log.Error("failed to connect to cloudinary", "error", err)
		os.Exit(1)
	}
	log.Info("cloudinary connected")

	// ── Token maker ───────────────────────────────────────────────────────────
	tokenMaker := token.New(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	stateRepo := repository.NewStateRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	lgaRepo := repository.NewLGARepository(db)

	schoolRepo := repository.NewSchoolRepository(db)
	facilityRepo := repository.NewSchoolFacilityRepository(db)

	sessionRepo := repository.NewAcademicSessionRepository(db)
	termRepo := repository.NewTermRepository(db)
	levelRepo := repository.NewLevelRepository(db)
	subLevelRepo := repository.NewSubLevelRepository(db)
	schoolLevelRepo := repository.NewSchoolLevelRepository(db)

	subjectRepo := repository.NewSubjectRepository(db)
	schoolSubjectRepo := repository.NewSchoolSubjectRepository(db)

	personnelRepo := repository.NewPersonnelRepository(db)
	transferRepo := repository.NewPersonnelTransferRepository(db)

	studentRepo := repository.NewStudentRepository(db)
	enrollmentRepo := repository.NewEnrollmentRepository(db)
	progressionRepo := repository.NewLevelProgressionRepository(db)

	scoreSheetRepo := repository.NewScoreSheetRepository(db)
	reportCardRepo := repository.NewReportCardRepository(db)
	gradeConfigRepo := repository.NewGradeConfigRepository(db)
	scoreConfigRepo := repository.NewScoreConfigRepository(db)

	reportRepo := repository.NewReportRepository(db)
	zonalReportRepo := repository.NewZonalReportRepository(gormDB)

	personnelAttendanceRepo := repository.NewPersonnelAttendanceRepository(db)
	studentAttendanceRepo := repository.NewStudentAttendanceRepository(db)
	predictionRepo := repository.NewPredictionRepository(db)
	recommendationRepo := repository.NewRecommendationRepository(db)

	// ── Use Cases ─────────────────────────────────────────────────────────────
	authUC := service.NewAuthService(userRepo, userRoleRepo, roleRepo, refreshTokenRepo, tokenMaker)
	userUC := service.NewUserService(userRepo, userRoleRepo, roleRepo)
	roleUC := service.NewRoleService(roleRepo, permRepo)

	stateUC := service.NewStateService(stateRepo)
	zoneUC := service.NewZoneService(zoneRepo)
	lgaUC := service.NewLGAService(lgaRepo)
	schoolUC := service.NewSchoolService(schoolRepo, facilityRepo, cloudinaryClient)
	academicUC := service.NewAcademicService(sessionRepo, termRepo)

	levelUC := service.NewLevelService(levelRepo, subLevelRepo, schoolLevelRepo)
	subjectUC := service.NewSubjectService(subjectRepo, schoolSubjectRepo)

	personnelUC := service.NewPersonnelService(personnelRepo, transferRepo, schoolRepo)

	studentUC := service.NewStudentService(
		studentRepo, enrollmentRepo, subLevelRepo, progressionRepo, levelRepo, schoolRepo, stateRepo, lgaRepo,
	)
	genderUC := service.NewGenderService(studentRepo)

	resultUC := service.NewResultService(
		scoreSheetRepo, reportCardRepo, gradeConfigRepo, scoreConfigRepo,
		enrollmentRepo, subLevelRepo, termRepo,
	)

	reportUC := service.NewReportService(reportRepo)
	zonalSummaryUC := service.NewZonalSummaryService(zonalReportRepo, sessionRepo)
	avatarUC := service.NewAvatarService(personnelRepo, studentRepo, cloudinaryClient)

	attendanceUC := service.NewAttendanceService(
		personnelAttendanceRepo, studentAttendanceRepo,
		personnelRepo, studentRepo, schoolRepo,
	)
	predictionUC := service.NewPredictionService(predictionRepo)
	recommendationUC := service.NewRecommendationService(recommendationRepo, ml.NewRecommenderClient(cfg.MLServiceURL, 30*time.Second))

	studentAuthUC := service.NewStudentAuthService(studentRepo, schoolRepo, refreshTokenRepo, tokenMaker)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC, userRepo)
	roleHandler := handler.NewRoleHandler(roleUC)
	stateHandler := handler.NewStateHandler(stateUC)
	zoneHandler := handler.NewZoneHandler(zoneUC)
	lgaHandler := handler.NewLGAHandler(lgaUC)
	schoolHandler := handler.NewSchoolHandler(schoolUC)
	academicHandler := handler.NewAcademicHandler(academicUC)
	levelHandler := handler.NewLevelHandler(levelUC)
	subjectHandler := handler.NewSubjectHandler(subjectUC)
	personnelHandler := handler.NewPersonnelHandler(personnelUC)
	studentHandler := handler.NewStudentHandler(studentUC)
	resultHandler := handler.NewResultHandler(resultUC)
	reportHandler := handler.NewReportHandler(reportUC)
	zonalSummaryHandler := handler.NewZonalSummaryHandler(zonalSummaryUC)
	genderHandler := handler.NewGenderHandler(genderUC)
	avatarHandler := handler.NewAvatarHandler(avatarUC)
	studentAuthHandler := handler.NewStudentAuthHandler(studentAuthUC)
	attendanceHandler := handler.NewAttendanceHandler(attendanceUC)
	predictionHandler := handler.NewPredictionHandler(predictionUC)
	recommendationHandler := handler.NewRecommendationHandler(recommendationUC)

	// ── Router ────────────────────────────────────────────────────────────────
	httpHandler := router.New(router.Deps{
		Log:            log,
		TokenMaker:     tokenMaker,
		RoleChecker:    userRoleRepo,
		Auth:           authHandler,
		User:           userHandler,
		Role:           roleHandler,
		State:          stateHandler,
		Zone:           zoneHandler,
		LGA:            lgaHandler,
		School:         schoolHandler,
		Academic:       academicHandler,
		Level:          levelHandler,
		Subject:        subjectHandler,
		Personnel:      personnelHandler,
		Student:        studentHandler,
		Result:         resultHandler,
		Report:         reportHandler,
		ZonalSummary:   zonalSummaryHandler,
		Gender:         genderHandler,
		Avatar:         avatarHandler,
		StudentAuth:    studentAuthHandler,
		Attendance:     attendanceHandler,
		Prediction:     predictionHandler,
		Recommendation: recommendationHandler,
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
