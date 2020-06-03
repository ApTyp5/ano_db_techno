package store

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type PostStore interface {
	Count(amount *uint) error
	SelectById(post *models.Post) error
	UpdateById(post *models.Post) error                                         // Edit
	InsertPostsByThreadSlug(thread *models.Thread, posts *[]*models.Post) error // thread.AddPosts
	InsertPostsByThreadId(thread *models.Thread, posts *[]*models.Post) error   // thread.AddPosts
	// threads.Posts
	SelectByThreadIdFlat(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
	// threads.Posts
	SelectByThreadIdTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
	// threads.Posts
	SelectByThreadIdParentTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error
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
		select PostNum from Status;
`)
	if err := row.Scan(amount); err != nil {
		return errors.Wrap(err, prefix)
	}

	return nil
}

func (P PSQLPostStore) SelectById(post *models.Post) error {
	prefix := "PSQL PostStore SelectById"
	row := P.db.QueryRow(`
		select u.NickName, p.Created, f.Slug, p.IsEdited, p.Message, postPar(p.*), p.Thread
			from Posts p
				join Users u on p.Author = u.Id
				join Threads t on p.Thread = t.Id
				join Forums f on t.Forum = f.Id
			where postId(p.*) = $1;
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
			where PostId(p.*) = $2
		returning 
			(select u.NickName from Users u join Posts p on p.Author = u.Id where PostId(p.*) = $2), 
		    Created, 
		    (select f.Slug from Posts p join Threads t on t.Id = p.Thread join Forums f on f.Id = t.Forum where PostId(p.*) = $2), 
		    IsEdited, Message, PostPar(p.*), Thread;
`, post.Message, post.Id)

	if err := row.Scan(&post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread); err != nil {
		return errors.Wrap(err, prefix)
	}

	return nil
}

func (P PSQLPostStore) InsertPostsByThreadSlug(thread *models.Thread, posts *[]*models.Post) error {
	if len(*posts) == 0 {
		return nil
	}

	valueArgs := make([]string, 0, len(*posts))
	for i := range *posts {
		nick := (*posts)[i].Author
		thsl := thread.Slug
		mess := (*posts)[i].Message
		parn := strconv.FormatInt(int64((*posts)[i].Parent), 10)

		if parn == "0" {
			valueArgs = append(valueArgs, "("+
				"(select Id from Users where nickname = '"+nick+"'),"+
				"(select Id from Threads where Slug = '"+thsl+"'),'"+
				mess+"', '{}')")
		} else {
			valueArgs = append(valueArgs, "("+
				"(select Id from Users where nickname = '"+nick+"'),"+
				"(select Id from Threads where Slug = '"+thsl+"'),'"+
				mess+"',"+
				"(select IdPath from Posts p where postId(p.*) = "+parn+")"+")")
		}

	}

	insertQuery := ` insert into Posts (Author, Thread, Message, IdPath)
					values` + strings.Join(valueArgs, ",")

	returnQuery := ` returning PostId(Posts.*), (select f.Slug from forums f join threads t on f.id = t.forum where t.id = thread),
                    (select u.NickName from Users u where Id = author), Thread, Created, IsEdited, Message, PostPar(Posts.*);`

	rows, err := P.db.Query(insertQuery + returnQuery)

	if err != nil {
		return errors.Wrap(err, "PSQLPostStore insertPostsByThread's id error")
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		post := &models.Post{}
		if err := rows.Scan(&(*posts)[i].Id, &(*posts)[i].Forum, &(*posts)[i].Author, &(*posts)[i].Thread, &(*posts)[i].Created,
			&post.IsEdited, &post.Message, &post.Parent); err != nil {
			return errors.Wrap(err, "PSQLPostStore insertPostsByThread's id SCAN error")
		}
		i++
	}

	return nil
}

func (P PSQLPostStore) InsertPostsByThreadId(thread *models.Thread, posts *[]*models.Post) error {
	if len(*posts) == 0 {
		return nil
	}

	valueArgs := make([]string, 0, len(*posts))
	for i := range *posts {
		nick := (*posts)[i].Author
		thid := strconv.FormatInt(int64(thread.Id), 10)
		mess := (*posts)[i].Message
		parn := strconv.FormatInt(int64((*posts)[i].Parent), 10)

		if parn == "0" {
			valueArgs = append(valueArgs, "("+
				"(select Id from Users where nickname = '"+nick+"'),"+
				thid+",'"+
				mess+"', '{}')")
		} else {
			valueArgs = append(valueArgs, "("+
				"(select Id from Users where nickname = '"+nick+"'),"+
				thid+",'"+
				mess+"',"+
				"(select IdPath from Posts where Id = "+parn+")"+
				")")
		}
	}

	insertQuery := ` 
					insert into Posts (Author, Thread, Message, IdPath)
					values` + strings.Join(valueArgs, ",")

	returnQuery := ` 
					returning PostId(Posts.*), (select f.Slug from forums f join threads t on f.id = t.forum where t.id = thread),
                    (select u.NickName from Users u where Id = author), Thread, Created, IsEdited, Message, PostPar(Posts.*);`

	rows, err := P.db.Query(insertQuery + returnQuery)

	if err != nil {
		return errors.Wrap(err, "PSQLPostStore insertPostsByThread's id error")
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		post := &models.Post{}
		if err := rows.Scan(&(*posts)[i].Id, &(*posts)[i].Forum, &(*posts)[i].Author, &(*posts)[i].Thread, &(*posts)[i].Created,
			&post.IsEdited, &post.Message, &post.Parent); err != nil {
			return errors.Wrap(err, "PSQLPostStore insertPostsByThread's id SCAN error")
		}
		i++
	}

	return nil
}

func (P PSQLPostStore) SelectByThreadIdFlat(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	hasSince := since >= 0

	query := `
		Select u.NickName, p.Created, f.Slug, postId(p.*), p.IsEdited, p.Message, postPar(p.*), p.Thread
		From Posts p
			join Users u on p.Author = u.Id
			join Threads t on t.Id = p.Thread
			join Forums f on f.Id = t.Forum
`
	if thread.Slug == "" {
		query += " where t.Id = $1 "
	} else {
		query += " where t.Slug = $1 "
	}

	if desc {
		if hasSince {
			query += " and postId(p.*) < $3 "
		}
		query += " Order By p.Created Desc, postId(p.*) Desc"
	} else {
		if hasSince {
			query += " and postId(p.*) > $3"
		}
		query += " Order By p.Created, postId(p.*)"
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

func (P PSQLPostStore) SelectByThreadIdTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	var (
		hasSince = since >= 0
		hasSlug  = thread.Slug != ""
		rows     *sql.Rows
		err      error
		query    = ""
		initPath = ""
	)

	if hasSince {
		initPath += " (select idpath from posts p "
		if hasSlug {
			initPath += " where Thread = (select Id from Threads where Slug = $1) "
		} else {
			initPath += " where Thread = $1 "
		}

		initPath += " and postId(p.*) = $3 ) "
	}

	query += `
			Select u.NickName, p.Created, f.Slug, postId(p.*), p.IsEdited, p.Message, postPar(p.*), p.Thread
				From Posts p
				join Users u on p.Author = u.Id
				join Threads t on t.Id = p.Thread
				join Forums f on f.Id = t.Forum
			`
	if hasSlug {
		query += " where Thread = (select Id from Threads where Slug = $1) "
	} else {
		query += " where Thread = $1 "
	}

	if hasSince {
		if desc {
			query += " and p.IdPath < " + initPath
		} else {
			query += " and p.IdPath > " + initPath
		}
	}

	if desc {
		query += " order by p.IdPath desc "
	} else {
		query += " order by p.IdPath "
	}

	if limit != 0 {
		query += " LIMIT $2; "
	}

	if hasSlug {
		if hasSince {
			rows, err = P.db.Query(query, thread.Slug, limit, since)
		} else {
			rows, err = P.db.Query(query, thread.Slug, limit)
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

func (P PSQLPostStore) SelectByThreadIdParentTree(posts *[]*models.Post, thread *models.Thread, limit int, since int, desc bool) error {
	var (
		hasLimit = limit > 0
		hasSince = since >= 0
		hasSlug  = thread.Slug != ""
		rows     *sql.Rows
		err      error
	)

	query := `
				with initialPosts as (
					select IdPath 
					from posts p
				`
	if hasSlug {
		query += " where Thread = (select Id from Threads where Slug = $1) "
	} else {
		query += " where Thread = $1 "
	}

	query += " and PostPar(p.*) = 0 "

	if hasSince {
		initPath := " (select IdPath[1] from posts p "
		if hasSlug {
			initPath += " where Thread = (select Id from Threads where Slug = $1) "
		} else {
			initPath += " where Thread = $1 "
		}

		initPath += " and postId(p.*) = $3 ) "

		if desc {
			query += " and p.IdPath[1] < " + initPath
		} else {
			query += " and p.IdPath[1] > " + initPath
		}
	}

	if desc {
		query += " order by p.IdPath[1] desc, p.IdPath[2:]"
	} else {
		query += " order by p.IdPath "
	}

	if hasLimit {
		query += " limit $2 "
	}

	query += " ) "

	query += `
			Select u.NickName, p.Created, f.Slug, postId(p.*), p.IsEdited, p.Message, postPar(p.*), p.Thread
				From initialPosts ip
				join Posts p on p.IdPath[1] = ip.IdPath[1]
				join Users u on p.Author = u.Id
				join Threads t on t.Id = p.Thread
				join Forums f on f.Id = t.Forum
			`

	if desc {
		query += " order by p.IdPath[1] desc, p.IdPath[2:]"
	} else {
		query += " order by p.IdPath "
	}

	if hasSlug {
		if hasSince {
			rows, err = P.db.Query(query, thread.Slug, limit, since)
		} else {
			rows, err = P.db.Query(query, thread.Slug, limit)
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
