package usecase

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
)

type PostUseCase interface {
	Details(postFull *models.PostFull, related []string) (int, interface{}) // /post/{id}/details
	Edit(post *models.Post) (int, interface{})                              // /post/{id}/details
}

type RDBPostUseCase struct {
	ps store.PostStore
	us store.UserStore
	fs store.ForumStore
	ts store.ThreadStore
}

func CreateRDBPostUseCase(db *pgx.ConnPool) PostUseCase {
	return RDBPostUseCase{
		ps: store.CreatePSQLPostStore(db),
		us: store.CreatePSQLUserStore(db),
		fs: store.CreatePSQLForumStore(db),
		ts: store.CreatePSQLThreadStore(db),
	}
}

func (uc RDBPostUseCase) Details(postFull *models.PostFull, related []string) (int, interface{}) {
	prefix := "RDBPostUseCase details"

	logs.Info("post id: ", postFull.Post.Id)

	if err := errors.Wrap(uc.ps.SelectById(postFull.Post), prefix); err != nil {
		return http.StatusNotFound, wrapStrError("post not found")
	}

	for _, str := range related {

		switch str {
		case "user":
			postFull.Author = &models.User{NickName: postFull.Post.Author}
			if err := uc.us.SelectByNickname(postFull.Author); err != nil {
				logs.Error(errors.Wrap(err, "unexpected user repo error"))
			}
		case "forum":
			postFull.Forum = &models.Forum{}
			postFull.Forum.Slug = postFull.Post.Forum
			if err := uc.fs.SelectBySlug(postFull.Forum); err != nil {
				logs.Error(errors.Wrap(err, "unxepected forum repo error"))
			}
		case "thread":
			postFull.Thread = &models.Thread{}
			postFull.Thread.Id = postFull.Post.Thread
			if err := uc.ts.SelectById(postFull.Thread); err != nil {
				logs.Error(errors.Wrap(err, "unexpected thread repo error"))
			}
		default:
			logs.Error(errors.New("unexpected related value: " + str))
		}
	}

	return http.StatusOK, postFull
}

func (uc RDBPostUseCase) Edit(post *models.Post) (int, interface{}) {
	if post.Message == "" {
		if err := errors.Wrap(uc.ps.SelectById(post), "RDBPostUseCase Edit"); err != nil {
			return http.StatusNotFound, wrapStrError("post not found")
		}
		return http.StatusOK, post
	}

	if err := errors.Wrap(uc.ps.UpdateById(post), "RDBPostUseCase Edit"); err != nil {
		return http.StatusNotFound, wrapStrError("post not found")
	}
	return http.StatusOK, post
}
