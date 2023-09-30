package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	driver     = "sqlite3"
	fileName   = "store.db"
	connString = fmt.Sprintf("file:%s?parseTime=true", fileName)
)

func NewSQLiteDB() *sql.DB {
	db, err := sql.Open(driver, connString)
	if err != nil {
		panic(err)
	}

	return db
}
