package config

import (
	"os"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
)

// Ensure that Redis env vars are required and no defaults are used.
func TestRedisEnvRequired_FailsWhenMissing(t *testing.T) {
	// Ensure variables are not set
	os.Unsetenv("REDIS_ADDRESS")
	os.Unsetenv("REDIS_USERNAME")
	os.Unsetenv("REDIS_PASSWORD")

	var r Redis
	err := cleanenv.ReadEnv(&r)
	if err == nil {
		t.Fatalf("expected error when required REDIS_* env vars are missing, got nil")
	}
}

func TestRedisEnvRequired_SucceedsWhenSet(t *testing.T) {
	// Set minimal required env vars
	os.Setenv("REDIS_ADDRESS", "localhost:6379")
	defer os.Unsetenv("REDIS_ADDRESS")
	os.Setenv("REDIS_USERNAME", "")
	defer os.Unsetenv("REDIS_USERNAME")
	os.Setenv("REDIS_PASSWORD", "")
	defer os.Unsetenv("REDIS_PASSWORD")

	var r Redis
	if err := cleanenv.ReadEnv(&r); err != nil {
		t.Fatalf("expected no error when REDIS_* env vars are set, got: %v", err)
	}
}

func TestMissingRequiredEnvVars(t *testing.T) {
	t.Run("returns missing vars", func(t *testing.T) {
		os.Setenv("CFG_TEST_PRESENT", "1")
		defer os.Unsetenv("CFG_TEST_PRESENT")
		missing := missingRequiredEnvVars([]string{"CFG_TEST_PRESENT", "CFG_TEST_MISSING"})
		if len(missing) != 1 || missing[0] != "CFG_TEST_MISSING" {
			t.Fatalf("expected CFG_TEST_MISSING, got %v", missing)
		}
	})

	t.Run("returns none when all present", func(t *testing.T) {
		os.Setenv("CFG_TEST_PRESENT_A", "1")
		defer os.Unsetenv("CFG_TEST_PRESENT_A")
		os.Setenv("CFG_TEST_PRESENT_B", "1")
		defer os.Unsetenv("CFG_TEST_PRESENT_B")

		missing := missingRequiredEnvVars([]string{"CFG_TEST_PRESENT_A", "CFG_TEST_PRESENT_B"})
		if len(missing) != 0 {
			t.Fatalf("expected no missing vars, got %v", missing)
		}
	})
}
