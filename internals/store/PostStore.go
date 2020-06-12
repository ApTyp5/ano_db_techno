package store

import (
	"context"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

type PostStore interface {
	Count(amount *uint) error
	SelectById(post *models.Post) error
	UpdateById(post *models.Post) error                                    // Edit
	InsertPostsByThread(thread *models.Thread, posts *[]models.Post) error // thread.AddPosts
	// threads.Posts
	SelectByThreadFlat(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
	// threads.Posts
	SelectByThreadTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
	// threads.Posts
	SelectByThreadParentTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
}

type PSQLPostStore struct {
	db *pgx.ConnPool
}

func CreatePSQLPostStore(db *pgx.ConnPool) PostStore {
	return PSQLPostStore{db: db}
}

func (P PSQLPostStore) Count(amount *uint) error {
	prefix := "PSQL PostStore Count"
	row := P.db.QueryRow(`
		select post_num from Status;
`)
	if err := row.Scan(amount); err != nil {
		return errors.Wrap(err, prefix)
	}

	return nil
}

func (P PSQLPostStore) SelectById(post *models.Post) error {
	prefix := "PSQL PostStore SelectById"
	row := P.db.QueryRow(`
		select p.author, p.Created, t.Forum, p.is_edited, p.Message, coalesce(p.Parent, 0), p.Thread
			from Posts p
				join Threads t on p.Thread = t.Id
			where p.id = $1;
`,
		post.Id)

	if err := row.Scan(&post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread); err != nil {
		return errors.Wrap(err, prefix)
	}

	return nil
}

func (P PSQLPostStore) UpdateById(post *models.Post) error {
	prefix := "PSQL PostStore UpdateById"
	row := P.db.QueryRow(`
		update Posts p
			set Message = $1
			where p.id = $2
		returning 
			p.author, 
		    Created, 
		    (select t.Forum from Posts p join Threads t on t.Id = p.Thread where p.Id = $2), 
		    is_edited, Message, coalesce(p.parent, 0), Thread;
`, post.Message, post.Id)

	if err := row.Scan(&post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread); err != nil {
		return errors.Wrap(err, prefix)
	}

	return nil
}

func (P PSQLPostStore) InsertPostsByThread(thread *models.Thread, posts *[]models.Post) error {
	tx, err := P.db.Begin()
	if err != nil {
		return errors.Wrap(err, "PSQLPostStore insertPostsByThread's id error")
	}
	defer tx.Rollback()

	if thread.Id < 0 {
		if err := tx.QueryRow("SELECT id, forum FROM threads WHERE slug = $1",
			thread.Slug).Scan(&thread.Id, &thread.Forum); err != nil {
			return errors.Wrap(err, "PSQLPostStore insertPostsByThread select thread id")
		}
	} else {
		if err := tx.QueryRow("SELECT forum FROM threads WHERE id = $1",
			thread.Id).Scan(&thread.Forum); err != nil {
			return errors.Wrap(err, "PSQLPostStore insertPostsByThread select thread id")
		}
	}

	if len(*posts) == 0 {
		return nil
	}

	stat, err := tx.Prepare("insert_post", "INSERT INTO posts (author, thread, message, parent) values "+
		"($1, $2, $3, nullif($4, 0))"+
		"RETURNING id, thread, created, is_edited, message, coalesce(parent, 0)")
	bt := tx.BeginBatch()

	if err != nil {
		return errors.Wrap(err, "PSQLPostStore insertPostsByThread prepare")
	}

	for i := range *posts {
		bt.Queue(stat.Name, []interface{}{(*posts)[i].Author, thread.Id, (*posts)[i].Message, (*posts)[i].Parent}, nil, nil)
	}

	if err := bt.Send(context.Background(), nil); err != nil {
		return err
	}

	for i := range *posts {
		if err := bt.QueryRowResults().Scan(&(*posts)[i].Id, &(*posts)[i].Thread, &(*posts)[i].Created,
			&(*posts)[i].IsEdited, &(*posts)[i].Message, &(*posts)[i].Parent); err != nil {
			return err
		}
		(*posts)[i].Forum = thread.Forum
	}

	_, _ = tx.Exec("update forums set post_num = post_num + $1 where slug = $2", len(*posts), thread.Forum)
	_, _ = tx.Exec("update status set post_num = post_num + $1", len(*posts))

	return tx.Commit()
}

func (P PSQLPostStore) SelectByThreadFlat(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	hasSince := since >= 0

	query := `
		Select p.author, p.Created, t.Forum, p.Id, p.is_edited, p.Message, coalesce(p.Parent, 0), p.Thread
		From Posts p
			join Threads t on t.Id = p.Thread
`
	query += " where t.Id = $1 "

	if desc {
		if hasSince {
			query += " and p.Id < $3 "
		}
		query += " Order By p.Created Desc, p.Id Desc"
	} else {
		if hasSince {
			query += " and p.Id > $3"
		}
		query += " Order By p.Created, p.Id"
	}

	query += `
		Limit $2;
			`
	var (
		rows *pgx.Rows
		err  error
	)

	if hasSince {
		rows, err = P.db.Query(query, thread.Id, limit, since)
	} else {
		rows, err = P.db.Query(query, thread.Id, limit)
	}

	if err != nil {
		return errors.Wrap(err, "PostRepo select by thread id flat error: ")
	}
	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}
		if err := rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id,
			&post.IsEdited, &post.Message, &post.Parent, &post.Thread); err != nil {
			return errors.Wrap(err, "select by thread id flat scan error")
		}

		*posts = append(*posts, post)
	}
	return nil
}

