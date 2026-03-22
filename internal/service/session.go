package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix      = "session:"
	userSessionsPrefix = "user_sessions:"
	sessionTTL         = 7 * 24 * time.Hour
)

type SessionService struct {
	rdb *redis.Client
}

func NewSessionService(rdb *redis.Client) *SessionService {
	return &SessionService{rdb: rdb}
}

type sessionData struct {
	ID           string    `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *SessionService) CreateSession(ctx context.Context, userID int64, refreshToken, deviceInfo, ipAddress, userAgent string) (*entity.UserSession, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(sessionTTL)

	data := sessionData{
		ID:           sessionID,
		UserID:       userID,
		RefreshToken: refreshToken,
		DeviceInfo:   deviceInfo,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    now,
		LastUsedAt:   now,
		ExpiresAt:    expiresAt,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session data: %w", err)
	}

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, sessionPrefix+sessionID, jsonData, sessionTTL)
	pipe.SAdd(ctx, userSessionsPrefix+fmt.Sprint(userID), sessionID)
	pipe.Expire(ctx, userSessionsPrefix+fmt.Sprint(userID), sessionTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &entity.UserSession{
		ID:         sessionID,
		UserID:     userID,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
		LastUsedAt: now,
		ExpiresAt:  expiresAt,
	}, nil
}

func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*entity.UserSession, error) {
	jsonData, err := s.rdb.Get(ctx, sessionPrefix+sessionID).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var data sessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &entity.UserSession{
		ID:         data.ID,
		UserID:     data.UserID,
		DeviceInfo: data.DeviceInfo,
		IPAddress:  data.IPAddress,
		UserAgent:  data.UserAgent,
		CreatedAt:  data.CreatedAt,
		LastUsedAt: data.LastUsedAt,
		ExpiresAt:  data.ExpiresAt,
	}, nil
}

func (s *SessionService) ListUserSessions(ctx context.Context, userID int64) ([]*entity.UserSession, error) {
	sessionIDs, err := s.rdb.SMembers(ctx, userSessionsPrefix+fmt.Sprint(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list session IDs: %w", err)
	}

	var sessions []*entity.UserSession
	for _, sessionID := range sessionIDs {
		session, err := s.GetSession(ctx, sessionID)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *SessionService) RevokeSession(ctx context.Context, userID int64, sessionID string) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.UserID != userID {
		return fmt.Errorf("session does not belong to user")
	}

	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, sessionPrefix+sessionID)
	pipe.SRem(ctx, userSessionsPrefix+fmt.Sprint(userID), sessionID)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

func (s *SessionService) RevokeAllSessions(ctx context.Context, userID int64, exceptSessionID string) error {
	sessionIDs, err := s.rdb.SMembers(ctx, userSessionsPrefix+fmt.Sprint(userID)).Result()
	if err != nil {
		return fmt.Errorf("failed to list session IDs: %w", err)
	}

	pipe := s.rdb.Pipeline()
	for _, sessionID := range sessionIDs {
		if sessionID == exceptSessionID {
			continue
		}
		pipe.Del(ctx, sessionPrefix+sessionID)
		pipe.SRem(ctx, userSessionsPrefix+fmt.Sprint(userID), sessionID)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	return nil
}

func (s *SessionService) UpdateLastUsed(ctx context.Context, sessionID string) error {
	jsonData, err := s.rdb.Get(ctx, sessionPrefix+sessionID).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	var data sessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	data.LastUsedAt = time.Now()

	newData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	ttl := time.Until(data.ExpiresAt)
	if ttl <= 0 {
		ttl = sessionTTL
	}

	return s.rdb.Set(ctx, sessionPrefix+sessionID, newData, ttl).Err()
}

func ParseDeviceInfo(userAgent string) string {
	ua := strings.ToLower(userAgent)

	var browser, os string

	switch {
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		browser = "Chrome"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		browser = "Safari"
	case strings.Contains(ua, "edg"):
		browser = "Edge"
	default:
		browser = "Unknown Browser"
	}

	switch {
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "mac"):
		os = "macOS"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		os = "iOS"
	default:
		os = "Unknown OS"
	}

	return browser + " on " + os
}
