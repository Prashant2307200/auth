package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

var (
	ErrGoogleAuthFailed   = errors.New("google authentication failed")
	ErrGoogleEmailMissing = errors.New("google email not available")
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type SSOConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

type SSOUsecase interface {
	GetGoogleAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (accessToken, refreshToken string, user *entity.User, isNewUser bool, err error)
}

type ssoUsecase struct {
	userRepo     interfaces.UserRepo
	tokenService interfaces.TokenService
	oauthConfig  *oauth2.Config
}

func NewSSOUsecase(userRepo interfaces.UserRepo, tokenService interfaces.TokenService, cfg SSOConfig) SSOUsecase {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &ssoUsecase{
		userRepo:     userRepo,
		tokenService: tokenService,
		oauthConfig:  oauthConfig,
	}
}

func (u *ssoUsecase) GetGoogleAuthURL(state string) string {
	return u.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (u *ssoUsecase) HandleGoogleCallback(ctx context.Context, code string) (string, string, *entity.User, bool, error) {
	token, err := u.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", "", nil, false, fmt.Errorf("%w: %v", ErrGoogleAuthFailed, err)
	}

	client := u.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", nil, false, fmt.Errorf("%w: %v", ErrGoogleAuthFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", nil, false, ErrGoogleAuthFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, false, fmt.Errorf("%w: %v", ErrGoogleAuthFailed, err)
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return "", "", nil, false, fmt.Errorf("%w: %v", ErrGoogleAuthFailed, err)
	}

	if googleUser.Email == "" {
		return "", "", nil, false, ErrGoogleEmailMissing
	}

	var user *entity.User

	user, err = u.userRepo.GetByGoogleID(ctx, googleUser.ID)
	if err == nil {
		accessToken, refreshToken, err := u.generateTokens(ctx, user.ID)
		return accessToken, refreshToken, user, false, err
	}

	user, err = u.userRepo.GetByEmail(ctx, googleUser.Email)
	if err == nil {
		if err := u.userRepo.LinkGoogleID(ctx, user.ID, googleUser.ID); err != nil {
			return "", "", nil, false, fmt.Errorf("failed to link google account: %w", err)
		}
		if err := u.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
			// Non-fatal, just log
		}
		accessToken, refreshToken, err := u.generateTokens(ctx, user.ID)
		return accessToken, refreshToken, user, false, err
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", nil, false, fmt.Errorf("failed to check existing user: %w", err)
	}

	newUser := &entity.User{
		Username:        generateUsernameFromEmail(googleUser.Email),
		Email:           googleUser.Email,
		Password:        "",
		ProfilePic:      googleUser.Picture,
		Role:            entity.RoleUser,
		EmailVerified:   true,
		EmailVerifiedAt: func() *time.Time { t := time.Now(); return &t }(),
		GoogleID:        &googleUser.ID,
	}

	userID, err := u.userRepo.Create(ctx, newUser)
	if err != nil {
		return "", "", nil, false, fmt.Errorf("failed to create user: %w", err)
	}
	newUser.ID = userID

	if err := u.userRepo.LinkGoogleID(ctx, userID, googleUser.ID); err != nil {
		// Non-fatal, user was created
	}
	if err := u.userRepo.MarkEmailVerified(ctx, userID); err != nil {
		// Non-fatal
	}

	accessToken, refreshToken, err := u.generateTokens(ctx, userID)
	return accessToken, refreshToken, newUser, true, err
}

func (u *ssoUsecase) generateTokens(ctx context.Context, userID int64) (string, string, error) {
	accessToken, err := u.tokenService.GenerateAccessToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := u.tokenService.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := u.tokenService.StoreRefreshToken(ctx, userID, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func generateUsernameFromEmail(email string) string {
	for i, c := range email {
		if c == '@' {
			return email[:i]
		}
	}
	return email
}
