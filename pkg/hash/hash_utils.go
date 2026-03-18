package hash

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

const CurrentCost = 12

// NeedsRehash returns true if the stored hash is not bcrypt or has lower cost
func NeedsRehash(hashedPassword string) (bool, error) {
	if hashedPassword == "" {
		return true, fmt.Errorf("empty hash")
	}

	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		// Not a valid bcrypt hash -> needs rehash
		return true, nil
	}

	if cost < CurrentCost {
		return true, nil
	}
	return false, nil
}

// HashPasswordWithCost creates a bcrypt hash using the provided cost.
// Exported for tests and rare cases; prefer HashPassword which uses CurrentCost.
func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash with cost %d: %w", cost, err)
	}
	return string(bytes), nil
}
