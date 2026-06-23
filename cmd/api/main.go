package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/config"
	"evinizinkocu-backend/internal/db"
	"evinizinkocu-backend/internal/infrastructure/fcm"
	"evinizinkocu-backend/internal/infrastructure/mailer"
	"evinizinkocu-backend/internal/infrastructure/repository"
	"evinizinkocu-backend/internal/middleware"
	transport "evinizinkocu-backend/internal/transport/http"
	"evinizinkocu-backend/internal/worker"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	log.Println("Starting Evinizin Koçu Backend...")

	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Connect to database
	database, err := db.Connect(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	// 3. Run database migrations
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// 4. Initialize Repositories
	userRepo := repository.NewPostgresUserRepository(database.Pool)
	coachRepo := repository.NewPostgresCoachRepository(database.Pool)
	studentRepo := repository.NewPostgresStudentRepository(database.Pool)
	catalogRepo := repository.NewPostgresCatalogRepository(database.Pool)
	examRepo := repository.NewPostgresExamRepository(database.Pool)
	trackersRepo := repository.NewPostgresTrackersRepository(database.Pool)
	paymentRepo := repository.NewPostgresPaymentRepository(database.Pool)
	notifRepo := repository.NewPostgresNotificationRepository(database.Pool)
	cmsRepo := repository.NewPostgresCMSRepository(database.Pool)

	// 5. Initialize Infrastructure Services (SMTP & FCM)
	mailerService := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	fcmService := fcm.NewGoogleFCMService(cfg.FirebaseCredentialJSON)

	// 6. Initialize Application Services
	authService := application.NewAuthService(userRepo, mailerService, cfg.JWTSecret, cfg.RefreshJWTSecret)
	coachService := application.NewCoachService(coachRepo, userRepo, mailerService, cfg.MailTo)
	studentService := application.NewStudentService(studentRepo, userRepo, coachRepo, mailerService)
	catalogService := application.NewCatalogService(catalogRepo)
	examService := application.NewExamService(examRepo, catalogRepo, studentRepo)
	trackersService := application.NewTrackersService(trackersRepo, studentRepo, examRepo, catalogRepo)
	paymentService := application.NewPaymentService(paymentRepo, coachRepo, studentRepo)
	notifService := application.NewNotificationService(notifRepo, studentRepo, userRepo, coachRepo, fcmService)
	cmsService := application.NewCMSService(cmsRepo)
	dashboardService := application.NewDashboardService(database.Pool)

	// 7. Start Background Workers
	notifWorker := worker.NewNotificationWorker(notifRepo, studentRepo, userRepo, coachRepo, fcmService, notifService)
	notifWorker.Start(15 * time.Second)
	defer notifWorker.Stop()

	// Daily coach deactivation expiry worker
	expiryTicker := time.NewTicker(6 * time.Hour)
	go func() {
		for range expiryTicker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_ = coachService.CheckAndProcessExpiredCoaches(ctx)
			cancel()
		}
	}()
	defer expiryTicker.Stop()

	// 8. Setup routing & HTTP Middlewares
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// CORS Config
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Replace in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health & Readiness checks
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "UP"})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := database.Pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "DOWN", "error": "Database ping failed"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "READY"})
	})

	// Auth Middleware helper
	authMiddleware := middleware.Authenticate(cfg.JWTSecret, userRepo)

	// 9. Register Route Handlers
	authHandler := transport.NewAuthHandler(authService)
	authHandler.RegisterRoutes(r, authMiddleware)

	coachHandler := transport.NewCoachHandler(coachService)
	coachHandler.RegisterRoutes(r, authMiddleware)

	studentHandler := transport.NewStudentHandler(studentService)
	studentHandler.RegisterRoutes(r, authMiddleware)

	catalogHandler := transport.NewCatalogHandler(catalogService)
	catalogHandler.RegisterRoutes(r, authMiddleware)

	examHandler := transport.NewExamHandler(examService, studentService)
	examHandler.RegisterRoutes(r, authMiddleware)

	trackersHandler := transport.NewTrackersHandler(trackersService, studentService)
	trackersHandler.RegisterRoutes(r, authMiddleware)

	paymentHandler := transport.NewPaymentHandler(paymentService, studentService)
	paymentHandler.RegisterRoutes(r, authMiddleware)

	notifHandler := transport.NewNotificationHandler(notifService)
	notifHandler.RegisterRoutes(r, authMiddleware)

	cmsHandler := transport.NewCMSHandler(cmsService)
	cmsHandler.RegisterRoutes(r, authMiddleware)

	dashboardHandler := transport.NewDashboardHandler(dashboardService)
	dashboardHandler.RegisterRoutes(r, authMiddleware)

	// 10. Start Server
	serverAddr := ":" + cfg.Port
	log.Printf("Server listening on %s\n", serverAddr)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
