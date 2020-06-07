package usecase

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/pkg/errors"
)

type ForumUseCase interface {
	Create(forum *models.Forum, err *error) int
	CreateThread(thread *models.Thread, err *error) int
	Details(forum *models.Forum, err *error) int
	Threads(threads *[]*models.Thread, forum *models.Forum, err *error, limit int, since string, desc bool) int
	Users(users *[]*models.User, forum *models.Forum, err *error, limit int, since string, desc bool) int
}

type RDBForumUseCase struct {
	fs store.ForumStore
	ts store.ThreadStore
	us store.UserStore
}

func CreateRDBForumUseCase(db *sql.DB) ForumUseCase {
	return RDBForumUseCase{
		fs: store.CreatePSQLForumStore(db),
		ts: store.CreatePSQLThreadStore(db),
		us: store.CreatePSQLUserStore(db),
	}
}

func (uc RDBForumUseCase) Create(forum *models.Forum, err *error) int {
	prefix := "RDBForumUseCase create"

	if *err = errors.Wrap(uc.fs.SelectBySlug(forum), prefix); *err == nil {
		return 409
	}

	if *err = errors.Wrap(uc.fs.Insert(forum), prefix); *err == nil {
		return 201
	}

	return 404
}

func (uc RDBForumUseCase) CreateThread(thread *models.Thread, err *error) int {
	prefix := "RDBForumUseCase createThread"

	if thread.Slug != "" {
		if *err = errors.Wrap(uc.ts.SelectBySlug(thread), prefix); *err == nil {
			return 409
		}
	}

	if *err = errors.Wrap(uc.ts.Insert(thread), prefix); *err == nil {
		return 201
	}

	return 404
}

func (uc RDBForumUseCase) Details(forum *models.Forum, err *error) int {
	prefix := "RDBForumUseCase details"
	if *err = errors.Wrap(uc.fs.SelectBySlug(forum), prefix); *err == nil {
		return 200
	}

	return 404
}

func (uc RDBForumUseCase) Threads(threads *[]*models.Thread, forum *models.Forum, err *error, limit int, since string, desc bool) int {
	prefix := "RDBForumUseCase threads"

	if *err = errors.Wrap(uc.ts.SelectByForum(threads, forum, limit, since, desc), prefix); *err == nil {
		if len(*threads) != 0 {
			return 200
		}
	}

	if *err = errors.Wrap(uc.fs.SelectBySlug(forum), prefix); *err == nil {
		return 200
	}

	return 404
}

func (uc RDBForumUseCase) Users(users *[]*models.User, forum *models.Forum, err *error, limit int, since string, desc bool) int {
	prefix := "RDBForumUseCase users"

	if *err = uc.fs.SelectBySlug(forum); *err != nil {
		return 404
	}

	if *err = errors.Wrap(uc.us.SelectByForum(users, forum, limit, since, desc), prefix); *err == nil {
		return 200
	}

	return 404
}
