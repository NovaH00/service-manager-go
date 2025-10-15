package utils

import (
	"log"
	"os"
)

// GetEnv returns the value of the environment variable 'key'.
// If the variable is not set, it returns 'defaultValue' and logs a message.
func GetEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Environment variable '%s' is not set, using default value '%s'", key, defaultValue)
		return defaultValue
	}
	return value
}
