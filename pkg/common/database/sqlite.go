package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	driver = "sqlite3"
)

func NewSQLiteDB(path string) *sql.DB {
	connString := fmt.Sprintf("file:%s?parseTime=true", path)

	db, err := sql.Open(driver, connString)
	if err != nil {
		panic(err)
	}

	return db
}
