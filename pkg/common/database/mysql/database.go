package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DBConnection struct {
	Conn *sql.DB
}

func NewDBInstance(user, pass, host, db string, port int) (*DBConnection, error) {
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&multiStatements=True",
		user, pass, host, port, db,
	)

	instance, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	return &DBConnection{Conn: instance}, nil
}

func (c *DBConnection) InitTables() {
	if err := initSchema(c.Conn); err != nil {
		panic(err)
	}

	if err := initTriggers(c.Conn); err != nil {
		panic(err)
	}

	if err := initStored(c.Conn); err != nil {
		panic(err)
	}
}
