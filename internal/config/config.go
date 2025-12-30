package config

import "os"

type Config struct {
	DBHost string
	DBUser string
	DBPass string
	DBName string
	DBPort string
}

func Load() *Config {
	return &Config{
		DBHost: os.Getenv("POSTGRES_HOST"),
		DBUser: os.Getenv("POSTGRES_USER"),
		DBPass: os.Getenv("POSTGRES_PASSWORD"),
		DBName: os.Getenv("POSTGRES_DB"),
		DBPort: os.Getenv("POSTGRES_PORT"),
	}

}
