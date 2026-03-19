package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Secrets struct {
	AccessTokenSecret  string `yaml:"access_token_secret" env:"ACCESS_TOKEN_SECRET" env-required:"true"`
	RefreshTokenSecret string `yaml:"refresh_token_secret" env:"REFRESH_TOKEN_SECRET" env-required:"true"`
	CookieSecret       string `yaml:"cookie_secret" env:"COOKIE_SECRET" env-required:"true"`
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

type Config struct {
	Secrets     Secrets    `yaml:"secrets"`
	Env         string     `yaml:"env" env:"ENV" env-required:"true" env-default:"dev"`
	HttpServer  HttpServer `yaml:"http_server"`
	Cloud       Cloud      `yaml:"cloud"`
	Redis       Redis      `yaml:"redis"`
	PostgresUri string     `yaml:"postgres_uri" env:"POSTGRES_URI" env-required:"true"`
}

func MustLoad() *Config {
	var configPath string

	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" {
		flags := flag.String("config", "", "path to config file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			log.Fatal("Config path is not set")
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("Cannot read config: %s", err.Error())
	}

	return &cfg
}
