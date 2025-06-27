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
)

var (
	sqlitePath = "./messages.db"
)

func InitSQLite() (err error) {
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
		p := sqlitePath
		if p != "" {
			sqlitePath = config.ACARSProcessorSettings.Database.SQLiteDatabasePath
		}
		log.Info(yo.Bet("Database is enabled at path %s", p).FRFR())
	}
	db, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	return err
}

func InitMariaDB() (err error) {
	dsn := config.ACARSProcessorSettings.Database.ConnectionString
	if dsn == "" {
		return errors.New("mariadb specified but connection string is not set")
	}
	u, err := url.Parse(dsn)
	if err != nil {
		log.Panic(yo.Uhh("unable to parse mariadb connection string: %s", err))
	}
	q := u.Query()
	q.Set("parseTime", "True")
	u.RawQuery = q.Encode()
	db, err = gorm.Open(mysql.Open(u.String()), &gorm.Config{})
	return err
}

func LoadSavedMessages() error {
	switch config.ACARSProcessorSettings.Database.Type {
	case "mariadb":
		if err := InitMariaDB(); err != nil {
			log.Fatal(yo.Uhh("unable to initialize mariadb, err: %s", err).FRFR())
		}
	// SQLite is used as a DB library even if we're not saving messages.
	default:
		if err := InitSQLite(); err != nil {
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
