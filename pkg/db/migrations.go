package db

import (
	"database/sql"
	"fmt"
	"log/slog"
)

// RunMigrations runs all database migrations in order
func RunMigrations(db *sql.DB) error {
	if err := MigrateUsersTable(db); err != nil {
		return err
	}
	if err := MigrateBusinessesTable(db); err != nil {
		return err
	}
	if err := MigrateBusinessUsersTable(db); err != nil {
		return err
	}
	if err := MigrateBusinessInvitesTable(db); err != nil {
		return err
	}
	if err := MigrateBusinessDomainsTable(db); err != nil {
		return err
	}
	return nil
}

// MigrateUsersTable creates the users table if it doesn't exist
func MigrateUsersTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(20) NOT NULL CHECK (char_length(username) >= 3),
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL CHECK (char_length(password) >= 6),
		profile_pic TEXT DEFAULT '',
		role INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);",
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			slog.Warn("Failed to create index", slog.String("index", idx), slog.Any("error", err))
		}
	}
	slog.Info("Users table migration completed successfully")
	return nil
}

// MigrateBusinessesTable creates the businesses table if it doesn't exist
func MigrateBusinessesTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS businesses (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		slug VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL,
		owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		signup_policy VARCHAR(20) DEFAULT 'closed',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create businesses table: %w", err)
	}
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_businesses_slug ON businesses(slug);",
		"CREATE INDEX IF NOT EXISTS idx_businesses_owner_id ON businesses(owner_id);",
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			slog.Warn("Failed to create index", slog.String("index", idx), slog.Any("error", err))
		}
	}
	slog.Info("Businesses table migration completed successfully")
	return nil
}

// MigrateBusinessUsersTable creates the business_users junction table if it doesn't exist
func MigrateBusinessUsersTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS business_users (
		id SERIAL PRIMARY KEY,
		business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(business_id, user_id)
	);
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create business_users table: %w", err)
	}
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_business_users_business_id ON business_users(business_id);",
		"CREATE INDEX IF NOT EXISTS idx_business_users_user_id ON business_users(user_id);",
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			slog.Warn("Failed to create index", slog.String("index", idx), slog.Any("error", err))
		}
	}
	slog.Info("Business_users table migration completed successfully")
	return nil
}

// MigrateBusinessInvitesTable creates the business_invites table if it doesn't exist
func MigrateBusinessInvitesTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS business_invites (
		id SERIAL PRIMARY KEY,
		business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
		email VARCHAR(255) NOT NULL,
		role INTEGER NOT NULL DEFAULT 0,
		invited_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token VARCHAR(64) NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		accepted_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create business_invites table: %w", err)
	}
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_business_invites_token ON business_invites(token);",
		"CREATE INDEX IF NOT EXISTS idx_business_invites_business_email ON business_invites(business_id, email);",
		"CREATE INDEX IF NOT EXISTS idx_business_invites_expires_at ON business_invites(expires_at);",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_business_invites_pending_unique ON business_invites(business_id, email) WHERE status = 'pending';",
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			slog.Warn("Failed to create index", slog.String("index", idx), slog.Any("error", err))
		}
	}
	slog.Info("Business_invites table migration completed successfully")
	return nil
}

// MigrateBusinessDomainsTable creates the business_domains table if it doesn't exist
func MigrateBusinessDomainsTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS business_domains (
		id SERIAL PRIMARY KEY,
		business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
		domain VARCHAR(255) NOT NULL,
		verified BOOLEAN DEFAULT false,
		auto_join_enabled BOOLEAN DEFAULT false,
		verification_token VARCHAR(64),
		verified_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(business_id, domain)
	);
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create business_domains table: %w", err)
	}
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_business_domains_domain ON business_domains(domain);",
		"CREATE INDEX IF NOT EXISTS idx_business_domains_verified_auto ON business_domains(domain, verified, auto_join_enabled);",
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			slog.Warn("Failed to create index", slog.String("index", idx), slog.Any("error", err))
		}
	}
	slog.Info("Business_domains table migration completed successfully")
	return nil
}
