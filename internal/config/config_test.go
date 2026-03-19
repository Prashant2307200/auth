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
