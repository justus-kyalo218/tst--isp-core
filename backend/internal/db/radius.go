package db

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var radiusDB *sql.DB

// InitRadius initializes the FreeRADIUS MySQL connection when RADIUS_DSN is provided.
// It is optional so local dev can run without a RADIUS database.
func InitRadius() (*sql.DB, error) {
	dsn := strings.TrimSpace(os.Getenv("RADIUS_DSN"))
	if dsn == "" {
		return nil, nil
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	radiusDB = db
	return radiusDB, nil
}

func RadiusDB() (*sql.DB, error) {
	if radiusDB == nil {
		return nil, errors.New("radius db not initialized")
	}
	return radiusDB, nil
}
