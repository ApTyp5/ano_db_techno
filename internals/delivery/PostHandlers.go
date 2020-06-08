package delivery

import (
	"github.com/ApTyp5/new_db_techno/internals/delivery/args"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecase"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	. "github.com/valyala/fasthttp"
	"strings"
)

type PostHandlerManager struct {
	uc usecase.PostUseCase
}

func CreatePostHandlerManager(db *pgx.ConnPool) PostHandlerManager {
	return PostHandlerManager{uc: usecase.CreateRDBPostUseCase(db)}
}

// /post/{id}/details
func (m PostHandlerManager) Details() RequestHandler {

	return func(ctx *RequestCtx) {
		var (
			prefix   = "post details handler:"
			postFull = models.PostFull{Post: &models.Post{}}
			err      error
		)
		if postFull.Post.Id, err = args.PathInt("id", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			return
		}

		related := args.QueryString("related", ctx)

		if ctx.SetStatusCode(m.uc.Details(&postFull, &err, strings.Split(related, ","))); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&postFull, ctx)
	}
}

// /post/{id}/details
func (m PostHandlerManager) Edit() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "post edit handler:"
			post   models.Post
			err    error
		)
		if err = args.GetBodyInterface(&post, ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if post.Id, err = args.PathInt("id", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if ctx.SetStatusCode(m.uc.Edit(&post, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&post, ctx)
	}
}
