package store

import (
	"database/sql"
	"fmt"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/pkg/errors"
)

type ThreadStore interface {
	Count(amount *uint) error
	Insert(thread *models.Thread) error                                                                     // forum.AddThread
	SelectByForum(threads *[]*models.Thread, forum *models.Forum, limit int, since string, desc bool) error // forum.GetThreads
	////////////////////////
	SelectBySlug(thread *models.Thread) error // Details
	SelectById(thread *models.Thread) error   // Details
	UpdateBySlug(thread *models.Thread) error // Edit
	UpdateById(thread *models.Thread) error   // Edit
}

type PSQLThreadStore struct {
	db *sql.DB
}

func CreatePSQLThreadStore(db *sql.DB) ThreadStore {
	return PSQLThreadStore{db: db}
}

func (P PSQLThreadStore) Count(amount *uint) error {
	row := P.db.QueryRow(`
		select ThreadNum from Status;
`)

	if err := row.Scan(amount); err != nil {
		return errors.Wrap(err, "PSQLThreadStore Count")
	}

	return nil
}

func (P PSQLThreadStore) Insert(thread *models.Thread) error {
	var row *sql.Row

	if len(thread.Created) == 0 {
		row = P.db.QueryRow(`
		insert into Threads (Author, Forum, Message, Slug, Title) values 
			((select Id from Users where NickName = $1), 
			 (select Id from Forums where Slug = $2), 
			 $3, 
			 (coalesce(nullif($4, ''))),
			 $5)
		returning Id, (coalesce(Slug, '')), Title, VoteNum, Created,
		    (select Slug from Forums where Id = Forum);
`, thread.Author, thread.Forum, thread.Message, thread.Slug, thread.Title)

		if err := row.Scan(&thread.Id, &thread.Slug, &thread.Title, &thread.Votes, &thread.Created, &thread.Forum); err != nil {
			return errors.Wrap(err, "PSQLThreadStore Insert")
		}

		return nil
	} else {
		row = P.db.QueryRow(`
		insert into Threads (Author, Forum, Message, Slug, Title, Created) values 
			((select Id from Users where NickName = $1), 
			 (select Id from Forums where Slug = $2), 
			 $3, 
			 $4, 
			 $5,
			 $6)
		returning Id, Slug, Title, VoteNum, (select Slug from Forums where Id = Forum);
`, thread.Author, thread.Forum, thread.Message, thread.Slug, thread.Title, thread.Created)

		if err := row.Scan(&thread.Id, &thread.Slug, &thread.Title, &thread.Votes, &thread.Forum); err != nil {
			return errors.Wrap(err, "PSQLThreadStore Insert")
		}

		return nil
	}
}

func (P PSQLThreadStore) SelectByForum(threads *[]*models.Thread, forum *models.Forum, limit int, since string, desc bool) error {
	var (
		rows *sql.Rows
		err  error
	)

	query1 := `	select t.Id, u.NickName, f.Slug, t.Created, t.Message, t.Slug, t.Title, t.VoteNum
				from Threads t
					join Users u on u.Id = t.Author
					join Forums f on f.Id = t.Forum
				where f.Slug = $1`

	query2 := " order by t.Created"

	if desc {
		query2 += " desc"
	}

	if limit != 0 {
		query2 += fmt.Sprintf(" limit %d", limit)
	}

	if since != "" {
		if desc {
			query1 += " and t.Created <= $2"
		} else {
			query1 += " and t.Created >= $2"
		}
		rows, err = P.db.Query(query1+query2, forum.Slug, since)
	} else {
		rows, err = P.db.Query(query1+query2, forum.Slug)
	}

	if err != nil {
		return errors.Wrap(err, "PSQLThreadStore selectByForum")
	}

	for rows.Next() {
		thread := &models.Thread{}
		if err := rows.Scan(&thread.Id, &thread.Author, &thread.Forum, &thread.Created, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes); err != nil {
			return errors.Wrap(err, "PSQLThreadStore selectByForum scan error")
		}
		*threads = append(*threads, thread)
	}

	return nil
}

func (P PSQLThreadStore) SelectBySlug(thread *models.Thread) error {
	row := P.db.QueryRow(`
		select t.Id, u.NickName, f.Slug, t.Created, t.Message, t.Title, t.VoteNum, t.Slug
		from Threads t
			join Users u on u.Id = t.Author
			join Forums f on f.Id = t.Forum
		where t.Slug = $1;
`, thread.Slug)

	return errors.Wrap(row.Scan(&thread.Id, &thread.Author, &thread.Forum, &thread.Created,
		&thread.Message, &thread.Title, &thread.Votes, &thread.Slug),
		"PSQLThreadStore SelectBySlug")
}

func (P PSQLThreadStore) SelectById(thread *models.Thread) error {
	row := P.db.QueryRow(`
		select t.Slug, u.NickName, f.Slug, t.Created, t.Message, t.Title, t.VoteNum
		from Threads t
			join Users u on u.Id = t.Author
			join Forums f on f.Id = t.Forum
		where t.Id = $1;
`, thread.Id)

	return errors.Wrap(row.Scan(&thread.Slug, &thread.Author, &thread.Forum, &thread.Created,
		&thread.Message, &thread.Title, &thread.Votes),
		"PSQLThreadStore SelectById")
}

func (P PSQLThreadStore) UpdateBySlug(thread *models.Thread) error {
	row := P.db.QueryRow(`
		update Threads t set Message = $1, Title = $2
		where t.Slug = $3
		returning (
		    select u.NickName 
		    from Users u
		        join Threads t on t.Author = u.Id
		    where t.Slug = $3
		), t.Created, (
		         select f.Slug
		         from Forums f
		         	join Threads t on t.Forum = f.Id
		         where t.Slug = $3
		     ), t.Id, t.Message, t.Title, t.VoteNum, t.Slug; 
`, thread.Message, thread.Title, thread.Slug)

	return errors.Wrap(row.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Id,
		&thread.Message, &thread.Title, &thread.Votes, &thread.Slug),
		"PSQLThreadStore UpdateBySlug")
}

func (P PSQLThreadStore) UpdateById(thread *models.Thread) error {
	row := P.db.QueryRow(`
		update Threads t set Message = $1, Title = $2
		where t.Id = $3
		returning (
		    select u.NickName 
		    from Users u
		        join Threads t on t.Author = u.Id
		    where t.Id = $3
		), t.Created, (
		         select f.Slug
		         from Forums f
		         	join Threads t on t.Forum = f.Id
		         where t.Id = $3
		     ), t.Slug, t.Message, t.Title, t.VoteNum, t.slug; 
`, thread.Message, thread.Title, thread.Id)

	return errors.Wrap(row.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Slug,
		&thread.Message, &thread.Title, &thread.Votes, &thread.Slug),
		"PSQLThreadStore UpdateById")
}
