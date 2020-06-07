package delivery

import (
	"github.com/ApTyp5/new_db_techno/internals/delivery/args"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecase"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	. "github.com/valyala/fasthttp"
)

type ServiceHandlerManager struct {
	uc usecase.ServiceUseCase
}

func CreateServiceHandlerManager(db *pgx.ConnPool) ServiceHandlerManager {
	return ServiceHandlerManager{uc: usecase.CreateRDBServiceUseCase(db)}
}

func (hm ServiceHandlerManager) Clear() RequestHandler {
	return func(ctx *RequestCtx) {
		var (
			err    error
			prefix = "service handler clear"
		)
		if ctx.SetStatusCode(hm.uc.Clear(&err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			return
		}
	}
}

func (hm ServiceHandlerManager) Status() RequestHandler {

	return func(ctx *RequestCtx) {
		var (
			status models.Status
			prefix = "service handler status"
			err    error
		)
		if ctx.SetStatusCode(hm.uc.Status(&status, &err)); err != nil {
			logs.Error(errors.Wrap(err, prefix))
			args.SetBodyError(err, ctx)
			return
		}

		args.SetBodyInterface(&status, ctx)
	}
}
