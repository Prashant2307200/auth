package dto

type RegisterRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=20,username"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	FirstName    string `json:"first_name,omitempty" validate:"omitempty,min=1,max=50"`
	LastName     string `json:"last_name,omitempty" validate:"omitempty,min=1,max=50"`
	Phone        string `json:"phone,omitempty" validate:"omitempty,e164"`
	ProfilePic   string `json:"profile_pic,omitempty"`
	Role         int    `json:"role,omitempty"`
	InviteToken  string `json:"invite_token,omitempty"`
	BusinessSlug string `json:"business_slug,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type ProfileUpdateRequest struct {
	Username   string `json:"username,omitempty" validate:"omitempty,min=3,max=20"`
	Email      string `json:"email,omitempty" validate:"omitempty,email"`
	ProfilePic string `json:"profile_pic,omitempty"`
}
