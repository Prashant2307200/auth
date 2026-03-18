package entity

import (
	"time"
)

const RoleUser = 0
const RoleAdmin = 1

// legacy numeric roles remain for backward compatibility

type User struct {
	ID         int64  `json:"id"`
	Username   string `json:"username" form:"username" validate:"required,min=3,max=20"`
	Email      string `json:"email" form:"email" validate:"required,email"`
	Password   string `json:"-" form:"password" validate:"required,min=6"`
	ProfilePic string `json:"profile_pic" form:"profile_pic"`
	Role       int    `json:"role" form:"role"`
	BusinessID *int64 `json:"business_id,omitempty"`

	TenantID  int64      `json:"tenant_id,omitempty"`
	RoleName  string     `json:"role_name,omitempty"`
	Status    string     `json:"status,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=20"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
	ProfilePic   string `json:"profile_pic"`
	Role         int    `json:"role"`
	InviteToken  string `json:"invite_token,omitempty"`
	BusinessSlug string `json:"business_slug,omitempty"`
}
