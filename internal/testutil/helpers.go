package testutil

import (
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

// CreateTestUser creates a test user with default values
func CreateTestUser() *entity.User {
	return &entity.User{
		ID:         1,
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "hashedpassword",
		ProfilePic: "",
		Role:       0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestUserWithID creates a test user with a specific ID
func CreateTestUserWithID(id int64) *entity.User {
	user := CreateTestUser()
	user.ID = id
	return user
}

// CreateTestUserWithEmail creates a test user with a specific email
func CreateTestUserWithEmail(email string) *entity.User {
	user := CreateTestUser()
	user.Email = email
	return user
}

func CreateTestBusiness() *entity.Business {
	return &entity.Business{
		ID:      1,
		Name:    "Acme Inc",
		Slug:    "acme-inc",
		Email:   "owner@acme.com",
		OwnerID: 1,
	}
}

func CreateTestBusinessWithSignupPolicy(slug, policy string) *entity.Business {
	b := CreateTestBusiness()
	b.Slug = slug
	b.SignupPolicy = policy
	return b
}

func CreateTestInvite(businessID int64, email, token, status string, expiresAt time.Time) *entity.BusinessInvite {
	return &entity.BusinessInvite{
		ID:         1,
		BusinessID: businessID,
		Email:      email,
		Role:       0,
		InvitedBy:  1,
		Token:      token,
		ExpiresAt:  expiresAt,
		Status:     status,
	}
}
