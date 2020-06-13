package usecases

import (
	_const "github.com/ApTyp5/new_db_techno/const"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/repositories"
	"github.com/jackc/pgx"
	"net/http"
)

type ForumUseCase interface {
	Create(forum *models.Forum) (int, interface{})
	CreateThread(thread *models.Thread) (int, interface{})
	Details(slug string) (int, interface{})
	Threads(slug string, limit int, since string, desc bool) (int, interface{})
	Users(slug string, limit int, since string, desc bool) (int, interface{})
}

type RDBForumUseCase struct {
	fs repositories.ForumRepo
	ts repositories.ThreadRepo
	us repositories.UserRepo
}

func CreateRDBForumUseCase(db *pgx.ConnPool) ForumUseCase {
	return RDBForumUseCase{
		fs: repositories.CreatePSQLForumRepo(db),
		ts: repositories.CreatePSQLThreadRepo(db),
		us: repositories.CreatePSQLUserRepo(db),
	}
}

func (forumUseCase RDBForumUseCase) Create(forum *models.Forum) (int, interface{}) {
	var err error
	if err = forumUseCase.fs.SelectBySlug(forum); err == nil {
		return http.StatusConflict, forum
	}
	if err != pgx.ErrNoRows {
		return http.StatusInternalServerError, wrapError(err)
	}

	if err = forumUseCase.us.SelectByNickname(&models.User{NickName: forum.User}); err == pgx.ErrNoRows {
		return http.StatusNotFound, wrapStrError("Author not found")
	}

	if err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	if err = forumUseCase.fs.Insert(forum); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusCreated, forum
}

func (forumUseCase RDBForumUseCase) CreateThread(thread *models.Thread) (int, interface{}) {
	var err error
	if err = forumUseCase.ts.SelectBySlugOrId(thread); err == nil {
		return http.StatusConflict, thread
	}

	if err != pgx.ErrNoRows {
		return http.StatusInternalServerError, wrapError(err)
	}

	user := &models.User{NickName: thread.Author}
	if err = forumUseCase.us.SelectByNickname(user); err == pgx.ErrNoRows {
		return http.StatusNotFound, wrapStrError("user not found")
	}

	if err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	if err = forumUseCase.fs.SelectBySlug(&models.Forum{Slug: thread.Forum}); err == pgx.ErrNoRows {
		return http.StatusNotFound, wrapStrError("forum not found")
	}

	if err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	if err := forumUseCase.ts.Insert(thread); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusCreated, thread
}

func (forumUseCase RDBForumUseCase) Details(slug string) (int, interface{}) {
	forum := &models.Forum{Slug: slug}
	if err := forumUseCase.fs.SelectBySlug(forum); err == nil {
		return http.StatusOK, forum
	} else if err == pgx.ErrNoRows {
		return http.StatusNotFound, wrapStrError("forum not found")
	} else {
		return http.StatusInternalServerError, wrapError(err)
	}
}

func (forumUseCase RDBForumUseCase) Threads(slug string, limit int, since string, desc bool) (int, interface{}) {
	forum := &models.Forum{Slug: slug}
	if err := forumUseCase.fs.SelectBySlug(forum); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("forum not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}

	threads := make([]models.Thread, 0, _const.BuffSize)
	if err := forumUseCase.ts.SelectByForum(&threads, forum, limit, since, desc); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusOK, &threads
}

func (forumUseCase RDBForumUseCase) Users(slug string, limit int, since string, desc bool) (int, interface{}) {
	forum := &models.Forum{Slug: slug}

	if err := forumUseCase.fs.SelectBySlug(forum); err != nil {

		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("forum not found")
		}
		return http.StatusInternalServerError, err
	}

	users := make([]models.User, 0, _const.BuffSize)
	if err := forumUseCase.us.SelectByForum(&users, forum, limit, since, desc); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, &users
}
