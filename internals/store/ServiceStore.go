package store

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
)

type ServiceStore interface {
	Clear() error
	Status(status *models.Status) error
}

type PSQLServiceStore struct {
	db *sql.DB
}

func CreatePSQLServiceStore(db *sql.DB) ServiceStore {
	return PSQLServiceStore{db: db}
}

func (ss PSQLServiceStore) Status(status *models.Status) error {
	row := ss.db.QueryRow(`
		select PostNum, ForumNum, ThreadNum, UserNum
		from Status;
`)
	if err := row.Scan(&status.Post, &status.Forum, &status.Thread, &status.User); err != nil {
		return errors.Wrap(err, "PSQL Service Store status:")
	}

	return nil
}

func (ss PSQLServiceStore) Clear() error {
	_, err := ss.db.Exec(`
		drop table if exists Votes cascade ;
		drop table if exists Posts cascade ;
		drop table if exists Threads cascade ;
		drop table if exists Forums cascade ;
		drop table if exists Users cascade ;
		drop table if exists Status cascade ;
`)

	if err != nil {
		logs.Error(err)
	}

	return nil
}
