package delivery

import (
	"database/sql"
	_const "github.com/ApTyp5/new_db_techno/const"
	"github.com/ApTyp5/new_db_techno/internals/delivery/args"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecase"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
	. "github.com/valyala/fasthttp"
)

type ThreadHandlerManager struct {
	uc usecase.ThreadUseCase
}

func CreateThreadHandlerManager(db *sql.DB) ThreadHandlerManager {
	return ThreadHandlerManager{
		uc: usecase.CreateRDBThreadUseCase(db),
	}
}

func (m ThreadHandlerManager) AddPosts() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "thread handler addPosts"
			posts  = make([]*models.Post, 0, _const.BuffSize)
			thread models.Thread
			err    error
		)
		if err = args.GetBodyInterface(&posts, ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			return
		}

		if thread.Id, err = args.PathInt("slug_or_id", ctx); err != nil {
			if thread.Slug, err = args.PathString("slug_or_id", ctx); err != nil {
				logs.Error(errors.Wrap(err, prefix))
				return
			}
		}

		if ctx.SetStatusCode(m.uc.AddPosts(&thread, &posts, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&posts, ctx)
	}
}

func (m ThreadHandlerManager) Details() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "thread handler details"
			thread models.Thread
			err    error
		)
		if thread.Id, err = args.PathInt("slug_or_id", ctx); err != nil {
			if thread.Slug, err = args.PathString("slug_or_id", ctx); err != nil {
				logs.Error(errors.Wrap(err, prefix))
				return
			}
		}

		if ctx.SetStatusCode(m.uc.Details(&thread, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&thread, ctx)
	}
}

func (m ThreadHandlerManager) Edit() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "thread delivery edit:"
			err    error
			thread models.Thread
		)
		if err = args.GetBodyInterface(&thread, ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			return
		}

		if thread.Id, err = args.PathInt("slug_or_id", ctx); err != nil {
			if thread.Slug, err = args.PathString("slug_or_id", ctx); err != nil {
				logs.Error(errors.Wrap(errors.New("Ошибка приведения к строке"), prefix))
				return
			}
		}

		if ctx.SetStatusCode(m.uc.Edit(&thread, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&thread, ctx)
	}
}

func (m ThreadHandlerManager) Posts() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "thread delivery posts:"
			err    error
			thread models.Thread
			posts  []*models.Post
		)
		if thread.Id, err = args.PathInt("slug_or_id", ctx); err != nil {
			if thread.Slug, err = args.PathString("slug_or_id", ctx); err != nil {
				logs.Error(errors.Wrap(errors.New("bad cast to string"), prefix))
				args.SetBodyError(err, ctx)
				return
			}
		}

		limit := args.QueryInt("limit", ctx)
		since := args.QueryInt("since", ctx)
		sort := args.QueryString("sort", ctx)
		desc := args.QueryBool("desc", ctx)

		posts = make([]*models.Post, 0, _const.BuffSize)

		if ctx.SetStatusCode(m.uc.Posts(&posts, &thread, &err, limit, since, sort, desc)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&posts, ctx)
	}
}

func (m ThreadHandlerManager) Vote() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "thread delivery vote:"
			err    error
			thread models.Thread
			vote   models.Vote
		)
		if thread.Id, err = args.PathInt("slug_or_id", ctx); err != nil {
			if thread.Slug, err = args.PathString("slug_or_id", ctx); err != nil {
				logs.Error(errors.Wrap(errors.New("bad cast to string"), prefix))
				args.SetBodyError(err, ctx)
				return
			}
		}

		if err = args.GetBodyInterface(&vote, ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if ctx.SetStatusCode(m.uc.Vote(&thread, &vote, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&thread, ctx)
	}
}
