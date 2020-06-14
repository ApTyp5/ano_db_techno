package repositories

import (
	"context"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"sort"
)

type PostRepo interface {
	Count(amount *uint) error
	SelectById(post *models.Post) error
	UpdateById(post *models.Post) error                                                          // Edit
	InsertPostsByThread(thread *models.Thread, posts []models.Post, nicks map[string]bool) error // thread.AddPosts
	// threads.Posts
	SelectByThread(posts *[]models.Post, thread *models.Thread, limit int, since int, desc bool, mode string) error
}

type PSQLPostRepo struct {
	db             *pgx.ConnPool
	count          *pgx.PreparedStatement
	selectById     *pgx.PreparedStatement
	updateById     *pgx.PreparedStatement
	insertByThread *pgx.PreparedStatement
	addForumUsers  *pgx.PreparedStatement
}

func CreatePSQLPostRepo(db *pgx.ConnPool) PostRepo {
	var err error
	prefix := "post_"
	repo := PSQLPostRepo{db: db}

	repo.count, err = db.Prepare(prefix+"count", "select post_num from Status;")
	panicIfErr(err)

	repo.selectById, err = db.Prepare(prefix+"selectById", `
		select p.author, p.Created, t.Forum, p.is_edited, p.Message, coalesce(p.Parent, 0), p.Thread
			from Posts p
				join Threads t on p.Thread = t.Id
			where p.id = $1;`)
	panicIfErr(err)

	repo.updateById, err = db.Prepare(prefix+"updateById", `
		update Posts p
			set Message = coalesce(nullif($1, ''), message)
			where p.id = $2
		returning 
			p.author, 
		    Created, 
		    (select t.Forum from Posts p join Threads t on t.Id = p.Thread where p.Id = $2), 
		    is_edited, Message, coalesce(p.parent, 0), Thread;
`)
	panicIfErr(err)

	repo.insertByThread, err = db.Prepare("InsertPostsByThread", `
	INSERT INTO posts (author, thread, message, parent, forum) values 
		($1, $2, $3, nullif($4, 0), $5)
		RETURNING id, thread, created, is_edited, message, coalesce(parent, 0), forum
	`)
	panicIfErr(err)

	repo.addForumUsers, err = db.Prepare(prefix+"addForumUsers", `
		insert into forum_users (forum, user_nick) values
			($1, $2) on conflict do nothing`)
	panicIfErr(err)

	return repo
}

func (postRepo PSQLPostRepo) Count(amount *uint) error {
	prefix := "PSQL PostRepo Count"
	row := postRepo.db.QueryRow(`
		select post_num from Status;
`)
	if err := row.Scan(amount); err != nil {
		return errors.Wrap(err, prefix)
	}

	return nil
}

func (postRepo PSQLPostRepo) SelectById(post *models.Post) error {
	return postRepo.db.QueryRow(
		postRepo.selectById.Name,
		post.Id).Scan(
		&post.Author,
		&post.Created,
		&post.Forum,
		&post.IsEdited,
		&post.Message,
		&post.Parent,
		&post.Thread)
}

func (postRepo PSQLPostRepo) UpdateById(post *models.Post) error {
	return postRepo.db.QueryRow(
		postRepo.updateById.Name,
		post.Message,
		post.Id).Scan(
		&post.Author,
		&post.Created,
		&post.Forum,
		&post.IsEdited,
		&post.Message,
		&post.Parent,
		&post.Thread)
}

func (postRepo PSQLPostRepo) InsertPostsByThread(thread *models.Thread, posts []models.Post, nicks map[string]bool) error {
	if len(posts) == 0 {
		return nil
	}

	tx, err := postRepo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bt := tx.BeginBatch()
	defer bt.Close()

	for i := range posts {
		bt.Queue(postRepo.insertByThread.Name,
			[]interface{}{
				posts[i].Author,
				thread.Id,
				posts[i].Message,
				posts[i].Parent,
				thread.Forum},
			nil, nil)
	}

	sortNicks := make([]string, 0, len(nicks))
	for nick := range nicks {
		sortNicks = append(sortNicks, nick)
	}
	sort.Strings(sortNicks)

	for _, nick := range sortNicks {
		bt.Queue(postRepo.addForumUsers.Name,
			[]interface{}{
				thread.Forum,
				nick,
			}, nil, nil)
	}

	if err := bt.Send(context.Background(), nil); err != nil {
		return err
	}

	for i := range posts {
		if err := bt.QueryRowResults().Scan(
			&posts[i].Id,
			&posts[i].Thread,
			&posts[i].Created,
			&posts[i].IsEdited,
			&posts[i].Message,
			&posts[i].Parent,
			&posts[i].Forum); err != nil {
			return err
		}
	}

	for _ = range nicks {
		_, err := bt.ExecResults()
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("update forums set post_num = post_num + $1 where slug = $2", len(posts), thread.Forum)
	if err != nil {
		return err
	}

	_, err = tx.Exec("update status set post_num = post_num + $1", len(posts))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (postRepo PSQLPostRepo) SelectByThread(posts *[]models.Post, thread *models.Thread, limit int, since int, desc bool, mode string) error {
	rows, err := postRepo.db.Query(
		"SELECT author, created, id,"+
			"is_edited, message, parent, "+
			"thread, forum "+
			"from select_posts_by_thread($1, $2, $3, $4, $5);",
		thread.Id,
		limit,
		since,
		desc,
		mode)
	if err != nil {
		return err
	}

	for rows.Next() {
		i := len(*posts)
		*posts = append(*posts, models.Post{})
		if err := rows.Scan(&(*posts)[i].Author, &(*posts)[i].Created, &(*posts)[i].Id,
			&(*posts)[i].IsEdited, &(*posts)[i].Message, &(*posts)[i].Parent,
			&(*posts)[i].Thread, &(*posts)[i].Forum); err != nil {
			return err
		}
	}

	return nil
}

func (postRepo PSQLPostRepo) SelectByThreadTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	var (
		hasSince  = since >= 0
		rows      *pgx.Rows
		err       error
		query     = ""
		sincePath = ""
	)

	tx, err := postRepo.db.Begin()
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

func (postRepo PSQLPostRepo) SelectByThreadParentTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	var (
		hasLimit = limit > 0
		hasSince = since >= 0
		rows     *pgx.Rows
		err      error

		query = ""
	)

	tx, err := postRepo.db.Begin()
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
