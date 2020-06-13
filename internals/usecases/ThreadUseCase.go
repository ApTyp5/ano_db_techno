package usecases

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/repositories"
	"github.com/jackc/pgx"
	"net/http"
	"strings"
)

type ThreadUseCase interface {
	// /thread/{slug_or_id}/create
	AddPosts(thread *models.Thread, posts []models.Post) (int, interface{})
	Details(thread *models.Thread) (int, interface{}) // /thread/{slug_or_id}/details
	Edit(thread *models.Thread) (int, interface{})    // /thread/{slug_or_id}/details
	// /thread/{slug_or_id}/posts
	Posts(posts *[]models.Post, thread *models.Thread, limit int, since int, sort string, desc bool) (int, interface{})
	Vote(thread *models.Thread, vote *models.Vote) (int, interface{}) // /thread/{slug_or_id}/vote
}

type RDBThreadUseCase struct {
	ts repositories.ThreadRepo
	ps repositories.PostRepo
	vs repositories.VoteRepo
	us repositories.UserRepo
}

func CreateRDBThreadUseCase(db *pgx.ConnPool) ThreadUseCase {
	return RDBThreadUseCase{
		ts: repositories.CreatePSQLThreadRepo(db),
		ps: repositories.CreatePSQLPostRepo(db),
		vs: repositories.CreatePSQLVoteRepo(db),
		us: repositories.CreatePSQLUserRepo(db),
	}
}

func (uc RDBThreadUseCase) AddPosts(thread *models.Thread, posts []models.Post) (int, interface{}) {
	if err := uc.ts.SelectBySlugOrId(thread); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("thread not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}

	nicks := make(map[string]bool)
	for i := range posts {
		if !nicks[posts[i].Author] {
			nicks[posts[i].Author] = true
		}
	}
	if err := uc.us.CheckExistance(nicks); err != nil {
		return http.StatusNotFound, wrapError(err)
	}

	if err := uc.ps.InsertPostsByThread(thread, posts); err != nil {
		if strings.Index(err.Error(), "posts_parent") >= 0 ||
			strings.Index(err.Error(), "another") >= 0 {
			return http.StatusConflict, wrapStrError("posts_parent or another conflict")
		}

		return http.StatusInternalServerError, wrapError(err)
	}

	if err := uc.us.AddForumUsers(nicks, thread.Forum); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusCreated, posts
}

func (uc RDBThreadUseCase) Details(thread *models.Thread) (int, interface{}) {
	if err := uc.ts.SelectBySlugOrId(thread); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("thread not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}
	return http.StatusOK, thread
}

func (uc RDBThreadUseCase) Edit(thread *models.Thread) (int, interface{}) {
	if err := uc.ts.Update(thread); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("thread not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}
	return http.StatusOK, thread
}

func (uc RDBThreadUseCase) Posts(posts *[]models.Post, thread *models.Thread, limit int, since int, sort string, desc bool) (int, interface{}) {
	if err := uc.ts.SelectBySlugOrId(thread); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("thread not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}

	if err := uc.ps.SelectByThread(posts, thread, limit, since, desc, sort); err != nil {
		if err != pgx.ErrNoRows {
			return http.StatusInternalServerError, wrapError(err)
		}
	}

	return http.StatusOK, posts
}

func (uc RDBThreadUseCase) Vote(thread *models.Thread, vote *models.Vote) (int, interface{}) {
	if err := uc.vs.InsertOrUpdate(vote, thread); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("user or thread not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusOK, thread
}
