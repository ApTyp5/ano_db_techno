package database

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/logs"
)

// DropTables -- отчистка схемы бд
func DropTables(db *sql.DB) {
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
		logs.Error(err)
	}
}
