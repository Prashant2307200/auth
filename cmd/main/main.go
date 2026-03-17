package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Prashant2307200/auth-service/internal/config"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/handler"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/seeder"
	"github.com/Prashant2307200/auth-service/internal/service"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/Prashant2307200/auth-service/pkg/rdb"
	"github.com/Prashant2307200/auth-service/pkg/uploader"
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

	cloudinary, err := uploader.Connect(cfg.Cloud.Name, cfg.Cloud.ApiKey, cfg.Cloud.ApiSecret)
	if err != nil {
		log.Fatalf("Failed to initialize the cloudinary: %s", err.Error())
		return
	}
	slog.Info("Connected to the cloudinary.")

	err = seeder.SeedUsers(context.Background(), userRepo)
	if err != nil {
		log.Fatalf("Failed to seed the users: %s", err.Error())
		return
	}
	slog.Info("Seeded the users.")

	cloudService := service.NewCoudinaryUploadService(cloudinary.Cld)
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

	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))
	router.Handle("/users/", http.StripPrefix("/users", userRouter))
	router.Handle("/business/", http.StripPrefix("/business", businessRouter))

	v1 := http.NewServeMux()
	authMiddleware := middleware.Authenticate(tokenService, cfg.Env)
	v1.Handle("/api/v1/", authMiddleware(http.StripPrefix("/api/v1", router)))

	server := &http.Server{
		Addr:    cfg.HttpServer.Addr,
		Handler: v1,
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

	<-done
	slog.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown the server", slog.String("error", err.Error()))
	}
}
