package config

import (
	"log"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerAddr   string
	DatabasePath string
	Token        string
	SteamApiKey  string
}

var Instance *AppConfig = new()

func new() *AppConfig {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Print(".env file is not located, skipping\n")
	}

	return &AppConfig{
		ServerAddr:   getEnv("SERVER_ADDR", "127.0.0.1:3000"),
		DatabasePath: getEnv("DB_PATH", "store.sqlite"),
		Token:        getEnv("SECRET_TOKEN", ""),
		SteamApiKey:  getEnv("STEAM_API_KEY", ""),
	}
}
