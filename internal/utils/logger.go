package utils

import (
	"log"
	"os"
)

// LogError はエラーログを記録する
func LogError(message string) {
	f, err := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open error.log: ", err)
		return
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println(message)
}
