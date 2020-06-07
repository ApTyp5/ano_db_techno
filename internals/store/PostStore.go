package store

import (
	"database/sql"
	"fmt"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type PostStore interface {
	Count(amount *uint) error
	SelectById(post *models.Post) error
	UpdateById(post *models.Post) error                                     // Edit
	InsertPostsByThread(thread *models.Thread, posts *[]*models.Post) error // thread.AddPosts
	// threads.Posts
	SelectByThreadFlat(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
	// threads.Posts
	SelectByThreadTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
	// threads.Posts
	SelectByThreadParentTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
}

type PSQLPostStore struct {
	db *sql.DB
}

func CreatePSQLPostStore(db *sql.DB) PostStore {
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

func (P PSQLPostStore) InsertPostsByThread(thread *models.Thread, posts *[]*models.Post) error {

	tx, err := P.db.Begin()
	if err != nil {
		return errors.Wrap(err, "PSQLPostStore insertPostsByThread's id error")
	}
	defer tx.Rollback()

	if thread.Id == 0 {
		if err := P.db.QueryRow("SELECT t.id, f.slug FROM threads t JOIN forums f ON t.forum = f.slug WHERE t.slug = $1", thread.Slug).Scan(&thread.Id, &thread.Forum); err != nil {
			return errors.Wrap(err, "PSQLPostStore insertPostsByThread select thread id")
		}
	} else {
		if err := P.db.QueryRow("SELECT f.slug FROM threads t JOIN forums f ON t.forum = f.slug WHERE t.id = $1", thread.Id).Scan(&thread.Forum); err != nil {
			return errors.Wrap(err, "PSQLPostStore insertPostsByThread select thread id")
		}
	}

	if len(*posts) == 0 {
		return nil
	}

	valueArgs := make([]string, 0, len(*posts))
	for i := range *posts {
		nick := (*posts)[i].Author
		thid := thread.Id
		mess := (*posts)[i].Message
		parn := (*posts)[i].Parent

		valueArgs = append(valueArgs, fmt.Sprintf("('%s', %d, '%s'", nick, thid, mess))

		if parn == 0 {
			valueArgs = append(valueArgs, "null)")
		} else {
			valueArgs = append(valueArgs, fmt.Sprintf("%d)", parn))
		}
	}

	query := "INSERT INTO posts (author, thread, message, parent) values " + strings.Join(valueArgs, ",") +
		"RETURNING id, thread, created, is_edited, message, coalesce(parent, 0)"

	rows, err := tx.Query(query)

	if err != nil {
		return errors.Wrap(err, "PSQLPostStore insertPostsByThread main insert")
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		if err := rows.Scan(&(*posts)[i].Id, &(*posts)[i].Thread, &(*posts)[i].Created,
			&(*posts)[i].IsEdited, &(*posts)[i].Message, &(*posts)[i].Parent); err != nil {
			return errors.Wrap(err, "PSQLPostStore insertPostsByThread's id SCAN error")
		}

		(*posts)[i].Forum = thread.Forum
		i++
	}

	return tx.Commit()
}

func (P PSQLPostStore) SelectByThreadFlat(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	hasSince := since >= 0

	query := `
		Select p.author, p.Created, t.Forum, p.Id, p.is_edited, p.Message, coalesce(p.Parent, 0), p.Thread
		From Posts p
			join Threads t on t.Id = p.Thread
`
	if thread.Slug == "" {
		query += " where t.Id = $1 "
	} else {
		query += " where t.Slug = $1 "
	}

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
		rows *sql.Rows
		err  error
	)
	if thread.Slug == "" {
		if hasSince {
			rows, err = P.db.Query(query, thread.Id, limit, since)
		} else {
			rows, err = P.db.Query(query, thread.Id, limit)
		}

	} else {
		if hasSince {
			rows, err = P.db.Query(query, thread.Slug, limit, since)
		} else {
			rows, err = P.db.Query(query, thread.Slug, limit)
		}
	}

	logs.Info("QUERY:\n", query)

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
		hasSince = since >= 0
		hasSlug  = thread.Slug != ""
		rows     *sql.Rows
		err      error
		withPart = ""
		query    = ""
		initPath = ""
	)

	withPart += `
			with recursive ph as (
				select Array [p.Id] as path, p.Id, 0 
				from posts p
				`

	if hasSlug {
		withPart += " join threads t on p.thread = t.Id " +
			" where p.Parent is null and t.Slug = $1 "
	} else {
		withPart +=
			" where p.Parent is null and p.Id = $1 "
	}

	withPart += `
			union all
				select ph.path || array [p.Id] as path, p.Id, coalesce(p.Parent, 0)
					from posts p
			join ph on p.parent = ph.Id
			)
				`

	query += `
			Select p.author, p.Created, t.Forum, p.Id, p.is_edited, p.Message, coalesce(p.Parent, 0), p.Thread
				From Posts p
				join ph on p.Id = ph.Id
				join Threads t on t.Id = p.Thread
			`

	if hasSince {
		initPath += " (select path from ph where Id = $3) "
		if desc {
			query += " where ph.Path < " + initPath
		} else {
			query += " where ph.Path > " + initPath
		}
	}

	if desc {
		query += " order by ph.Path desc "
	} else {
		query += " order by ph.Path "
	}

	if limit != 0 {
		query += " LIMIT $2; "
	}

	logs.Info("QUERY:\n", withPart+query)

	if hasSlug {
		if hasSince {
			rows, err = P.db.Query(withPart+query, thread.Slug, limit, since)
		} else {
			rows, err = P.db.Query(withPart+query, thread.Slug, limit)
		}
	} else {
		if hasSince {
			rows, err = P.db.Query(query, thread.Id, limit, since)
		} else {
			rows, err = P.db.Query(query, thread.Id, limit)
		}
	}

	if err != nil {
		return errors.Wrap(err, "select by thread id tree error")
	}
	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}

		if err := rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited,
			&post.Message, &post.Parent, &post.Thread); err != nil {
			return errors.Wrap(err, "select by thread id tree scan error")
		}

		*posts = append(*posts, post)
	}
	return nil
}

