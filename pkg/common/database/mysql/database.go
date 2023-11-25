package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func NewDBInstance(user, pass, host, db string, port int) *sql.DB {
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&multiStatements=True",
		user, pass, host, port, db,
	)

	instance, err := sql.Open("mysql", connString)
	if err != nil {
		panic(err)
	}

	if err := initSchema(instance); err != nil {
		panic(err)
	}

	if err := initTriggers(instance); err != nil {
		panic(err)
	}

	if err := initStored(instance); err != nil {
		panic(err)
	}

	return instance
}
