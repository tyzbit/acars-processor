package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func ReadFile(filePath string) []byte {
	// Read the content of the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error(Attention("Error reading the file: %v", err))
			os.Exit(1)
		}
	}
	return content
}

func WriteFile(filePath string, contents []byte) {
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error(Attention("Error writing file: %s", err))
	}
}

// Saves a file and returns true if the file changed
func UpdateFile(filePath string, contents []byte) (changed bool) {
	file := ReadFile(filePath)
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error(Attention("Error writing file: %s", err))
	}
	return string(file) != string(contents)
}
