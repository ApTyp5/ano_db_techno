package database

import (
	"github.com/jackc/pgx"
)

func Connect(connStr string, connNum int) *pgx.ConnPool {
	config, err := pgx.ParseConnectionString(connStr)
	if err != nil {
		panic(err)
	}

	config.PreferSimpleProtocol = false

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     config,
		MaxConnections: 100,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	}

	pool, err := pgx.NewConnPool(poolConfig)

	if err != nil {
		panic(err)
	}

	return pool
}
