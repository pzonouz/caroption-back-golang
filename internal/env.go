package env

import "os"

func GetString(key string, fallback string) string {
	stringKey, ok := os.LookupEnv(key)
	if ok {
		return stringKey
	}

	return fallback
}
