package delivery

import (
	_const "github.com/ApTyp5/new_db_techno/const"
	"github.com/ApTyp5/new_db_techno/internals/delivery/args"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecase"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	. "github.com/valyala/fasthttp"
)

type ForumHandlerManager struct {
	uc usecase.ForumUseCase
}

func CreateForumHandlerManager(db *pgx.ConnPool) ForumHandlerManager {
	return ForumHandlerManager{
		uc: usecase.CreateRDBForumUseCase(db),
	}
}

// /forum/create
func (m ForumHandlerManager) Create() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			err    error
			forum  models.Forum
			prefix = "forum handler create"
		)
		if err := args.GetBodyInterface(&forum, ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if ctx.SetStatusCode(m.uc.Create(&forum, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
		}

		args.SetBodyInterface(&forum, ctx)
	}
}

// /forum/{slug}/create
func (m ForumHandlerManager) CreateThread() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "forumHandler createThread"
			thread models.Thread
			err    error
		)
		if err = args.GetBodyInterface(&thread, ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if thread.Forum, err = args.PathString("slug", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if ctx.SetStatusCode(m.uc.CreateThread(&thread, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&thread, ctx)
	}
}

// /forum/{slug}/details
func (m ForumHandlerManager) Details() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "forum details handler"
			forum  models.Forum
			err    error
		)
		if forum.Slug, err = args.PathString("slug", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if ctx.SetStatusCode(m.uc.Details(&forum, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&forum, ctx)
	}
}

// /forum/{slug}/threads
func (m ForumHandlerManager) Threads() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix  = "forum threads handler"
			threads []*models.Thread
			forum   models.Forum
			err     error
		)
		if forum.Slug, err = args.PathString("slug", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		limit := args.QueryInt("limit", ctx)
		since := args.QueryString("since", ctx)
		desc := args.QueryBool("desc", ctx)
		threads = make([]*models.Thread, 0, _const.BuffSize)

		if ctx.SetStatusCode(m.uc.Threads(&threads, &forum, &err, limit, since, desc)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&threads, ctx)
	}
}

// /forum/{slug}/users
func (m ForumHandlerManager) Users() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "forum users handler"
			forum  models.Forum
			users  []*models.User
			err    error
		)
		if forum.Slug, err = args.PathString("slug", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		limit := args.QueryInt("limit", ctx)
		since := args.QueryString("since", ctx)
		desc := args.QueryBool("desc", ctx)
		users = make([]*models.User, 0, _const.BuffSize)

		if ctx.SetStatusCode(m.uc.Users(&users, &forum, &err, limit, since, desc)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&users, ctx)
	}
}
