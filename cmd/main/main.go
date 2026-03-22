package main

import (
	"context"
	"encoding/hex"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/Prashant2307200/auth-service/internal/config"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/handler"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/logging"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
	"github.com/Prashant2307200/auth-service/internal/seeder"
	"github.com/Prashant2307200/auth-service/internal/service"
	authgrpcproto "github.com/Prashant2307200/auth-service/internal/transport/grpc/proto"
	grpcserver "github.com/Prashant2307200/auth-service/internal/transport/grpc/server"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/Prashant2307200/auth-service/pkg/invitetoken"
	"github.com/Prashant2307200/auth-service/pkg/ratelimit"
	"github.com/Prashant2307200/auth-service/pkg/rdb"
)

func main() {
	cfg := config.MustLoad()

	database, err := db.Connect(cfg.PostgresUri)
	if err != nil {
		slog.Error("Failed to initialize the storage", slog.Any("error", err))
		os.Exit(1)
	}
	if err := db.RunMigrations(database.Db); err != nil {
		slog.Error("Failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	rdb, err := rdb.Connect(cfg.Redis.Addr, cfg.Redis.User, cfg.Redis.Pass)
	if err != nil {
		slog.Error("Failed to initialize the cache", slog.Any("error", err))
		os.Exit(1)
	}

	defer func() {
		if err := database.Db.Close(); err != nil {
			slog.Error("Failed to close database connection", slog.Any("error", err))
		}
		if err := rdb.Rdb.Close(); err != nil {
			slog.Error("Failed to close redis connection", slog.Any("error", err))
		}
		slog.Info("Cleanup successful.")
	}()

	userRepo, err := repository.NewUserRepo(database.Db)
	if err != nil {
		slog.Error("Failed to initialize the user repository", slog.Any("error", err))
		os.Exit(1)
		return
	}

	businessRepo, err := repository.NewBusinessRepo(database.Db)
	if err != nil {
		slog.Error("Failed to initialize the business repository", slog.Any("error", err))
		os.Exit(1)
		return
	}

	memberRepo, err := repository.NewMemberRepo(database.Db)
	if err != nil {
		slog.Error("Failed to initialize the member repository", slog.Any("error", err))
		os.Exit(1)
	}
	auditRepo, err := repository.NewAuditRepo(database.Db)
	if err != nil {
		slog.Error("Failed to initialize the audit repository", slog.Any("error", err))
		os.Exit(1)
	}

	// Seed only in dev, or when explicitly enabled and not in production.
	// This prevents accidental seeding in production even if the env var is set.
	if cfg.Env == "dev" || (os.Getenv("SEED_ON_STARTUP") == "true" && cfg.Env != "prod") {
		err = seeder.SeedAll(context.Background(), userRepo, businessRepo)
		if err != nil {
			slog.Error("Failed to seed the database", slog.Any("error", err))
			os.Exit(1)
			return
		}
		slog.Info("Seeded the database.")
	} else {
		slog.Info("Skipping DB seeding on startup.")
	}

	cloudService := service.NewCloudinaryUploadService(cfg.Cloud.Name, cfg.Cloud.ApiKey, cfg.Cloud.ApiSecret)
	// Allow configurable JWT key locations via config, with fallbacks to legacy paths for compatibility.
	publicKeyPath := cfg.JWT.PublicKeyPath
	privateKeyPath := cfg.JWT.PrivateKeyPath
	if publicKeyPath == "" {
		publicKeyPath = "keys/public.pem"
	}
	if privateKeyPath == "" {
		privateKeyPath = "keys/private.pem"
	}
	tokenService, err := service.NewJWTTokenService(rdb.Rdb, publicKeyPath, privateKeyPath, cfg.Secrets.RefreshTokenSecret)
	if err != nil {
		slog.Error("Failed to initialize token service", slog.Any("error", err))
		os.Exit(1)
	}

	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := handler.NewUserHandler(userUseCase)
	businessUseCase := usecase.NewBusinessUseCase(businessRepo, userRepo)
	businessHandler := handler.NewBusinessHandler(businessUseCase)

	userRouter := http.NewServeMux()
	userHandler.RegisterRoutes(userRouter)

	businessRouter := http.NewServeMux()
	businessHandler.RegisterRoutes(businessRouter)

	authUseCase := usecase.NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService).WithAudit(auditRepo)
	authHandler := handler.NewAuthHandler(authUseCase, cfg.Env)

	var emailService usecase.EmailService = service.NoopEmailService{}
	if cfg.Email.APIKey != "" {
		emailService = service.NewMailerooService(service.MailerooConfig{
			APIKey:    cfg.Email.APIKey,
			FromEmail: cfg.Email.FromEmail,
			FromName:  cfg.Email.FromName,
			BaseURL:   cfg.Email.BaseURL,
		})
	}

	passwordResetRepo := repository.NewPasswordResetRepo(database.Db)
	passwordResetUC := usecase.NewPasswordResetUsecase(userRepo, passwordResetRepo, emailService, tokenService, auditRepo)
	passwordResetHandler := handler.NewPasswordResetHandler(passwordResetUC)

	emailVerificationRepo := repository.NewEmailVerificationRepo(database.Db)
	emailVerificationUC := usecase.NewEmailVerificationUsecase(userRepo, emailVerificationRepo, emailService, auditRepo)
	emailVerificationHandler := handler.NewEmailVerificationHandler(emailVerificationUC)

	mfaRepo := repository.NewMFARepo(database.Db)
	var mfaEncryptionKey []byte
	if cfg.MFA.EncryptionKey != "" {
		var decErr error
		mfaEncryptionKey, decErr = hex.DecodeString(cfg.MFA.EncryptionKey)
		if decErr != nil || len(mfaEncryptionKey) != 32 {
			slog.Error("MFA_ENCRYPTION_KEY must be 64 hex characters (32 bytes). Generate with: openssl rand -hex 32")
			os.Exit(1)
		}
	} else {
		slog.Warn("MFA_ENCRYPTION_KEY not set — TOTP secrets will be stored unencrypted. Set this in production!")
	}
	mfaUC := usecase.NewMFAUsecase(userRepo, mfaRepo, auditRepo, mfaEncryptionKey)
	authUseCase.WithMFA(mfaUC)
	mfaHandler := handler.NewMFAHandler(mfaUC, userRepo)

	authRouter := http.NewServeMux()
	authHandler.RegisterRoutes(authRouter)
	passwordResetHandler.RegisterRoutes(authRouter)
	emailVerificationHandler.RegisterRoutes(authRouter)
	mfaHandler.RegisterRoutes(authRouter)

	if cfg.OAuth.GoogleClientID != "" && cfg.OAuth.GoogleClientSecret != "" {
		ssoUC := usecase.NewSSOUsecase(userRepo, tokenService, usecase.SSOConfig{
			GoogleClientID:     cfg.OAuth.GoogleClientID,
			GoogleClientSecret: cfg.OAuth.GoogleClientSecret,
			GoogleRedirectURL:  cfg.OAuth.GoogleRedirectURL,
		})
		ssoHandler := handler.NewSSOHandler(ssoUC, cfg.Env, cfg.Email.BaseURL)
		ssoHandler.RegisterRoutes(authRouter)
		slog.Info("Google SSO enabled")
	}

	sessionService := service.NewSessionService(rdb.Rdb)
	sessionHandler := handler.NewSessionHandler(sessionService, cfg.Env)
	sessionHandler.RegisterRoutes(authRouter)

	auditHandler := handler.NewAuditHandler(auditRepo)
	auditHandler.RegisterRoutes(authRouter)
	authRateLimiter := ratelimit.NewRateLimiter(0.083, 1)
	authRouterWithRateLimit := wrapRateLimitedRoutes(authRouter, authRateLimiter, []string{"/register/", "/login/", "/forgot-password", "/reset-password"})

	teamUC := usecase.NewTeamUsecase(memberRepo, auditRepo, service.NoopEmailService{}, invitetoken.NewGenerator(cfg.Secrets.RefreshTokenSecret, 24))
	teamHandler := handler.NewTeamHandler(teamUC, func(next http.Handler) http.Handler { return next })
	teamRouter := http.NewServeMux()
	teamHandler.RegisterRoutes(teamRouter)
	teamHTTP := middleware.TenantFromHeader(http.StripPrefix("/team", teamRouter))

	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()
	go func() {
		for range cleanupTicker.C {
			authRateLimiter.Cleanup(1 * time.Hour)
		}
	}()

	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix("/auth", authRouterWithRateLimit))
	router.Handle("/users/", http.StripPrefix("/users", userRouter))
	router.Handle("/business/", http.StripPrefix("/business", businessRouter))
	router.Handle("/team/", teamHTTP)

	v1 := http.NewServeMux()

	// Health endpoints (public)
	// keep existing health usecase handler for backward compatibility
	healthUseCase := usecase.NewHealthUseCase(userRepo, rdb.Rdb)
	healthHandler := handler.NewHealthHandler(healthUseCase)
	v1.Handle("/health", healthHandler)
	v1.Handle("/health/", healthHandler)

	// System-level live/ready endpoints that perform DB + Redis checks
	sysHealth := handler.NewSystemHealthHandler(database.Db, rdb.Rdb)
	v1.HandleFunc("/health/live", sysHealth.Live)
	v1.HandleFunc("/health/ready", sysHealth.Ready)

	if cfg.Env == "dev" {
		devHandler := handler.NewDevHandler(userRepo, businessRepo)
		v1.HandleFunc("POST /seed-db", devHandler.SeedDB)
	}

	authMiddleware := middleware.Authenticate(tokenService, cfg.Env)
	v1.Handle("/api/v1/", authMiddleware(http.StripPrefix("/api/v1", router)))

	// Register Prometheus metrics endpoint after other v1 routes are configured.
	handler.RegisterMetricsHandler(v1)

	handler := middleware.SecurityHeaders(logging.RequestIDMiddleware(v1))
	server := &http.Server{
		Addr:              cfg.HttpServer.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("Server starting...", slog.String("address", cfg.HttpServer.Addr))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", slog.Any("error", err))
			log.Fatal("Failed to start server.")
		}
		if err == http.ErrServerClosed {
			slog.Info("Server stopped gracefully.")
		}
	}()

	grpcListener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %s", err.Error())
	}

	grpcServer := grpc.NewServer()
	tokenGRPC := grpcserver.NewTokenService(tokenService, userRepo)
	publicKeyGRPC := grpcserver.NewPublicKeyService(businessRepo)
	authgrpcproto.RegisterTokenServiceServer(grpcServer, tokenGRPC)
	authgrpcproto.RegisterPublicKeyServiceServer(grpcServer, publicKeyGRPC)

	go func() {
		slog.Info("gRPC server starting...", slog.String("address", ":9090"))
		if err := grpcServer.Serve(grpcListener); err != nil {
			slog.Error("gRPC server failed", slog.Any("error", err))
		}
	}()

	<-done
	slog.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown the server", slog.String("error", err.Error()))
	}

	grpcServer.GracefulStop()
	if err := grpcListener.Close(); err != nil {
		slog.Error("Failed to close gRPC listener", slog.Any("error", err))
	}
}

func wrapRateLimitedRoutes(handler http.Handler, limiter *ratelimit.RateLimiter, routes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range routes {
			// Use prefix matching so mounted paths (with prefixes) are correctly matched.
			if strings.HasPrefix(r.URL.Path, route) {
				if !limiter.Allow(r) {
					w.Header().Set("Retry-After", "60")
					utils.SendErrorResponse(w, http.StatusTooManyRequests, utils.RATE_LIMITED, "Rate limit exceeded")
					return
				}
				break
			}
		}
		handler.ServeHTTP(w, r)
	})
}
