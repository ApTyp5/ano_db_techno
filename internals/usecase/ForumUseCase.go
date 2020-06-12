package usecase

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
)

type ForumUseCase interface {
	Create(forum *models.Forum) (int, interface{})
	CreateThread(thread *models.Thread) (int, interface{})
	Details(forum *models.Forum) (int, interface{})
	Threads(threads *[]*models.Thread, forum *models.Forum, limit int, since string, desc bool) (int, interface{})
	Users(users *[]*models.User, forum *models.Forum, limit int, since string, desc bool) (int, interface{})
}

type RDBForumUseCase struct {
	fs store.ForumStore
	ts store.ThreadStore
	us store.UserStore
}

func CreateRDBForumUseCase(db *pgx.ConnPool) ForumUseCase {
	return RDBForumUseCase{
		fs: store.CreatePSQLForumStore(db),
		ts: store.CreatePSQLThreadStore(db),
		us: store.CreatePSQLUserStore(db),
	}
}

func (uc RDBForumUseCase) Create(forum *models.Forum) (int, interface{}) {
	prefix := "RDBForumUseCase create"
	if err := errors.Wrap(uc.fs.SelectBySlug(forum), prefix); err == nil {
		return http.StatusConflict, forum
	}
	if err := errors.Wrap(uc.fs.Insert(forum), prefix); err == nil {
		return http.StatusCreated, forum
	}
	return http.StatusNotFound, wrapStrError("Author not found")
}

func (uc RDBForumUseCase) CreateThread(thread *models.Thread) (int, interface{}) {
	prefix := "RDBForumUseCase createThread"
	if err := errors.Wrap(uc.ts.Insert(thread), prefix); err == nil {
		return http.StatusCreated, thread
	} else if errors.Cause(err).Error() == "conflict" {
		return http.StatusConflict, thread
	}
	return http.StatusNotFound, wrapStrError("Author or Forum not found")
}

func (uc RDBForumUseCase) Details(forum *models.Forum) (int, interface{}) {
	prefix := "RDBForumUseCase details"
	if err := errors.Wrap(uc.fs.SelectBySlug(forum), prefix); err == nil {
		return http.StatusOK, forum
	}

	return 404, wrapStrError("Forum not found")
}

func (uc RDBForumUseCase) Threads(threads *[]*models.Thread, forum *models.Forum,
	limit int, since string, desc bool) (int, interface{}) {
	prefix := "RDBForumUseCase threads"

	if err := errors.Wrap(uc.ts.SelectByForum(threads, forum, limit, since, desc), prefix); err == nil {
		if len(*threads) != 0 {
			return http.StatusOK, threads
		}
	}

	if err := errors.Wrap(uc.fs.SelectBySlug(forum), prefix); err == nil {
		return http.StatusOK, threads
	}

	return http.StatusNotFound, wrapStrError("forum not found")
}

func (uc RDBForumUseCase) Users(users *[]*models.User, forum *models.Forum,
	limit int, since string, desc bool) (int, interface{}) {
	prefix := "RDBForumUseCase users"

	if err := errors.Wrap(uc.us.SelectByForum(users, forum, limit, since, desc), prefix); err == nil {
		return http.StatusOK, users
	} else {
		logs.Info("error: ", err)
	}

	return http.StatusNotFound, wrapStrError("forum not found")
}
