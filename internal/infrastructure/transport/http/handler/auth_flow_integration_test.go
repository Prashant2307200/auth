package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/service"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/Prashant2307200/auth-service/pkg/rdb"
	"github.com/stretchr/testify/require"
)

type authFlowEnv struct {
	PostgresURI string
	RedisAddr   string
	RedisUser   string
	RedisPass   string
}

func TestAuthFlow_Integration(t *testing.T) {
	env := loadIntegrationEnv()
	if env.PostgresURI == "" || env.RedisAddr == "" {
		t.Skip("integration env not set")
	}

	postgres, err := db.Connect(env.PostgresURI)
	require.NoError(t, err)
	defer postgres.Db.Close()
	require.NoError(t, db.RunMigrations(postgres.Db))

	_, err = postgres.Db.Exec(`TRUNCATE TABLE business_domains, business_invites, business_users, businesses, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	redisConn, err := rdb.Connect(env.RedisAddr, env.RedisUser, env.RedisPass)
	require.NoError(t, err)
	defer redisConn.Rdb.Close()
	require.NoError(t, redisConn.Rdb.FlushDB(context.Background()).Err())

	pubPath, privPath := writeTempRSAKeyPair(t)
	tokenService, err := service.NewJWTTokenService(redisConn.Rdb, pubPath, privPath, "integration-refresh-secret")
	require.NoError(t, err)

	userRepo, err := repository.NewUserRepo(postgres.Db)
	require.NoError(t, err)
	businessRepo, err := repository.NewBusinessRepo(postgres.Db)
	require.NoError(t, err)
	cloud := &testutil.MockCloudService{}

	authUC := usecase.NewAuthUseCase(userRepo, businessRepo, tokenService, cloud)
	authHandler := NewAuthHandler(authUC, "dev")
	authRouter := http.NewServeMux()
	authHandler.RegisterRoutes(authRouter)

	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))
	api := middleware.Authenticate(tokenService, "dev")(http.StripPrefix("/api/v1", router))

	unique := strconv.FormatInt(time.Now().UnixNano(), 10)
	registerBody := map[string]any{
		"username":    "user_" + unique,
		"email":       "user_" + unique + "@example.com",
		"password":    "Password1!",
		"profile_pic": "https://res.cloudinary.com/demo/image/upload/v1/pic.jpg",
	}
	b, _ := json.Marshal(registerBody)

	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/", bytes.NewReader(b))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	api.ServeHTTP(registerResp, registerReq)
	require.Equal(t, http.StatusOK, registerResp.Code)

	var accessCookie *http.Cookie
	for _, c := range registerResp.Result().Cookies() {
		if c.Name == "access_token" {
			accessCookie = c
			break
		}
	}
	require.NotNil(t, accessCookie)

	profileReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/profile/", nil)
	profileReq.AddCookie(accessCookie)
	profileResp := httptest.NewRecorder()
	api.ServeHTTP(profileResp, profileReq)
	require.Equal(t, http.StatusOK, profileResp.Code)

	loginBody := map[string]any{
		"email":    registerBody["email"],
		"password": "Password1!",
	}
	lb, _ := json.Marshal(loginBody)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/", bytes.NewReader(lb))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	api.ServeHTTP(loginResp, loginReq)
	require.Equal(t, http.StatusOK, loginResp.Code)
	require.NotEmpty(t, loginResp.Result().Cookies())
}

func TestAuthFlow_Integration_RefreshInvalidToken(t *testing.T) {
	env := loadIntegrationEnv()
	if env.PostgresURI == "" || env.RedisAddr == "" {
		t.Skip("integration env not set")
	}

	postgres, err := db.Connect(env.PostgresURI)
	require.NoError(t, err)
	defer postgres.Db.Close()
	require.NoError(t, db.RunMigrations(postgres.Db))

	_, err = postgres.Db.Exec(`TRUNCATE TABLE business_domains, business_invites, business_users, businesses, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	redisConn, err := rdb.Connect(env.RedisAddr, env.RedisUser, env.RedisPass)
	require.NoError(t, err)
	defer redisConn.Rdb.Close()
	require.NoError(t, redisConn.Rdb.FlushDB(context.Background()).Err())

	pubPath, privPath := writeTempRSAKeyPair(t)
	tokenService, err := service.NewJWTTokenService(redisConn.Rdb, pubPath, privPath, "integration-refresh-secret")
	require.NoError(t, err)

	userRepo, err := repository.NewUserRepo(postgres.Db)
	require.NoError(t, err)
	businessRepo, err := repository.NewBusinessRepo(postgres.Db)
	require.NoError(t, err)
	cloud := &testutil.MockCloudService{}

	authUC := usecase.NewAuthUseCase(userRepo, businessRepo, tokenService, cloud)
	authHandler := NewAuthHandler(authUC, "dev")
	authRouter := http.NewServeMux()
	authHandler.RegisterRoutes(authRouter)

	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))
	api := middleware.Authenticate(tokenService, "dev")(http.StripPrefix("/api/v1", router))

	refreshReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/refresh/", nil)
	refreshReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-refresh-token"})
	refreshResp := httptest.NewRecorder()
	api.ServeHTTP(refreshResp, refreshReq)

	require.Equal(t, http.StatusUnauthorized, refreshResp.Code)
	var er map[string]any
	require.NoError(t, json.NewDecoder(refreshResp.Body).Decode(&er))
	require.Equal(t, "UNAUTHORIZED", er["code"])
}

func loadIntegrationEnv() authFlowEnv {
	return authFlowEnv{
		PostgresURI: os.Getenv("INTEGRATION_POSTGRES_URI"),
		RedisAddr:   os.Getenv("INTEGRATION_REDIS_ADDRESS"),
		RedisUser:   os.Getenv("INTEGRATION_REDIS_USERNAME"),
		RedisPass:   os.Getenv("INTEGRATION_REDIS_PASSWORD"),
	}
}

func writeTempRSAKeyPair(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privDER := x509.MarshalPKCS1PrivateKey(key)
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	dir := t.TempDir()
	privPath := filepath.Join(dir, "private.pem")
	pubPath := filepath.Join(dir, "public.pem")

	privFile, err := os.Create(privPath)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(privFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER}))
	require.NoError(t, privFile.Close())

	pubFile, err := os.Create(pubPath)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
	require.NoError(t, pubFile.Close())

	return pubPath, privPath
}
