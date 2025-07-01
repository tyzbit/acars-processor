package database

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/decorate"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB         = &gorm.DB{}
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
	if !Config.ACARSProcessorSettings.Database.Enabled {
		log.Info(Content("Database is not enabled"))
		sqlitePath = "file::memory:?cache=shared"
	} else {
		p := sqlitePath
		if p != "" {
			sqlitePath = Config.ACARSProcessorSettings.Database.SQLiteDatabasePath
		}
		log.Info(Success("Database path set to %s", p))
	}
	DB, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	return err
}

func InitMariaDB() (err error) {
	dsn := Config.ACARSProcessorSettings.Database.ConnectionString
	if dsn == "" {
		return errors.New("mariadb specified but connection string is not set")
	}
	u, err := url.Parse(dsn)
	if err != nil {
		log.Fatal(Attention("unable to parse mariadb connection string: %s", err))
	}
	q := u.Query()
	q.Set("parseTime", "True")
	u.RawQuery = q.Encode()
	DB, err = gorm.Open(mysql.Open(u.String()), &gorm.Config{})
	return err
}

func LoadSavedMessages() error {
	switch Config.ACARSProcessorSettings.Database.Type {
	case "mariadb":
		if err := InitMariaDB(); err != nil {
			log.Fatal(Attention("unable to initialize mariadb, err: %s", err))
		}
		// SQLite is used as a DB library even if we're not saving messages.
	default:
		if err := InitSQLite(); err != nil {
			log.Fatal(Attention("unable to initialize sqlite, err: %s", err))
		}
	}
	if Config.ACARSProcessorSettings.Database.Enabled {
		log.Info(Content("%s database initialized", Config.ACARSProcessorSettings.Database.Type))
	}

	return nil
}
