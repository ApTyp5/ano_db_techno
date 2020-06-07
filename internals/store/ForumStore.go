package store

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/pkg/errors"
)

type ForumStore interface {
	SelectBySlug(forum *models.Forum) error
	Insert(forum *models.Forum) error
	Count(num *uint) error
}

type PSQLForumStore struct {
	db *sql.DB
}

func CreatePSQLForumStore(db *sql.DB) ForumStore {
	return PSQLForumStore{
		db: db,
	}
}

func (fs PSQLForumStore) SelectBySlug(forum *models.Forum) error {
	return errors.Wrap(
		fs.db.QueryRow(
			`
		SELECT post_num, thread_num, title, slug, responsible
			FROM forums
		WHERE slug = $1;
		`,
			forum.Slug).Scan(
			&forum.Posts,
			&forum.Threads,
			&forum.Title,
			&forum.Slug,
			&forum.User),
		"PSQL forumStore selectBySlug")
}

func (fs PSQLForumStore) Insert(forum *models.Forum) error {
	return errors.Wrap(fs.db.QueryRow(
		`
			INSERT INTO FORUMS (slug, title, responsible)
			VALUES ($1, $2, (select nick_name from Users where nick_name = $3))
			RETURNING slug, title, responsible, post_num, thread_num
			`,
		forum.Slug,
		forum.Title,
		forum.User).Scan(&forum.Slug,
		&forum.Title,
		&forum.User,
		&forum.Posts,
		&forum.Threads),
		"PSQL forumStore Insert")
}

func (fs PSQLForumStore) Count(num *uint) error {
	return errors.Wrap(fs.db.QueryRow(
		`
			SELECT forum_num FROM status;
			`).Scan(num),
		"PSQL forumStore Count")
}
