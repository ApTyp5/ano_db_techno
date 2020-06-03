package usecase

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
)

type PostUseCase interface {
	Details(postFull *models.PostFull, err *error, related []string) int // /post/{id}/details
	Edit(post *models.Post, err *error) int                              // /post/{id}/details
}

type RDBPostUseCase struct {
	ps store.PostStore
	us store.UserStore
	fs store.ForumStore
	ts store.ThreadStore
}

func CreateRDBPostUseCase(db *sql.DB) PostUseCase {
	return RDBPostUseCase{
		ps: store.CreatePSQLPostStore(db),
		us: store.CreatePSQLUserStore(db),
		fs: store.CreatePSQLForumStore(db),
		ts: store.CreatePSQLThreadStore(db),
	}
}

func (uc RDBPostUseCase) Details(postFull *models.PostFull, err *error, related []string) int {
	prefix := "RDBPostUseCase details"

	if *err = errors.Wrap(uc.ps.SelectById(postFull.Post), prefix); *err != nil {
		return 404
	}

	for _, str := range related {
		switch str {
		case "user":
			postFull.Author.NickName = postFull.Post.Author
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

	return 200
}

func (uc RDBPostUseCase) Edit(post *models.Post, err *error) int {
	if *err = errors.Wrap(uc.ps.UpdateById(post), "RDBPostUseCase Edit"); *err != nil {
		return 404
	}

	return 200
}
