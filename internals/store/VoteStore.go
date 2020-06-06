package store

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/pkg/errors"
)

type VoteStore interface {
	Insert(vote *models.Vote, thread *models.Thread) error // thread.Vote
	Update(vote *models.Vote, thread *models.Thread) error // thread.Vote
}

type PSQLVoteStore struct {
	db *sql.DB
}

func CreatePSQLVoteStore(db *sql.DB) VoteStore {
	return PSQLVoteStore{db: db}
}

func (P PSQLVoteStore) Update(vote *models.Vote, thread *models.Thread) error {
	tx, err := P.db.Begin()
	if err != nil {
		return errors.Wrap(err, "PSQLVoteStore Update begin")
	}

	defer tx.Rollback()

	query := `
		update Votes set Voice = $1
		where Author = (select Id from Users where NickName = $2) 
			and 
`

	if thread.Slug == "" {
		query += "Thread = $3;"
		_, err = P.db.Exec(query, vote.Voice, vote.NickName, thread.Id)
	} else {
		query += "Thread = (select Id from Threads where Slug = $3);"
		_, err = P.db.Exec(query, vote.Voice, vote.NickName, thread.Slug)
	}

	if err != nil {
		return errors.Wrap(err, "PSQLVoteStore Update insert")
	}

	selectQuery := `
		select u.NickName, th.Created, th.Forum,
	    	th.Message, th.Id, th.Title, th.VoteNum, th.Slug
		from Threads th
			join Users u on u.Id = th.Author
			`

	var row *sql.Row
	if thread.Slug == "" {
		selectQuery += "where th.Id = $1;"
		row = P.db.QueryRow(selectQuery, thread.Id)
	} else {
		selectQuery += "where th.Slug = $1;"
		row = P.db.QueryRow(selectQuery, thread.Slug)
	}

	if err = errors.Wrap(row.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Message,
		&thread.Id, &thread.Title, &thread.Votes, &thread.Slug), "PSQLVoteStore Update"); err != nil {
		return err
	}

	return tx.Commit()
}

func (P PSQLVoteStore) Insert(vote *models.Vote, thread *models.Thread) error {
	tx, err := P.db.Begin()
	if err != nil {
		return errors.Wrap(err, "PSQLVoteStore Insert begin")
	}

	defer tx.Rollback()

	query := `insert into Votes (Author, Thread, Voice)
				values ((select Id from Users where NickName = $1),`
	if thread.Slug == "" {
		query += "$2, $3);"
		_, err = P.db.Exec(query, vote.NickName, thread.Id, vote.Voice)
	} else {
		query += "(select Id from Threads where Slug = $2), $3);"
		_, err = P.db.Exec(query, vote.NickName, thread.Slug, vote.Voice)
	}

	if err != nil {
		return errors.Wrap(err, "PSQLVoteStore Insert insert")
	}

	selectQuery := `
		select u.NickName, th.Created, f.Slug,
	    	th.Message, th.Id, th.Title, th.VoteNum, th.Slug
		from Threads th
			join Users u on u.Id = th.Author
			join Forums f on f.Id = th.Forum
			`

	var row *sql.Row
	if thread.Slug == "" {
		selectQuery += "where th.Id = $1;"
		row = P.db.QueryRow(selectQuery, thread.Id)
	} else {
		selectQuery += "where th.Slug = $1;"
		row = P.db.QueryRow(selectQuery, thread.Slug)
	}

	if err = errors.Wrap(row.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Message,
		&thread.Id, &thread.Title, &thread.Votes, &thread.Slug), "PSQLVoteStore Insert"); err != nil {
		return err
	}

	return tx.Commit()
}
