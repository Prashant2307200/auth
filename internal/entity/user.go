package entity

import "time"

type User struct {
	ID         int64     `json:"id"`
	Username   string    `json:"username" form:"username" validate:"required,min=3,max=20"`
	Email      string    `json:"email" form:"email" validate:"required,email"`
	Password   string    `json:"password" form:"password" validate:"required,min=6"`
	ProfilePic string    `json:"profile_pic" form:"profile_pic"`
	Role       int       `json:"role" form:"role"`
	BusinessID *int64    `json:"business_id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
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
