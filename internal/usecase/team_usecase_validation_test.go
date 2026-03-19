package usecase

import "testing"

func TestValidateInviteEmail_Valid(t *testing.T) {
	u := &teamUsecase{}
	cases := []string{
		"user@example.com",
		"a@b.co",
		"name.surname+tag@sub.domain.org",
	}
	for _, c := range cases {
		if err := u.ValidateInviteEmail(c); err != nil {
			t.Fatalf("expected valid email %s, got error: %v", c, err)
		}
	}
}

func TestValidateInviteEmail_Invalid(t *testing.T) {
	u := &teamUsecase{}
	cases := []string{
		"",
		"plainaddress",
		"noatsymbol.com",
		"@missinglocal",
	}
	for _, c := range cases {
		if err := u.ValidateInviteEmail(c); err == nil {
			t.Fatalf("expected invalid email %s to error", c)
		}
	}
}

func TestValidateRole_Valid(t *testing.T) {
	u := &teamUsecase{}
	for i := 1; i <= 4; i++ {
		if err := u.ValidateRole(i); err != nil {
			t.Fatalf("expected valid role %d, got error: %v", i, err)
		}
	}
}

func TestValidateRole_Invalid(t *testing.T) {
	u := &teamUsecase{}
	cases := []int{0, 5, -1, 100}
	for _, c := range cases {
		if err := u.ValidateRole(c); err == nil {
			t.Fatalf("expected invalid role %d to error", c)
		}
	}
}
