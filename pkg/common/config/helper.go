package config

import (
	"os"
	"strconv"
	"strings"
)

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	value := getEnv(key, "")

	if value, err := strconv.Atoi(value); err == nil {
		return value
	}

	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := getEnv(key, "")

	if value, err := strconv.ParseBool(value); err == nil {
		return value
	}

	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, sep string) []string {
	value := getEnv(key, "")

	if value == "" {
		return defaultValue
	}

	return strings.Split(value, sep)
}
