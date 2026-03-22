package entity

import "time"

// UserMFA stores MFA configuration for a user
type UserMFA struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	SecretEncrypted  string     `json:"-"`
	BackupCodesHash  []string   `json:"-"`
	EnabledAt        *time.Time `json:"enabled_at,omitempty"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (m *UserMFA) IsEnabled() bool {
	return m.EnabledAt != nil
}
