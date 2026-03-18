package dto

type UpdateUserRequest struct {
	Username   string `json:"username,omitempty" validate:"omitempty,min=3,max=20,username"`
	Email      string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName  string `json:"first_name,omitempty" validate:"omitempty,min=1,max=50"`
	LastName   string `json:"last_name,omitempty" validate:"omitempty,min=1,max=50"`
	Phone      string `json:"phone,omitempty" validate:"omitempty,e164"`
	ProfilePic string `json:"profile_pic,omitempty"`
}
