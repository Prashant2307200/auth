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

	db, err := db.Connect(cfg.PostgresUri)
	if err != nil {
		log.Fatalf("Failed to initialize the storage: %s", err.Error())
	}

	rdb, err := rdb.Connect(cfg.Redis.Addr, cfg.Redis.User, cfg.Redis.Pass)
	if err != nil {
		log.Fatalf("Failed to initialize the cache: %s", err.Error())
	}

	defer func() {
		if err := db.Db.Close(); err != nil {
			log.Fatalf("Failed to close the storage: %s", err.Error())
		}
		if err := rdb.Rdb.Close(); err != nil {
			log.Fatalf("Failed to close the cache: %s", err.Error())
		}
		slog.Info("Cleanup successfull.")
	}()

	userRepo, err := repository.NewUserRepo(db.Db)
	if err != nil {
		log.Fatalf("Failed to initialize the user repository: %s", err.Error())
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
	tokenService := service.NewJWTTokenService(rdb.Rdb, "keys/public.pem", "keys/private.pem", cfg.Secrets.RefreshTokenSecret)

	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := handler.NewUserHandler(userUseCase)

	userRouter := http.NewServeMux()
	userHandler.RegisterRoutes(userRouter)

	authUseCase := usecase.NewAuthUseCase(userRepo, tokenService, cloudService)
	authHandler := handler.NewAuthHandler(authUseCase, cfg.Env)

	authRouter := http.NewServeMux()
	authHandler.RegisterRoutes(authRouter)

	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))
	router.Handle("/users/", http.StripPrefix("/users", userRouter))

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
			log.Fatal("Failed to start server.")
		}
		slog.Info("Server stopped in the background.", slog.String("error", err.Error()))
	}()

	<-done
	slog.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown the server", slog.String("error", err.Error()))
	}
}
