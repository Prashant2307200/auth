package main

import (
	"context"
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
	"github.com/Prashant2307200/auth-service/internal/seeder"
	"github.com/Prashant2307200/auth-service/internal/service"
	grpcserver "github.com/Prashant2307200/auth-service/internal/transport/grpc/server"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/Prashant2307200/auth-service/pkg/ratelimit"
	"github.com/Prashant2307200/auth-service/pkg/rdb"
)

func main() {
	cfg := config.MustLoad()

	database, err := db.Connect(cfg.PostgresUri)
	if err != nil {
		log.Fatalf("Failed to initialize the storage: %s", err.Error())
	}
	if err := db.RunMigrations(database.Db); err != nil {
		log.Fatalf("Failed to run migrations: %s", err.Error())
	}

	rdb, err := rdb.Connect(cfg.Redis.Addr, cfg.Redis.User, cfg.Redis.Pass)
	if err != nil {
		log.Fatalf("Failed to initialize the cache: %s", err.Error())
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
		log.Fatalf("Failed to initialize the user repository: %s", err.Error())
		return
	}

	businessRepo, err := repository.NewBusinessRepo(database.Db)
	if err != nil {
		log.Fatalf("Failed to initialize the business repository: %s", err.Error())
		return
	}

	// Seed only in dev or when explicitly enabled via SEED_ON_STARTUP env var.
	if cfg.Env == "dev" || os.Getenv("SEED_ON_STARTUP") == "true" {
		err = seeder.SeedAll(context.Background(), userRepo, businessRepo)
		if err != nil {
			log.Fatalf("Failed to seed the database: %s", err.Error())
			return
		}
		slog.Info("Seeded the database.")
	} else {
		slog.Info("Skipping DB seeding on startup.")
	}

	cloudService := service.NewCloudinaryUploadService(cfg.Cloud.Name, cfg.Cloud.ApiKey, cfg.Cloud.ApiSecret)
	tokenService, err := service.NewJWTTokenService(rdb.Rdb, "keys/public.pem", "keys/private.pem", cfg.Secrets.RefreshTokenSecret)
	if err != nil {
		log.Fatalf("Failed to initialize token service: %s", err.Error())
	}

	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := handler.NewUserHandler(userUseCase)
	businessUseCase := usecase.NewBusinessUseCase(businessRepo, userRepo)
	businessHandler := handler.NewBusinessHandler(businessUseCase)

	userRouter := http.NewServeMux()
	userHandler.RegisterRoutes(userRouter)

	businessRouter := http.NewServeMux()
	businessHandler.RegisterRoutes(businessRouter)

	authUseCase := usecase.NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
	authHandler := handler.NewAuthHandler(authUseCase, cfg.Env)

	authRouter := http.NewServeMux()
	authHandler.RegisterRoutes(authRouter)
	authRateLimiter := ratelimit.NewRateLimiter(0.083, 1)
	authRouterWithRateLimit := wrapRateLimitedRoutes(authRouter, authRateLimiter, []string{"/register/", "/login/"})

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
	_ = grpcserver.NewTokenService(tokenService, userRepo)
	_ = grpcserver.NewPublicKeyService(businessRepo)

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
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
				break
			}
		}
		handler.ServeHTTP(w, r)
	})
}
