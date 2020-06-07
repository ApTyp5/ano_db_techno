package usecase

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
)

type ServiceUseCase interface {
	Clear(err *error) int
	Status(serverStatus *models.Status, err *error) int
}

type RDBServiceUseCase struct {
	ss store.ServiceStore
}

func CreateRDBServiceUseCase(db *sql.DB) ServiceUseCase {
	return RDBServiceUseCase{
		ss: store.CreatePSQLServiceStore(db),
	}
}

func (uc RDBServiceUseCase) Clear(err *error) int {
	if *err = uc.ss.Clear(); *err != nil {
		logs.Info("service delivery clear", errors.Wrap(*err, "unexpected useCase error"))
		return 600
	}
	return 200
}

func (uc RDBServiceUseCase) Status(serverStatus *models.Status, err *error) int {
	if *err = errors.Wrap(uc.ss.Status(serverStatus), "RDB ServiceUseCase Status"); *err != nil {
		logs.Error(*err)
		return 600
	}

	return 200
}
