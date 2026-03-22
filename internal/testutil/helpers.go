package testutil

import (
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/seeder"
)

// CreateTestUser creates a test user matching seeder.SeedUsersData[1] (testuser)
func CreateTestUser() *entity.User {
	s := seeder.SeedUsersData[1]
	return &entity.User{
		ID:         1,
		Username:   s.Username,
		Email:      s.Email,
		Password:   "hashedpassword",
		ProfilePic: "",
		Role:       s.Role,
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

// CreateTestAdmin creates an admin user matching seeder.SeedUsersData[0] (admin)
func CreateTestAdmin() *entity.User {
	s := seeder.SeedUsersData[0]
	return &entity.User{
		ID:         1,
		Username:   s.Username,
		Email:      s.Email,
		Password:   "hashedpassword",
		ProfilePic: "",
		Role:       s.Role,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestAdminWithID creates an admin user with a specific ID
func CreateTestAdminWithID(id int64) *entity.User {
	admin := CreateTestAdmin()
	admin.ID = id
	return admin
}

// CreateTestBusiness creates a business matching seeder.SeedBusinessesData[0] (Acme)
func CreateTestBusiness() *entity.Business {
	s := seeder.SeedBusinessesData[0]
	return &entity.Business{
		ID:           1,
		Name:         s.Name,
		Slug:         s.Slug,
		Email:        s.Email,
		OwnerID:      1,
		SignupPolicy: s.SignupPolicy,
	}
}

// CreateTestBusinessWithSignupPolicy creates a business with custom slug and signup policy
func CreateTestBusinessWithSignupPolicy(slug, policy string) *entity.Business {
	b := CreateTestBusiness()
	b.Slug = slug
	b.SignupPolicy = policy
	return b
}

// CreateTestInvite creates a test invite; email/token/status from seeder.SeedInvitesData when applicable
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

// CreateTestBusinessWithID creates a business with a specific ID (uses seed Acme data)
func CreateTestBusinessWithID(id int64) *entity.Business {
	b := CreateTestBusiness()
	b.ID = id
	return b
}

// CreateTestDomain creates a domain matching seeder.SeedDomainsData when verified/autoJoin match
func CreateTestDomain(businessID int64, domain string, verified, autoJoin bool) *entity.BusinessDomain {
	d := &entity.BusinessDomain{
		ID:              1,
		BusinessID:      businessID,
		Domain:          domain,
		Verified:        verified,
		AutoJoinEnabled: autoJoin,
	}
	if verified {
		now := time.Now()
		d.VerifiedAt = &now
	}
	return d
}

// CreateTestUserList returns users with given IDs (from seed testuser base)
func CreateTestUserList(ids ...int64) []*entity.User {
	out := make([]*entity.User, len(ids))
	for i, id := range ids {
		out[i] = CreateTestUserWithID(id)
	}
	return out
}
