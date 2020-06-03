package database

import (
	"database/sql"
)

func Connect(connStr string, connNum int) *sql.DB {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	conn.SetMaxOpenConns(connNum)

	if err = conn.Ping(); err != nil {
		panic(err)
	}

	return conn
}
