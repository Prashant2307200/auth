package seeder

type UserSeed struct {
	Username string
	Email    string
	Password string // plaintext, hashed at seed time
	Role     int
}

var SeedUsersData = []UserSeed{
	{Username: "admin", Email: "admin@example.com", Password: "admin123", Role: 1},
	{Username: "testuser", Email: "test@example.com", Password: "test123", Role: 0},
	{Username: "demo", Email: "demo@example.com", Password: "demo123", Role: 0},
}

type BusinessSeed struct {
	Name         string
	Slug         string
	Email        string
	OwnerEmail   string
	SignupPolicy string
}

var SeedBusinessesData = []BusinessSeed{
	{Name: "Acme Inc", Slug: "acme-inc", Email: "owner@acme.com", OwnerEmail: "admin@example.com", SignupPolicy: "closed"},
	{Name: "Demo Corp", Slug: "demo-corp", Email: "demo@democorp.com", OwnerEmail: "demo@example.com", SignupPolicy: "open"},
}

type InviteSeed struct {
	BusinessSlug   string
	Email          string
	Role           int
	InvitedByEmail string
	Token          string
}

var SeedInvitesData = []InviteSeed{
	{BusinessSlug: "acme-inc", Email: "invited@acme.com", Role: 0, InvitedByEmail: "admin@example.com", Token: "inv-token-acme-1"},
	{BusinessSlug: "demo-corp", Email: "employee@democorp.com", Role: 0, InvitedByEmail: "demo@example.com", Token: "inv-token-demo-1"},
}

type DomainSeed struct {
	BusinessSlug    string
	Domain          string
	Verified        bool
	AutoJoinEnabled bool
}

var SeedDomainsData = []DomainSeed{
	{BusinessSlug: "acme-inc", Domain: "acme.com", Verified: true, AutoJoinEnabled: true},
}
