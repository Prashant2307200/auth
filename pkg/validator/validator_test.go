package validator

import "testing"

func TestValidateEmail(t *testing.T) {
	ok, err := ValidateEmail("test@example.com")
	if !ok || err != nil {
		t.Fatalf("expected valid email, got err=%v", err)
	}
	ok, _ = ValidateEmail("invalid-email")
	if ok {
		t.Fatalf("expected invalid email")
	}
}

func TestValidatePassword(t *testing.T) {
	// valid
	ok, err := ValidatePassword("Abcdef1!")
	if !ok || err != nil {
		t.Fatalf("expected valid password, got %v", err)
	}
	// too short
	ok, _ = ValidatePassword("Ab1!")
	if ok {
		t.Fatalf("expected invalid password (too short)")
	}
	// missing special
	ok, _ = ValidatePassword("Abcdef12")
	if ok {
		t.Fatalf("expected invalid password (no special)")
	}
}

func TestValidateUsername(t *testing.T) {
	ok, err := ValidateUsername("user_name")
	if !ok || err != nil {
		t.Fatalf("expected valid username, got %v", err)
	}
	ok, _ = ValidateUsername("ab")
	if ok {
		t.Fatalf("expected invalid username (too short)")
	}
	ok, _ = ValidateUsername("user-with-dash")
	if ok {
		t.Fatalf("expected invalid username (dash not allowed)")
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	ok, err := ValidatePhoneNumber("+14155552671")
	if !ok || err != nil {
		t.Fatalf("expected valid phone, got %v", err)
	}
	ok, _ = ValidatePhoneNumber("4155552671")
	if ok {
		t.Fatalf("expected invalid phone")
	}
}

func TestValidateProfilePicURL(t *testing.T) {
	ok, err := ValidateProfilePicURL("")
	if !ok || err != nil {
		t.Fatalf("expected empty profile pic to be valid, got %v", err)
	}

	ok, err = ValidateProfilePicURL("https://res.cloudinary.com/demo/image/upload/v1/pic.jpg")
	if !ok || err != nil {
		t.Fatalf("expected valid profile pic url, got %v", err)
	}

	ok, _ = ValidateProfilePicURL("javascript:alert(1)")
	if ok {
		t.Fatalf("expected invalid profile pic url scheme")
	}

	ok, _ = ValidateProfilePicURL("not-a-url")
	if ok {
		t.Fatalf("expected invalid profile pic url format")
	}
}
