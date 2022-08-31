package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ihexxa/quickshare/src/db/rdb"
)

type SQLite struct {
	rdb.IDB
	dbPath string
}

func NewSQLite(dbPath string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_synchronous=FULL", dbPath))
	if err != nil {
		return nil, err
	}

	return &SQLite{
		IDB:    db,
		dbPath: dbPath,
	}, nil
}
