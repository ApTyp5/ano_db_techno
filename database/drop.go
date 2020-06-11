package database

import (
	"github.com/jackc/pgx"
)

// DropTables -- отчистка схемы бд
func DropTables(db *pgx.ConnPool) {
	_, err := db.Exec(`
drop sequence if exists posts_id_seq cascade ;
drop function if exists PostId;
drop function if exists PostPar;

drop table if exists Votes;
drop table if exists Posts;
drop table if exists Threads;
drop table if exists Forums;
drop table if exists Users;
drop table if exists Status;
`)

	if err != nil {
		panic(err)
	}
}

func TruncTables(db *pgx.ConnPool) {
	_, err := db.Exec(`
truncate table if exists Votes;
truncate table if exists Posts;
truncate table if exists Threads;
truncate table if exists Forums;
truncate table if exists Users;
truncate table if exists Status;
`)
	if err != nil {
		panic(err)
	}
}
