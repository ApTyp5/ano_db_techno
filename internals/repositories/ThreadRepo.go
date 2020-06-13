package repositories

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
)

type ThreadRepo interface {
	Count(amount *uint) error
	Insert(thread *models.Thread) error                                                                    // forum.AddThread
	SelectByForum(threads *[]models.Thread, forum *models.Forum, limit int, since string, desc bool) error // forum.GetThreads
	////////////////////////
	SelectBySlugOrId(thread *models.Thread) error // Details
	Update(thread *models.Thread) error           // Edit
}

type PSQLThreadRepo struct {
	db               *pgx.ConnPool
	count            *pgx.PreparedStatement
	insert           *pgx.PreparedStatement
	selectByIdOrSlug *pgx.PreparedStatement
	updateByIdOrSlug *pgx.PreparedStatement
}

func CreatePSQLThreadRepo(db *pgx.ConnPool) ThreadRepo {
	prefix := "thread_"
	var err error
	repo := PSQLThreadRepo{db: db}

	repo.count, err = db.Prepare(prefix+"count", "select thread_num from Status")
	panicIfErr(err)

	repo.selectByIdOrSlug, err = db.Prepare(prefix+"selectByIdOrSlug", `
	SELECT id, author, forum, created, message, title, vote_num, coalesce(slug, '')
	FROM threads WHERE slug = $1 OR id = $2;`)
	panicIfErr(err)

	repo.insert, err = db.Prepare(prefix+"insert", `
			INSERT INTO threads (
			author, 
			forum, 
			message, 
			slug, 
			title, 
			created
		) VALUES (
			(SELECT nick_name FROM users where nick_name = $1), 
			(SELECT slug FROM forums WHERE slug = $2),
			$3, 
			nullif($4,''), 
			$5,
			nullif($6, to_timestamp(0))
		)
		returning 
			id, 
			author,
			forum,
			message,
			(coalesce(slug, '')), 
			title,
			created,
			vote_num;
`)
	panicIfErr(err)

	repo.updateByIdOrSlug, err = db.Prepare(prefix+"updateBySlugOrId", `
	UPDATE threads SET message = COALESCE(nullif($1, ''), message), title = COALESCE(nullif($2, ''), title)
		WHERE slug = $3 OR id = $4 
	returning 
		id, 
		author,
		forum,
		message,
		(coalesce(slug, '')), 
		title,
		created,
		vote_num;
	`)
	panicIfErr(err)

	return repo
}

func (threadRepo PSQLThreadRepo) Count(amount *uint) error {
	return threadRepo.db.QueryRow(
		threadRepo.count.Name).Scan(
		amount)
}

func (threadRepo PSQLThreadRepo) Insert(thread *models.Thread) error {
	return threadRepo.db.QueryRow(
		threadRepo.insert.Name,
		thread.Author,
		thread.Forum,
		thread.Message,
		thread.Slug,
		thread.Title,
		thread.Created).Scan(
		&thread.Id,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Slug,
		&thread.Title,
		&thread.Created,
		&thread.Votes)
}

func (threadRepo PSQLThreadRepo) SelectByForum(threads *[]models.Thread, forum *models.Forum,
	limit int, since string, desc bool) error {
	rows, err := threadRepo.db.Query("SELECT id, author, forum,"+
		"created, message, slug,"+
		"title, vote_num "+
		"from select_threads_by_forum($1, $2, nullif($3, ''), $4);",
		forum.Slug, limit, since, desc)

	if err != nil {
		return err
	}

	for rows.Next() {
		i := len(*threads)
		*threads = append(*threads, models.Thread{})
		if err := rows.Scan(&(*threads)[i].Id, &(*threads)[i].Author, &(*threads)[i].Forum,
			&(*threads)[i].Created, &(*threads)[i].Message, &(*threads)[i].Slug,
			&(*threads)[i].Title, &(*threads)[i].Votes); err != nil {
			return err
		}
	}

	return nil
}

func (threadRepo PSQLThreadRepo) SelectBySlugOrId(thread *models.Thread) error {
	return threadRepo.db.QueryRow(
		threadRepo.selectByIdOrSlug.Name,
		thread.Slug,
		thread.Id).Scan(
		&thread.Id,
		&thread.Author,
		&thread.Forum,
		&thread.Created,
		&thread.Message,
		&thread.Title,
		&thread.Votes,
		&thread.Slug)
}

func (threadRepo PSQLThreadRepo) Update(thread *models.Thread) error {
	return threadRepo.db.QueryRow(
		threadRepo.updateByIdOrSlug.Name,
		thread.Message,
		thread.Title,
		thread.Slug,
		thread.Id).Scan(
		&thread.Id,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Slug,
		&thread.Title,
		&thread.Created,
		&thread.Votes)
}
