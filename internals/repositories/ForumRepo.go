package repositories

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
)

type ForumRepo interface {
	SelectBySlug(forum *models.Forum) error
	Insert(forum *models.Forum) error
	Count(num *uint) error
}

type PSQLForumRepo struct {
	db           *pgx.ConnPool
	selectBySlug *pgx.PreparedStatement
	insert       *pgx.PreparedStatement
	count        *pgx.PreparedStatement
}

func CreatePSQLForumRepo(db *pgx.ConnPool) ForumRepo {
	var err error
	prefix := "forum_"
	repo := PSQLForumRepo{
		db: db,
	}

	repo.selectBySlug, err = db.Prepare(prefix+"selectBySlug", `
		SELECT post_num, thread_num, title, slug, responsible
			FROM forums
		WHERE slug = $1;
	`)
	panicIfErr(err)

	repo.insert, err = db.Prepare(prefix+"insert", `
		INSERT INTO FORUMS (slug, title, responsible)
		VALUES ($1, $2, (select nick_name from Users where nick_name = $3))
		RETURNING slug, title, responsible, post_num, thread_num
	`)
	panicIfErr(err)

	repo.count, err = db.Prepare(prefix+"count", "SELECT forum_num FROM status;")
	panicIfErr(err)

	return repo
}

func (forumRepo PSQLForumRepo) SelectBySlug(forum *models.Forum) error {
	return forumRepo.db.QueryRow(
		forumRepo.selectBySlug.Name,
		forum.Slug).Scan(
		&forum.Posts,
		&forum.Threads,
		&forum.Title,
		&forum.Slug,
		&forum.User)
}

func (forumRepo PSQLForumRepo) Insert(forum *models.Forum) error {
	return forumRepo.db.QueryRow(
		forumRepo.insert.Name,
		forum.Slug,
		forum.Title,
		forum.User).Scan(&forum.Slug,
		&forum.Title,
		&forum.User,
		&forum.Posts,
		&forum.Threads)
}

func (forumRepo PSQLForumRepo) Count(num *uint) error {
	return forumRepo.db.QueryRow(forumRepo.count.Name).Scan(num)
}
