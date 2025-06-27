package main

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	sqlitePath = "./messages.db"
)

func InitSQLite(l logger.Interface) (err error) {
	// Create the folder path if it doesn't exist
	_, err = os.Stat(sqlitePath)
	if errors.Is(err, fs.ErrNotExist) {
		dirPath := filepath.Dir(sqlitePath)
		if err := os.MkdirAll(dirPath, 0660); err != nil {
			return err
		}
	}
	if !config.ACARSProcessorSettings.Database.Enabled {
		log.Info(yo.FYI("Database is not enabled").FRFR())
		sqlitePath = "file::memory:?cache=shared"
	} else {
		p := config.ACARSProcessorSettings.Database.SQLiteDatabasePath
		log.Info(yo.Bet("Database is enabled at path %s", p).FRFR())
		if p != "" {
			sqlitePath = config.ACARSProcessorSettings.Database.SQLiteDatabasePath
		}
	}
	db, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{Logger: l})
	return err
}

func InitMariaDB(l logger.Interface) (err error) {
	dsn := config.ACARSProcessorSettings.Database.ConnectionString
	if !config.ACARSProcessorSettings.Database.Enabled {
		if err = InitSQLite(l); err != nil {
			return err
		}
	} else {
		p := config.ACARSProcessorSettings.Database.SQLiteDatabasePath
		log.Info(yo.Bet("Database is enabled at path %s", p).FRFR())
		if p != "" {
			sqlitePath = config.ACARSProcessorSettings.Database.SQLiteDatabasePath
		}
	}
	u, err := url.Parse(dsn)
	if err != nil {
		log.Panic(yo.Uhh("unable to parse mariadb connection string: %s", err))
	}
	q := u.Query()
	q.Set("parseTime", "True")
	u.RawQuery = q.Encode()
	db, err = gorm.Open(mysql.Open(u.String()), &gorm.Config{Logger: l})
	return err
}

func LoadSavedMessages() error {
	// Increase verbosity of the database if the loglevel is higher than Info
	var logConfig logger.Interface
	if log.GetLevel() > log.DebugLevel {
		logConfig = logger.Default.LogMode(logger.Info)
	}

	switch config.ACARSProcessorSettings.Database.Type {
	case "mariadb":
		if err := InitMariaDB(logConfig); err != nil {
			log.Fatal(yo.Uhh("unable to initialize mariadb, err: %s", err).FRFR())
		}
	// SQLite is used as a DB library even if we're not saving messages.
	default:
		if err := InitSQLite(logConfig); err != nil {
			log.Fatal(yo.Uhh("unable to initialize sqlite, err: %s", err).FRFR())
		}

	}

	// ACARS
	am := []ACARSMessage{}
	if err := db.AutoMigrate(ACARSMessage{}); err != nil {
		log.Fatal(yo.Uhh("Unable to automigrate ACARSMessage type: %s", err).FRFR())
	}
	db.Find(&am, ACARSMessage{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}})
	for _, a := range am {
		ACARSMessageQueue <- a.ID
	}
	if config.ACARSProcessorSettings.Database.Enabled {
		log.Info(yo.FYI("Loaded %d ACARS messages from the db", len(am)).FRFR())
	}

	// VDLM2
	vm := []VDLM2Message{}
	if err := db.AutoMigrate(VDLM2Message{}); err != nil {
		log.Fatal(yo.Uhh("Unable to automigrate VDLM2Message type: %s", err).FRFR())
	}
	db.Find(&vm, VDLM2Message{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}})
	for _, v := range vm {
		VDLM2MessageQueue <- v.ID
	}
	if config.ACARSProcessorSettings.Database.Enabled {
		log.Info(yo.FYI("Loaded %d VDLM2 messages from the db", len(vm)).FRFR())
	}

	// Ollama filter
	if err := db.AutoMigrate(OllamaFilterResult{}); err != nil {
		log.Fatal(yo.Uhh("Unable to automigrate Ollama filter type: %s", err).FRFR())
	}

	return nil
}
