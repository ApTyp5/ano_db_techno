package usecase

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
)

type ServiceUseCase interface {
	Clear() (int, interface{})
	Status(serverStatus *models.Status) (int, interface{})
}

type RDBServiceUseCase struct {
	ss store.ServiceStore
}

func CreateRDBServiceUseCase(db *pgx.ConnPool) ServiceUseCase {
	return RDBServiceUseCase{
		ss: store.CreatePSQLServiceStore(db),
	}
}

func (uc RDBServiceUseCase) Clear() (int, interface{}) {
	if err := uc.ss.Clear(); err != nil {
		logs.Info("service delivery clear", errors.Wrap(err, "unexpected useCase error"))
		return unknownError()
	}
	return http.StatusOK, nil
}

func (uc RDBServiceUseCase) Status(serverStatus *models.Status) (int, interface{}) {
	if err := errors.Wrap(uc.ss.Status(serverStatus), "RDB ServiceUseCase Status"); err != nil {
		logs.Error(err)
		return unknownError()
	}

	return http.StatusOK, serverStatus
}
