package config

type AppConfig struct {
	ServerAddr   string
	DatabasePath string
	Token        string
}

var Instance *AppConfig = new()

func new() *AppConfig {
	return &AppConfig{
		ServerAddr:   getEnv("SERVER_ADDR", "0.0.0.0:3000"),
		DatabasePath: getEnv("DB_PATH", "store.sqlite"),
		Token:        getEnv("SECRET_TOKEN", ""),
	}
}
