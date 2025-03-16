package utils

import (
	"os"
)

// GetEnv は環境変数を取得する
func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
