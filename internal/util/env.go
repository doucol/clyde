package util

import (
	"fmt"
	"os"
	"strings"
)

// Helper functions for environment variables
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return fallback
}

func GetEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		result := strings.ToLower(value)
		return result == "true" || result == "yes" || result == "1"
	}
	return fallback
}
