package usecases

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/repositories"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
)

type ServiceUseCase interface {
	Clear() (int, interface{})
	Status(serverStatus *models.Status) (int, interface{})
}

type RDBServiceUseCase struct {
	ss repositories.ServiceStore
}

func CreateRDBServiceUseCase(db *pgx.ConnPool) ServiceUseCase {
	return RDBServiceUseCase{
		ss: repositories.CreatePSQLServiceStore(db),
	}
}

func (uc RDBServiceUseCase) Clear() (int, interface{}) {
	if err := uc.ss.Clear(); err != nil {
		return unknownError()
	}
	return http.StatusOK, nil
}

func (uc RDBServiceUseCase) Status(serverStatus *models.Status) (int, interface{}) {
	if err := errors.Wrap(uc.ss.Status(serverStatus), "RDB ServiceUseCase Status"); err != nil {
		return unknownError()
	}

	return http.StatusOK, serverStatus
}
