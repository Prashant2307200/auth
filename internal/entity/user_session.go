package entity

import "time"

// UserSession represents an active user session
type UserSession struct {
	ID         string    `json:"id"`
	UserID     int64     `json:"user_id"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Current    bool      `json:"current,omitempty"`
}
