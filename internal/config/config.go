package config

import (
	"flag"
	"os"
	"strings"
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
)

type Secrets struct {
	// Optional: access JWTs are signed with RSA (see jwt key paths). Kept for YAML compatibility / future use.
	AccessTokenSecret string `yaml:"access_token_secret" env:"ACCESS_TOKEN_SECRET"`
	// Required: used for refresh token signing (HS256).
	RefreshTokenSecret string `yaml:"refresh_token_secret" env:"REFRESH_TOKEN_SECRET" env-required:"true"`
	// Optional: reserved for signed/encrypted cookies; HttpOnly cookies are used without signing today.
	CookieSecret string `yaml:"cookie_secret" env:"COOKIE_SECRET"`
}

type Cloud struct {
	Name      string `yaml:"name" env:"NAME" env-required:"true"`
	ApiKey    string `yaml:"api_key" env:"API_KEY" env-required:"true"`
	ApiSecret string `yaml:"api_secret" env:"API_SECRET" env-required:"true"`
}

type HttpServer struct {
	Addr string `yaml:"address" env-required:"true"`
}

type Redis struct {
	Addr string `yaml:"address" env:"REDIS_ADDRESS" env-required:"true"`
	User string `yaml:"username" env:"REDIS_USERNAME" env-required:"true"`
	Pass string `yaml:"password" env:"REDIS_PASSWORD" env-required:"true"`
}

type Email struct {
	APIKey    string `yaml:"api_key" env:"MAILEROO_API_KEY"`
	FromEmail string `yaml:"from_email" env:"MAILEROO_FROM_EMAIL"`
	FromName  string `yaml:"from_name" env:"MAILEROO_FROM_NAME"`
	BaseURL   string `yaml:"base_url" env:"APP_BASE_URL"`
}

type OAuth struct {
	GoogleClientID     string `yaml:"google_client_id" env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `yaml:"google_client_secret" env:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `yaml:"google_redirect_url" env:"GOOGLE_REDIRECT_URL"`
}

type MFA struct {
	// EncryptionKey is a 32-byte hex-encoded key used for AES-256-GCM encryption of TOTP secrets.
	// Required when MFA is enabled. Generate with: openssl rand -hex 32
	EncryptionKey string `yaml:"encryption_key" env:"MFA_ENCRYPTION_KEY"`
}

type Config struct {
	Secrets     Secrets    `yaml:"secrets"`
	Env         string     `yaml:"env" env:"ENV" env-required:"true" env-default:"dev"`
	HttpServer  HttpServer `yaml:"http_server"`
	Cloud       Cloud      `yaml:"cloud"`
	Redis       Redis      `yaml:"redis"`
	Email       Email      `yaml:"email"`
	OAuth       OAuth      `yaml:"oauth"`
	MFA         MFA        `yaml:"mfa"`
	PostgresUri string     `yaml:"postgres_uri" env:"POSTGRES_URI" env-required:"true"`
	// Optional JWT key paths; if empty, code may fall back to legacy defaults.
	JWT struct {
		PublicKeyPath  string `yaml:"public_key_path" env:"JWT_PUBLIC_KEY_PATH"`
		PrivateKeyPath string `yaml:"private_key_path" env:"JWT_PRIVATE_KEY_PATH"`
	} `yaml:"jwt"`
}

func MustLoad() *Config {
	var configPath string

	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" {
		flags := flag.String("config", "", "path to config file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			slog.Error("Config path is not set")
			os.Exit(1)
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Error("Config file does not exist", slog.String("path", configPath))
		os.Exit(1)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		slog.Error("Cannot read config", slog.Any("error", err))
		os.Exit(1)
	}

	missing := missingRequiredEnvVars([]string{
		"POSTGRES_URI",
		"REDIS_ADDRESS",
		"REDIS_USERNAME",
		"REDIS_PASSWORD",
		"REFRESH_TOKEN_SECRET",
		"NAME",
		"API_KEY",
		"API_SECRET",
	})
	if len(missing) > 0 {
		slog.Error("Missing required environment variables", slog.String("missing", strings.Join(missing, ", ")))
		os.Exit(1)
	}

	return &cfg
}

func missingRequiredEnvVars(vars []string) []string {
	missing := make([]string, 0)
	for _, v := range vars {
		if _, ok := os.LookupEnv(v); !ok {
			missing = append(missing, v)
		}
	}
	return missing
}
