package usecases

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/repositories"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

type ThreadUseCase interface {
	// /thread/{slug_or_id}/create
	AddPosts(thread *models.Thread, posts *[]models.Post) (int, interface{})
	Details(thread *models.Thread) (int, interface{}) // /thread/{slug_or_id}/details
	Edit(thread *models.Thread) (int, interface{})    // /thread/{slug_or_id}/details
	// /thread/{slug_or_id}/posts
	Posts(posts *[]*models.Post, thread *models.Thread,
		limit int, since int, sort string, desc bool) (int, interface{})
	Vote(thread *models.Thread, vote *models.Vote) (int, interface{}) // /thread/{slug_or_id}/vote
}

type RDBThreadUseCase struct {
	ts repositories.ThreadStore
	ps repositories.PostStore
	vs repositories.VoteStore
	us repositories.UserStore
}

func CreateRDBThreadUseCase(db *pgx.ConnPool) ThreadUseCase {
	return RDBThreadUseCase{
		ts: repositories.CreatePSQLThreadStore(db),
		ps: repositories.CreatePSQLPostStore(db),
		vs: repositories.CreatePSQLVoteStore(db),
		us: repositories.CreatePSQLUserStore(db),
	}
}

func (uc RDBThreadUseCase) AddPosts(thread *models.Thread, posts *[]models.Post) (int, interface{}) {
	prefix := "RDB thread use case add posts"

	if err := errors.Wrap(uc.ps.InsertPostsByThread(thread, posts), prefix); err != nil {
		if strings.Index(errors.Cause(err).Error(), "posts_parent") >= 0 ||
			strings.Index(errors.Cause(err).Error(), "another") >= 0 {
			return http.StatusConflict, wrapStrError("posts_parent or another conflict")
		}

		return http.StatusNotFound, wrapStrError("user not found")
	}

	return http.StatusCreated, posts
}

func (uc RDBThreadUseCase) Details(thread *models.Thread) (int, interface{}) {
	prefix := "RDB thread use case details"

	if thread.Id < 0 {
		if err := errors.Wrap(uc.ts.SelectBySlug(thread), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("Thread not found")
		}

		return http.StatusOK, thread
	}

	if err := errors.Wrap(uc.ts.SelectById(thread), prefix); err != nil {
		return http.StatusNotFound, wrapStrError("thread not found")
	}

	return http.StatusOK, thread
}

func (uc RDBThreadUseCase) Edit(thread *models.Thread) (int, interface{}) {
	prefix := "RDB thread use case edit"

	if thread.Id < 0 {
		if thread.Title == "" && thread.Message == "" {
			if err := errors.Wrap(uc.ts.SelectBySlug(thread), prefix); err != nil {
				return http.StatusNotFound, wrapStrError("thread not found")
			}
			return http.StatusOK, thread
		}

		if err := errors.Wrap(uc.ts.UpdateBySlug(thread), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("Thread not found")
		}

		return http.StatusOK, thread
	}

	if thread.Title == "" && thread.Message == "" {
		if err := errors.Wrap(uc.ts.SelectById(thread), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("thread does not exists")
		}

		return http.StatusOK, thread
	}

	if err := errors.Wrap(uc.ts.UpdateById(thread), prefix); err != nil {
		return http.StatusNotFound, wrapStrError("thread not found")
	}

	return http.StatusOK, thread
}

func (uc RDBThreadUseCase) Posts(posts *[]*models.Post, thread *models.Thread,
	limit int, since int, sort string, desc bool) (int, interface{}) {
	prefix := "RDB thread use case posts"

	if thread.Id >= 0 {
		if err := errors.Wrap(uc.ts.SelectById(thread), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("thread not found: " + err.Error())
		}
	} else {
		if err := errors.Wrap(uc.ts.SelectBySlug(thread), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("thread not found: " + err.Error())
		}
	}

	switch sort {
	case "tree":
		if err := errors.Wrap(uc.ps.SelectByThreadTree(posts, thread, limit, since, desc), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("thread tree not found: " + err.Error())
		}
		return http.StatusOK, posts

	case "parent_tree":
		if err := errors.Wrap(uc.ps.SelectByThreadParentTree(posts, thread, limit, since, desc), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("thread parent tree not found: " + err.Error())
		}
		return http.StatusOK, posts
	}

	if err := errors.Wrap(uc.ps.SelectByThreadFlat(posts, thread, limit, since, desc), prefix); err != nil {
		return http.StatusNotFound, wrapStrError("thread flat not found: " + err.Error())
	}

	return http.StatusOK, posts
}

func (uc RDBThreadUseCase) Vote(thread *models.Thread, vote *models.Vote) (int, interface{}) {

	if err := uc.vs.InsertOrUpdate(vote, thread); err != nil {
		return http.StatusNotFound, wrapError(err)
	}

	return http.StatusOK, thread
}
