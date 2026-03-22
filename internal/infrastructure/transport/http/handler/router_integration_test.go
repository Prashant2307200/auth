package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/logging"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildTestMux(t *testing.T) http.Handler {
	t.Helper()
	userRepo := &testutil.MockUserRepo{}
	userRepo.On("List", mock.Anything).Return([]*entity.User{}, nil)
	userRepo.On("GetById", mock.Anything, mock.Anything).Return((*entity.User)(nil), nil)

	businessRepo := &testutil.MockBusinessRepo{}
	businessRepo.On("GetUserBusinesses", mock.Anything, mock.Anything).Return([]*entity.Business{}, nil)

	tokenService := testutil.NewTestTokenService(t)

	userUC := usecase.NewUserUseCase(userRepo)
	businessUC := usecase.NewBusinessUseCase(businessRepo, userRepo)
	userHandler := NewUserHandler(userUC)
	businessHandler := NewBusinessHandler(businessUC)

	userRouter := http.NewServeMux()
	userHandler.RegisterRoutes(userRouter)
	businessRouter := http.NewServeMux()
	businessHandler.RegisterRoutes(businessRouter)

	router := http.NewServeMux()
	router.Handle("/users/", http.StripPrefix("/users", userRouter))
	router.Handle("/business/", http.StripPrefix("/business", businessRouter))

	mockCloud := &testutil.MockCloudService{}
	authUC := usecase.NewAuthUseCase(userRepo, businessRepo, tokenService, mockCloud)
	authHandler := NewAuthHandler(authUC, "test")
	authRouter := http.NewServeMux()
	authHandler.RegisterRoutes(authRouter)
	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))

	authMiddleware := middleware.Authenticate(tokenService, "test")
	apiHandler := authMiddleware(http.StripPrefix("/api/v1", router))

	healthUC := usecase.NewHealthUseCase(userRepo, nil)
	healthHandler := NewHealthHandler(healthUC)

	v1 := http.NewServeMux()
	v1.Handle("/health", healthHandler)
	v1.Handle("/health/", healthHandler)
	v1.Handle("/api/v1/", apiHandler)

	return middleware.SecurityHeaders(logging.RequestIDMiddleware(v1))
}

func TestRouterIntegration_HealthReturns200WithSecurityHeaders(t *testing.T) {
	mux := buildTestMux(t)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
}

func TestRouterIntegration_ProtectedRouteReturns401WithoutAuth(t *testing.T) {
	mux := buildTestMux(t)

	req := httptest.NewRequest("GET", "/api/v1/users/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRouterIntegration_ProtectedRouteSucceedsWithValidToken(t *testing.T) {
	tokenService := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}
	userRepo.On("List", mock.Anything).Return([]*entity.User{testutil.CreateTestUserWithID(1)}, nil)
	userRepo.On("GetById", mock.Anything, int64(42)).Return(testutil.CreateTestAdminWithID(42), nil)
	businessRepo := &testutil.MockBusinessRepo{}
	businessRepo.On("GetUserBusinesses", mock.Anything, mock.Anything).Return([]*entity.Business{}, nil)

	claims := jwt.MapClaims{
		"userId": float64(42),
		"exp":    time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenStr, err := token.SignedString(tokenService.AccessSecret)
	require.NoError(t, err)

	userUC := usecase.NewUserUseCase(userRepo)
	businessUC := usecase.NewBusinessUseCase(businessRepo, userRepo)
	userHandler := NewUserHandler(userUC)
	businessHandler := NewBusinessHandler(businessUC)
	userRouter := http.NewServeMux()
	userHandler.RegisterRoutes(userRouter)
	businessRouter := http.NewServeMux()
	businessHandler.RegisterRoutes(businessRouter)
	router := http.NewServeMux()
	router.Handle("/users/", http.StripPrefix("/users", userRouter))
	router.Handle("/business/", http.StripPrefix("/business", businessRouter))
	mockCloud := &testutil.MockCloudService{}
	authUC := usecase.NewAuthUseCase(userRepo, businessRepo, tokenService, mockCloud)
	authHandler := NewAuthHandler(authUC, "test")
	authRouter := http.NewServeMux()
	authHandler.RegisterRoutes(authRouter)
	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))

	authMiddleware := middleware.Authenticate(tokenService, "test")
	apiHandler := authMiddleware(http.StripPrefix("/api/v1", router))
	healthUC := usecase.NewHealthUseCase(userRepo, nil)
	healthHandler := NewHealthHandler(healthUC)
	v1 := http.NewServeMux()
	v1.Handle("/health", healthHandler)
	v1.Handle("/health/", healthHandler)
	v1.Handle("/api/v1/", apiHandler)
	mux := middleware.SecurityHeaders(logging.RequestIDMiddleware(v1))

	req := httptest.NewRequest("GET", "/api/v1/users/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var users []map[string]any
	err = json.NewDecoder(rr.Body).Decode(&users)
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, "testuser", users[0]["username"])
}
