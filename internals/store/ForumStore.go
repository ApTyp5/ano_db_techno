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
	prefix := "PSQL forumStore selectBySlug"
	row := fs.db.QueryRow(`
		select PostNum, ThreadNum, Title, Slug, (select NickName from Users where Id = Responsible)
			from Forums
			where Slug = $1;
`, forum.Slug)

	return errors.Wrap(row.Scan(
		&forum.Posts,
		&forum.Threads,
		&forum.Title,
		&forum.Slug,
		&forum.User,
	), prefix)
}

func (fs PSQLForumStore) Insert(forum *models.Forum) error {
	prefix := "PSQL forumStore Insert"
	row := fs.db.QueryRow(`
		Insert into Forums (Slug, Title, Responsible)
		values ($2, $3, (
		    select Id
		    from Users
		    where NickName = $1
		))
		returning Slug, Title, (
		    select NickName from Users where NickName = $1
		);
`,
		forum.User,
		forum.Slug,
		forum.Title,
	)

	return errors.Wrap(row.Scan(&forum.Slug, &forum.Title, &forum.User), prefix)
}

func (fs PSQLForumStore) Count(num *uint) error {
	prefix := "PSQL forumStore Count"
	row := fs.db.QueryRow(`
		select ForumNum from Status;
`)

	if err := row.Scan(num); err != nil {
		return errors.Wrap(err, prefix)
	}

	return nil
}
