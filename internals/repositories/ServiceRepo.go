package repositories

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
)

type ServiceRepo interface {
	Clear() error
	Status(status *models.Status) error
}

type PSQLServiceRepo struct {
	db     *pgx.ConnPool
	status *pgx.PreparedStatement
}

func CreatePSQLServiceRepo(db *pgx.ConnPool) ServiceRepo {
	var err error
	prefix := "service_"
	repo := PSQLServiceRepo{db: db}

	repo.status, err = db.Prepare(prefix+"status", `
		select post_num, forum_num, thread_num, user_num
		from Status;
	`)
	panicIfErr(err)

	return repo
}

func (serviceRepo PSQLServiceRepo) Status(status *models.Status) error {
	return serviceRepo.db.QueryRow(
		serviceRepo.status.Name).Scan(
		&status.Post,
		&status.Forum,
		&status.Thread,
		&status.User)
}

func (serviceRepo PSQLServiceRepo) Clear() error {
	_, err := serviceRepo.db.Exec(`
	TRUNCATE TABLE votes CASCADE ;
	TRUNCATE TABLE posts CASCADE ;
	TRUNCATE TABLE threads CASCADE ;
	TRUNCATE TABLE forums CASCADE ;
	TRUNCATE TABLE users CASCADE ;
	TRUNCATE TABLE status CASCADE ;
	INSERT INTO status DEFAULT VALUES ;`)
	return err
}
