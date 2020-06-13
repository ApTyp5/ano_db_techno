package usecases

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/repositories"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
)

type PostUseCase interface {
	Details(postFull *models.PostFull, related []string) (int, interface{}) // /post/{id}/details
	Edit(post *models.Post) (int, interface{})                              // /post/{id}/details
}

type RDBPostUseCase struct {
	ps repositories.PostRepo
	us repositories.UserRepo
	fs repositories.ForumRepo
	ts repositories.ThreadRepo
}

func CreateRDBPostUseCase(db *pgx.ConnPool) PostUseCase {
	return RDBPostUseCase{
		ps: repositories.CreatePSQLPostRepo(db),
		us: repositories.CreatePSQLUserRepo(db),
		fs: repositories.CreatePSQLForumRepo(db),
		ts: repositories.CreatePSQLThreadRepo(db),
	}
}

func (uc RDBPostUseCase) Details(postFull *models.PostFull, related []string) (int, interface{}) {

	if err := uc.ps.SelectById(postFull.Post); err != nil {
		return http.StatusNotFound, wrapStrError("post not found")
	}

	for _, str := range related {
		switch str {
		case "user":
			postFull.Author = &models.User{NickName: postFull.Post.Author}
			if err := uc.us.SelectByNickname(postFull.Author); err != nil {
				return http.StatusInternalServerError, wrapError(errors.Wrap(err, "user"))
			}
		case "forum":
			postFull.Forum = &models.Forum{}
			postFull.Forum.Slug = postFull.Post.Forum
			if err := uc.fs.SelectBySlug(postFull.Forum); err != nil {
				return http.StatusInternalServerError, wrapError(errors.Wrap(err, "forum"))

			}
		case "thread":
			postFull.Thread = &models.Thread{}
			postFull.Thread.Id = postFull.Post.Thread
			if err := uc.ts.SelectBySlugOrId(postFull.Thread); err != nil {
				return http.StatusInternalServerError, wrapError(errors.Wrap(err, "thread"))
			}
		}
	}

	return http.StatusOK, postFull
}

func (uc RDBPostUseCase) Edit(post *models.Post) (int, interface{}) {
	if err := uc.ps.UpdateById(post); err == pgx.ErrNoRows {
		return http.StatusNotFound, wrapStrError("post not found")
	} else if err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusOK, post
}