func (P PSQLPostStore) SelectByThreadParentTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	var (
		hasLimit = limit > 0
		hasSince = since >= 0
		hasSlug  = thread.Slug != ""
		rows     *sql.Rows
		err      error

		query = ""
	)

	if hasSince {
		query += ` 
			with recursive since as (
				select id, coalesce(parent, 0) as parent
				from posts where id = ` + strconv.FormatInt(int64(since), 10) + `
					union all
				select p.id, coalesce(p.parent, 0) as parent
				from since s 
					join posts p on p.id = s.parent
			),
					`
	} else {
		query = " with recursive "
	}

	query += `
		init as (
			select p.Id
			from posts p 
			`

	if hasSlug {
		query += " join threads t on p.thread = t.id where t.Slug = $1 "
	} else {
		query += " where p.thread = $1 "
	}
	query += " and p.Parent is null "

	if hasSince {
		if desc {
			query += " and p.Id < (select id from since where parent = 0)"
		} else {
			query += " and p.Id > (select id from since where parent = 0)"
		}
	}

	query += " order by p.Id "
	if desc {
		query += " desc "
	}

	if hasLimit {
		query += " limit $2 "
	}
	query += "), "

	query += `
		ph as (
			select array [p.Id] as path, p.Id, coalesce(p.Parent, 0) 
			from posts p join init on p.id = init.id
				union all
			select ph.path || array [p.id] as path, p.id, coalesce(p.parent, 0)
			from posts p
				join ph on p.parent = ph.id
		)
`

	query += `
			Select p.Author, p.Created, t.Forum, p.Id, p.is_edited, p.Message, coalesce(p.Parent, 0), p.Thread
				From ph
				join Posts p on p.id = ph.id
				join Threads t on t.Id = p.Thread
			`

	if desc {
		query += " order by ph.path[1] desc, ph.path[2:]"
	} else {
		query += " order by ph.path "
	}

	logs.Info("QUERY:\n", query)

	if hasSlug {
		rows, err = P.db.Query(query, thread.Slug, limit)

	} else {
		rows, err = P.db.Query(query, thread.Id, limit)
	}

	if err != nil {
		return errors.Wrap(err, "select by thread id parent tree error")
	}
	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}

		if err := rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited,
			&post.Message, &post.Parent, &post.Thread); err != nil {
			return errors.Wrap(err, "select by thread id tree scan error")
		}

		*posts = append(*posts, post)
	}
	return nil
}
