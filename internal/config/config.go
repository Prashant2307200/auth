package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Secrets struct {
	AccessTokenSecret  string `yaml:"access_token_secret" env:"ACCESS_TOKEN_SECRET" env-required:"true" env-default:"kRbZnaqqtfk7xHklLwPEf/bt+OsvxEFgoNNiPkyoD8C9+4pHRbiPS6vdv1nGzPR+I3L6Sy9u3c/iPLD7fQqR7g=="`
	RefreshTokenSecret string `yaml:"refresh_token_secret" env:"REFRESH_TOKEN_SECRET" env-required:"true"`
	CookieSecret       string `yaml:"cookie_secret" env:"COOKIE_SECRET" env-required:"true"`
}

type Cloud struct {
	Name      string `yaml:"name" env:"NAME" env-required:"true" env-default:"dyrb0ytef"`
	ApiKey    string `yaml:"api_key" env:"API_KEY" env-required:"true" env-default:"888923427284328"`
	ApiSecret string `yaml:"api_secret" env:"API_SECRET" env-required:"true" env-default:"wDYpMIwf_og6dKfEgAfmqes1a9w"`
}

type HttpServer struct {
	Addr string `yaml:"address" env-required:"true"`
}

type Redis struct {
	Addr string `yaml:"address" env:"REDIS_ADDRESS" env-required:"true" env-default:"localhost:6379"`
	User string `yaml:"username" env:"REDIS_USERNAME" env-required:"true" env-default:""`
	Pass string `yaml:"password" env:"REDIS_PASSWORD" env-required:"true" env-default:""`
}

type Config struct {
	Secrets     Secrets    `yaml:"secrets"`
	Env         string     `yaml:"env" env:"ENV" env-required:"true" env-default:"dev"`
	HttpServer  HttpServer `yaml:"http_server"`
	Cloud       Cloud      `yaml:"cloud"`
	Redis       Redis      `yaml:"redis"`
	PostgresUri string     `yaml:"postgres_uri" env:"POSTGRES_URI" env-required:"true" env-default:"postgresql://auth_db_521h_user:O1Si4boEobcNGqzp5R0AjAt5mDBXeTeX@dpg-d1sf2f8dl3ps73a7su1g-a.singapore-postgres.render.com/auth_db_521h"`
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
