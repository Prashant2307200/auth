package entity

import "time"

type Role struct {
	ID          int64     `json:"id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
