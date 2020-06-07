package store

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

type VoteStore interface {
	Insert(vote *models.Vote, thread *models.Thread) error // thread.Vote
	Update(vote *models.Vote, thread *models.Thread) error // thread.Vote
}

type PSQLVoteStore struct {
	db *pgx.ConnPool
}

func CreatePSQLVoteStore(db *pgx.ConnPool) VoteStore {
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
		where Author = $2 
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
		select th.author, th.Created, th.Forum,
	    	th.Message, th.Id, th.Title, th.vote_num, th.Slug
		from Threads th
			`

	var row *pgx.Row
	if thread.Slug == "" {
		selectQuery += "where th.Id = $1;"
		row = tx.QueryRow(selectQuery, thread.Id)
	} else {
		selectQuery += "where th.Slug = $1;"
		row = tx.QueryRow(selectQuery, thread.Slug)
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
				values ($1,`

	if thread.Slug == "" {
		query += "(select slug from threads where id = $2), $3);"
		_, err = tx.Exec(query, vote.NickName, thread.Id, vote.Voice)
	} else {
		query += "$2, $3);"
		_, err = tx.Exec(query, vote.NickName, thread.Slug, vote.Voice)
	}

	if err != nil {
		return errors.Wrap(err, "PSQLVoteStore Insert insert")
	}

	selectQuery := `
		select th.author, th.Created, th.Forum,
	    	th.Message, th.Id, th.Title, th.Vote_num, th.Slug
		from Threads th
			`

	var row *pgx.Row
	if thread.Slug == "" {
		selectQuery += "where th.Id = $1;"
		row = tx.QueryRow(selectQuery, thread.Id)
	} else {
		selectQuery += "where th.Slug = $1;"
		row = tx.QueryRow(selectQuery, thread.Slug)
	}

	if err = errors.Wrap(row.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Message,
		&thread.Id, &thread.Title, &thread.Votes, &thread.Slug), "PSQLVoteStore Insert"); err != nil {
		return err
	}

	return tx.Commit()
}
