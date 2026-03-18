package entity

import "time"

type BusinessMember struct {
	ID         int64      `json:"id"`
	BusinessID int64      `json:"business_id"`
	UserID     *int64     `json:"user_id,omitempty"`
	Email      string     `json:"email,omitempty"`
	RoleID     int64      `json:"role_id"`
	Status     string     `json:"status"`
	InvitedBy  *int64     `json:"invited_by,omitempty"`
	InvitedAt  time.Time  `json:"invited_at,omitempty"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at,omitempty"`
}
