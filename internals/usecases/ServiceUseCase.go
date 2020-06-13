package usecases

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/repositories"
	"github.com/jackc/pgx"
	"net/http"
)

type ServiceUseCase interface {
	Clear() (int, interface{})
	Status(serverStatus *models.Status) (int, interface{})
}

type RDBServiceUseCase struct {
	ss repositories.ServiceRepo
}

func CreateRDBServiceUseCase(db *pgx.ConnPool) ServiceUseCase {
	return RDBServiceUseCase{
		ss: repositories.CreatePSQLServiceRepo(db),
	}
}

func (uc RDBServiceUseCase) Clear() (int, interface{}) {
	if err := uc.ss.Clear(); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}
	return http.StatusOK, nil
}

func (uc RDBServiceUseCase) Status(serverStatus *models.Status) (int, interface{}) {
	if err := uc.ss.Status(serverStatus); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}
	return http.StatusOK, serverStatus
}
