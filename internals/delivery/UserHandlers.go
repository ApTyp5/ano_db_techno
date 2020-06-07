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

type UserHandlerManager struct {
	uc usecase.UserUseCase
}

func CreateUserHandlerManager(db *pgx.ConnPool) UserHandlerManager {
	return UserHandlerManager{uc: usecase.CreateRDBUserUseCase(db)}
}

func (m UserHandlerManager) Create() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "user delivery create"
			err    error
			users  []*models.User
		)
		users = make([]*models.User, 0, _const.BuffSize)
		users = append(users, &models.User{})

		if err = args.GetBodyInterface(&users[0], ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			return
		}

		if users[0].NickName, err = args.PathString("nickname", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			return
		}

		if ctx.SetStatusCode(m.uc.Create(&users, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyInterface(&users, ctx)
			return
		}

		args.SetBodyInterface(&users[0], ctx)
	}
}

func (m UserHandlerManager) Profile() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "user delivery profile"
			err    error
			user   models.User
		)
		if user.NickName, err = args.PathString("nickname", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if ctx.SetStatusCode(m.uc.Get(&user, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&user, ctx)
	}
}

func (m UserHandlerManager) UpdateProfile() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			prefix = "user delivery updateProfile"
			err    error
			user   models.User
		)
		if err = args.GetBodyInterface(&user, ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if user.NickName, err = args.PathString("nickname", ctx); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		if ctx.SetStatusCode(m.uc.Update(&user, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&user, ctx)
	}
}
