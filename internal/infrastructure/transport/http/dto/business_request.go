package dto

type BusinessCreateRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Slug  string `json:"slug" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
}

type BusinessUpdateRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Slug  string `json:"slug,omitempty" validate:"omitempty,min=2,max=50"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}
