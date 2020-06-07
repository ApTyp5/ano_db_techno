package database

import (
	"github.com/jackc/pgx"
)

func Connect(connStr string, connNum int) *pgx.ConnPool {
	connConfig, err := pgx.ParseConnectionString(connStr)
	if err != nil {
		panic(err)
	}

	conn, err := pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     connConfig,
			MaxConnections: connNum,
		},
	)
	if err != nil {
		panic(err)
	}

	return conn
}
