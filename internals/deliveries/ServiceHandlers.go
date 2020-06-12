package deliveries

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecases"
	"github.com/jackc/pgx"
	. "github.com/labstack/echo"
)

type ServiceHandlerManager struct {
	uc usecases.ServiceUseCase
}

func CreateServiceHandlerManager(db *pgx.ConnPool) ServiceHandlerManager {
	return ServiceHandlerManager{uc: usecases.CreateRDBServiceUseCase(db)}
}

func (hm ServiceHandlerManager) Clear() HandlerFunc {
	return func(c Context) error {
		return c.JSON(hm.uc.Clear())
	}
}

func (hm ServiceHandlerManager) Status() HandlerFunc {
	return func(c Context) error {
		status := models.Status{}
		return c.JSON(hm.uc.Status(&status))
	}
}
