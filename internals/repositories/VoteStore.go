package repositories

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

type VoteStore interface {
	Insert(vote *models.Vote, thread *models.Thread) error         // thread.Vote
	Update(vote *models.Vote, thread *models.Thread) error         // thread.Vote
	InsertOrUpdate(vote *models.Vote, thread *models.Thread) error // thread.Vote
}

type PSQLVoteStore struct {
	db *pgx.ConnPool
}

func CreatePSQLVoteStore(db *pgx.ConnPool) VoteStore {
	return PSQLVoteStore{db: db}
}

func (P PSQLVoteStore) InsertOrUpdate(vote *models.Vote, thread *models.Thread) error {
	tx, err := P.db.Begin()
	if err != nil {
		return errors.Wrap(err, "PSQLVoteStore Update begin")
	}
	defer tx.Rollback()

	if err := tx.QueryRow("select nick_name from users where nick_name = $1;", vote.NickName).Scan(&vote.NickName); err != nil {
		return errors.New("user not exists")
	}

	if thread.Id >= 0 {
		if err := tx.QueryRow("select id from threads where id = $1", thread.Id).Scan(&thread.Id); err != nil {
			return errors.New("thread does not exist")
		}
	} else {
		if err := tx.QueryRow("select id from threads where slug = $1", thread.Slug).Scan(&thread.Id); err != nil {
			return errors.New("thread does not exist")
		}
	}

	if err := tx.QueryRow("select author, thread from votes where author = $1 and thread = $2",
		vote.NickName, thread.Id).Scan(&vote.NickName, &thread.Id); err != nil {
		if _, err := tx.Exec("insert into votes (author, thread, voice) values ($1, $2, $3)",
			vote.NickName, thread.Id, vote.Voice); err != nil {
			return errors.Wrap(err, "insert")
		}
	} else {
		if _, err := tx.Exec("update votes set voice = $1 where author = $2 and thread = $3",
			vote.Voice, vote.NickName, thread.Id); err != nil {
			return errors.Wrap(err, "update")
		}
	}

	if err := tx.QueryRow("select author, created, forum, message, id, title, vote_num, slug "+
		"from threads where id = $1", thread.Id).Scan(&thread.Author, &thread.Created, &thread.Forum,
		&thread.Message, &thread.Id, &thread.Title, &thread.Votes, &thread.Slug); err != nil {
		return errors.Wrap(err, "select thread")
	}

	return tx.Commit()
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

	if thread.Id >= 0 {
		query += "Thread = $3;"
		_, err = tx.Exec(query, vote.Voice, vote.NickName, thread.Id)
	} else {
		query += "Thread = (select Id from Threads where Slug = $3);"
		_, err = tx.Exec(query, vote.Voice, vote.NickName, thread.Slug)
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
	if thread.Id >= 0 {
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

	if thread.Id >= 0 {
		query += "$2, $3);"
		_, err = tx.Exec(query, vote.NickName, thread.Id, vote.Voice)
	} else {
		query += "(SELECT id FROM threads WHERE slug = $2), $3);"
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
	if thread.Id >= 0 {
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
