package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func readFile(filePath string) []byte {
	filePath = os.Getenv("HOME") + "/" + filePath
	// Read the content of the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("Error reading the file: %v", err)
			os.Exit(1)
		}
	}
	return content
}

func writeFile(filePath string, contents []byte) {
	filePath = os.Getenv("HOME") + "/" + filePath
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error("Error writing file: %w", err)
	}
}
