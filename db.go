package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	sqlitePath = "./messages.db"
)

func LoadSavedMessages() error {
	// Increase verbosity of the database if the loglevel is higher than Info
	var logConfig logger.Interface
	if log.GetLevel() > log.DebugLevel {
		logConfig = logger.Default.LogMode(logger.Info)
	}

	// Create the folder path if it doesn't exist
	var err error
	_, err = os.Stat(sqlitePath)
	if errors.Is(err, fs.ErrNotExist) {
		dirPath := filepath.Dir(sqlitePath)
		if err := os.MkdirAll(dirPath, 0660); err != nil {
			log.Error(yo.Uhh("unable to make directory path %s, err: %w", dirPath, err).FRFR())
		}
	}

	if !config.ACARSProcessorSettings.SaveMessages {
		log.Info(yo.FYI("Database is not enabled").FRFR())
		sqlitePath = "file::memory:?cache=shared"
	} else {
		p := config.ACARSProcessorSettings.SQLiteDatabasePath
		log.Info(yo.Bet("Database is enabled at path %s", p).FRFR())
		if p != "" {
			sqlitePath = config.ACARSProcessorSettings.SQLiteDatabasePath
		}
	}
	db, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{Logger: logConfig})

	if err != nil {
		return err
	}
	tx := db.Begin()
	// ACARS
	am := []ACARSMessage{}
	if err := tx.AutoMigrate(ACARSMessage{}); err != nil {
		log.Fatal(yo.Uhh("Unable to automigrate ACARSMessage type: %s", err).FRFR())
	}
	for _, a := range am {
		ACARSMessageQueue <- a.ID
	}
	log.Info(yo.FYI("Loaded %d ACARS messages from the db", len(am)).FRFR())

	// VDLM2
	vm := []VDLM2Message{}
	if err := tx.AutoMigrate(VDLM2Message{}); err != nil {
		log.Fatal(yo.Uhh("Unable to automigrate VDLM2Message type: %s", err).FRFR())
	}
	tx.Find(&vm)
	for _, v := range vm {
		VDLM2MessageQueue <- v.ID
	}
	log.Info(yo.FYI("Loaded %d VDLM2 messages from the db", len(vm)).FRFR())
	tx.Commit()
	return nil
}
