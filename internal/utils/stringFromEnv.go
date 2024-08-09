package utils

import "os"

func StringFromEnv(key string, defaultValue string) string {
	env, exists := os.LookupEnv(key)
	if !exists {
		env = defaultValue
	}
	return env
}
