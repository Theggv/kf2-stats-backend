package config

import (
	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerAddr  string
	Token       string
	SteamApiKey string
	Domain      string

	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     int
	DBName     string

	JwtAccessSecretKey  string
	JwtAccessExpiresIn  string
	JwtRefreshSecretKey string
	JwtRefreshExpiresIn string
}

var Instance *AppConfig = new()

func new() *AppConfig {
	godotenv.Load(".env")

	config := AppConfig{
		ServerAddr:  getEnv("SERVER_ADDR", "127.0.0.1:3000"),
		Token:       getEnv("SECRET_TOKEN", ""),
		SteamApiKey: getEnv("STEAM_API_KEY", ""),
		Domain:      getEnv("DOMAIN", "localhost"),

		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "db"),
		DBPort:     getEnvAsInt("DB_PORT", 3306),
		DBName:     getEnv("DB_NAME", "stats"),

		JwtAccessSecretKey:  getEnv("JWT_ACCESS_SECRET_KEY", ""),
		JwtAccessExpiresIn:  getEnv("JWT_ACCESS_EXPIRES_IN", "15m"),
		JwtRefreshSecretKey: getEnv("JWT_REFRESH_SECRET_KEY", ""),
		JwtRefreshExpiresIn: getEnv("JWT_REFRESH_EXPIRES_IN", "30d"),
	}

	if config.Token == "" {
		panic("SECRET_TOKEN is not set. Check your .env file.")
	}

	if config.JwtAccessSecretKey == "" || config.JwtAccessSecretKey == "long_access_token_secret_key" {
		panic("JWT_ACCESS_SECRET_KEY is not set. Check your .env file.")
	}

	if config.JwtRefreshSecretKey == "" || config.JwtAccessSecretKey == "long_refresh_token_secret_key" {
		panic("JWT_REFRESH_SECRET_KEY is not set. Check your .env file.")
	}

	return &config
}
