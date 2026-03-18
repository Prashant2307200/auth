package entity

import "time"

const (
	SignupPolicyClosed     = "closed"
	SignupPolicyOpen       = "open"
	SignupPolicyInviteOnly = "invite_only"
)

type Business struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name" validate:"required,min=2,max=100"`
	Slug         string     `json:"slug" validate:"required,min=2,max=50"`
	Email        string     `json:"email" validate:"required,email"`
	OwnerID      int64      `json:"owner_id"`
	Plan         string     `json:"plan,omitempty"`
	CreatedByID  int64      `json:"created_by_id,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	SignupPolicy string     `json:"signup_policy"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`
}

type BusinessUser struct {
	ID         int64     `json:"id"`
	BusinessID int64     `json:"business_id"`
	UserID     int64     `json:"user_id"`
	Role       int       `json:"role"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

const (
	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRevoked  = "revoked"
	InviteStatusExpired  = "expired"
)

type BusinessInvite struct {
	ID         int64      `json:"id"`
	BusinessID int64      `json:"business_id"`
	Email      string     `json:"email"`
	Role       int        `json:"role"`
	InvitedBy  int64      `json:"invited_by"`
	Token      string     `json:"token"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Status     string     `json:"status"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at,omitempty"`
}

type BusinessDomain struct {
	ID                int64      `json:"id"`
	BusinessID        int64      `json:"business_id"`
	Domain            string     `json:"domain"`
	Verified          bool       `json:"verified"`
	AutoJoinEnabled   bool       `json:"auto_join_enabled"`
	VerificationToken string     `json:"verification_token,omitempty"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at,omitempty"`
}
