package config

import (
	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerAddr  string
	Token       string
	SteamApiKey string

	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     int
	DBName     string
}

var Instance *AppConfig = new()

func new() *AppConfig {
	godotenv.Load(".env")

	return &AppConfig{
		ServerAddr:  getEnv("SERVER_ADDR", "127.0.0.1:3000"),
		Token:       getEnv("SECRET_TOKEN", ""),
		SteamApiKey: getEnv("STEAM_API_KEY", ""),

		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "db"),
		DBPort:     getEnvAsInt("DB_PORT", 3306),
		DBName:     getEnv("DB_NAME", "stats"),
	}
}
