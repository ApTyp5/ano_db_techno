package usecase

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
	"strings"
)

type ThreadUseCase interface {
	// /thread/{slug_or_id}/create
	AddPosts(thread *models.Thread, posts *[]*models.Post, err *error) int
	Details(thread *models.Thread, err *error) int // /thread/{slug_or_id}/details
	Edit(thread *models.Thread, err *error) int    // /thread/{slug_or_id}/details
	// /thread/{slug_or_id}/posts
	Posts(posts *[]*models.Post, thread *models.Thread, err *error, limit int, since int, sort string, desc bool) int
	Vote(thread *models.Thread, vote *models.Vote, err *error) int // /thread/{slug_or_id}/vote
}

type RDBThreadUseCase struct {
	ts store.ThreadStore
	ps store.PostStore
	vs store.VoteStore
	us store.UserStore
}

func CreateRDBThreadUseCase(db *sql.DB) ThreadUseCase {
	return RDBThreadUseCase{
		ts: store.CreatePSQLThreadStore(db),
		ps: store.CreatePSQLPostStore(db),
		vs: store.CreatePSQLVoteStore(db),
		us: store.CreatePSQLUserStore(db),
	}
}

func (uc RDBThreadUseCase) AddPosts(thread *models.Thread, posts *[]*models.Post, err *error) int {
	prefix := "RDB thread use case add posts"

	if *err = errors.Wrap(uc.ps.InsertPostsByThread(thread, posts), prefix); *err != nil {
		logs.Info("ERROR:", (*err).Error())
		if strings.Index(errors.Cause(*err).Error(), "posts_parent") >= 0 ||
			strings.Index(errors.Cause(*err).Error(), "another") >= 0 {
			return 409
		}
		return 404
	}

	return 201
}

func (uc RDBThreadUseCase) Details(thread *models.Thread, err *error) int {
	prefix := "RDB thread use case details"
	if thread.Slug != "" {
		if *err = errors.Wrap(uc.ts.SelectBySlug(thread), prefix); *err != nil {
			return 404
		}
		return 200
	}

	if *err = errors.Wrap(uc.ts.SelectById(thread), prefix); *err != nil {
		return 404
	}

	return 200
}

func (uc RDBThreadUseCase) Edit(thread *models.Thread, err *error) int {
	prefix := "RDB thread use case edit"
	if thread.Slug != "" {
		if thread.Title == "" && thread.Message == "" {
			if *err = errors.Wrap(uc.ts.SelectBySlug(thread), prefix); *err != nil {
				return 404
			}
			return 200
		}
		if *err = errors.Wrap(uc.ts.UpdateBySlug(thread), prefix); *err != nil {
			return 404
		}

		return 200
	}

	if thread.Title == "" && thread.Message == "" {
		if *err = errors.Wrap(uc.ts.SelectById(thread), prefix); *err != nil {
			return 404
		}
		return 200
	}

	if *err = errors.Wrap(uc.ts.UpdateById(thread), prefix); *err != nil {
		return 404
	}

	return 200
}

func (uc RDBThreadUseCase) Posts(posts *[]*models.Post, thread *models.Thread, err *error, limit int, since int, sort string, desc bool) int {
	prefix := "RDB thread use case posts"

	if thread.Slug == "" {
		if *err = errors.Wrap(uc.ts.SelectById(thread), prefix); *err != nil {
			return 404
		}
	} else {
		if *err = errors.Wrap(uc.ts.SelectBySlug(thread), prefix); *err != nil {
			return 404
		}
	}

	switch sort {
	case "tree":
		if *err = errors.Wrap(uc.ps.SelectByThreadTree(posts, thread, limit, since, desc), prefix); *err != nil {
			return 404
		}
		return 200

	case "parent_tree":
		if *err = errors.Wrap(uc.ps.SelectByThreadParentTree(posts, thread, limit, since, desc), prefix); *err != nil {
			return 404
		}
		return 200
	}

	if *err = errors.Wrap(uc.ps.SelectByThreadFlat(posts, thread, limit, since, desc), prefix); *err != nil {
		return 404
	}
	return 200
}

func (uc RDBThreadUseCase) Vote(thread *models.Thread, vote *models.Vote, err *error) int {

	if *err = errors.Wrap(uc.vs.Insert(vote, thread), "RDB thread use case vote"); *err != nil {
		if strings.Index(errors.Cause(*err).Error(), "author") >= 0 {
			return 404
		}
	}

	if *err = errors.Wrap(uc.vs.Update(vote, thread), "RDB thread use case vote"); *err != nil {
		return 404
	}

	return 200
}