func (P PSQLPostStore) SelectByThreadTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	var (
		hasSince  = since >= 0
		rows      *pgx.Rows
		err       error
		query     = ""
		sincePath = ""
	)

	tx, err := P.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if thread.Id < 0 {
		if err := tx.QueryRow("select id, forum from threads where slug = $1", thread.Slug).Scan(&thread.Id, &thread.Forum); err != nil {
			return err
		}
	} else {
		if err := tx.QueryRow("select forum from threads where id = $1", thread.Id).Scan(&thread.Forum); err != nil {
			return err
		}
	}

	query += `
			Select author, Created, Id, is_edited, Message, coalesce(Parent, 0), Thread
				From Posts
			`

	if hasSince {
		sincePath += " (select path from posts where Id = $3) "
		if desc {
			query += " where path < " + sincePath
		} else {
			query += " where path > " + sincePath
		}
		query += " and thread = $1 "
	} else {
		query += " where thread = $1 "
	}

	if desc {
		query += " order by path desc "
	} else {
		query += " order by path "
	}

	if limit != 0 {
		query += " LIMIT $2; "
	}

	if hasSince {
		rows, err = tx.Query(query, thread.Id, limit, since)
	} else {
		rows, err = tx.Query(query, thread.Id, limit)
	}

	if err != nil {
		return errors.Wrap(err, "select by thread id tree error")
	}
	defer rows.Close()

	for rows.Next() {
		post := &models.Post{Forum: thread.Forum}

		if err := rows.Scan(&post.Author, &post.Created, &post.Id, &post.IsEdited,
			&post.Message, &post.Parent, &post.Thread); err != nil {
			return errors.Wrap(err, "select by thread id tree scan error")
		}

		*posts = append(*posts, post)
	}
	return tx.Commit()
}

func (P PSQLPostStore) SelectByThreadParentTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	var (
		hasLimit = limit > 0
		hasSince = since >= 0
		rows     *pgx.Rows
		err      error

		query = ""
	)

	tx, err := P.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if thread.Id < 0 {
		if err := tx.QueryRow("select id, forum from threads where slug = $1", thread.Slug).Scan(&thread.Id, &thread.Forum); err != nil {
			return err
		}
	} else if err := tx.QueryRow("select forum from threads where id = $1", thread.Id).Scan(&thread.Forum); err != nil {
		return err
	}

	query += "with init as ( select p.Id from posts p " +
		" where p.thread = $1 " +
		" and p.Parent is null "

	if hasSince {
		if desc {
			query += " and id < (select path[1] from posts where id = $3)"
		} else {
			query += " and id > (select path[1] from posts where id = $3)"
		}
	}

	query += " order by p.Id "
	if desc {
		query += " desc "
	}

	if hasLimit {
		query += " limit $2 "
	}
	query += ")"

	query += `
			Select p.Author, p.Created, p.Id, p.is_edited, p.Message, coalesce(p.Parent, 0), p.Thread
				From init join posts p on init.id = p.path[1]
			`

	if desc {
		query += " order by p.path[1] desc, p.path[2:]"
	} else {
		query += " order by p.path "
	}

	if hasSince {
		rows, err = tx.Query(query, thread.Id, limit, since)
	} else {
		rows, err = tx.Query(query, thread.Id, limit)
	}

	if err != nil {
		return errors.Wrap(err, "select by thread id parent tree error")
	}
	defer rows.Close()

	for rows.Next() {
		post := &models.Post{Forum: thread.Forum}

		if err := rows.Scan(&post.Author, &post.Created, &post.Id, &post.IsEdited,
			&post.Message, &post.Parent, &post.Thread); err != nil {
			return errors.Wrap(err, "select by thread id tree scan error")
		}

		*posts = append(*posts, post)
	}
	return nil
}
