package entity

import "time"

// AuditLog represents an immutable audit record for actions performed within a business
type AuditLog struct {
	ID         int64                  `json:"id"`
	BusinessID int64                  `json:"business_id"`
	UserID     int64                  `json:"user_id"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   *int64                 `json:"entity_id,omitempty"`
	OldValues  map[string]interface{} `json:"old_values,omitempty"`
	NewValues  map[string]interface{} `json:"new_values,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at,omitempty"`
	UpdatedAt  time.Time              `json:"updated_at,omitempty"`
}
