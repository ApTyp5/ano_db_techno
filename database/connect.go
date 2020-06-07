package database

import (
	"database/sql"
)

func Connect(connStr string, connNum int) *sql.DB {
	conn, err := sql.Open("pgx", connStr)

	if err != nil {
		panic(err)
	}

	return conn
}
