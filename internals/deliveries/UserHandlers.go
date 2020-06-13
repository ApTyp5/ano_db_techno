package deliveries

import (
	_const "github.com/ApTyp5/new_db_techno/const"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecases"
	"github.com/jackc/pgx"
	. "github.com/labstack/echo"
)

type UserHandlerManager struct {
	uc usecases.UserUseCase
}

func CreateUserHandlerManager(db *pgx.ConnPool) UserHandlerManager {
	return UserHandlerManager{uc: usecases.CreateRDBUserUseCase(db)}
}

func (m UserHandlerManager) Create() HandlerFunc {
	return func(c Context) error {
		var (
			err   error
			users = make([]models.User, 0, _const.BuffSize)
			user  = models.User{NickName: c.Param("nickname")}
		)

		if err = c.Bind(&user); err != nil {
			return c.JSON(retError(err))
		}

		return c.JSON(m.uc.Create(users, &user))
	}
}

func (m UserHandlerManager) Profile() HandlerFunc {
	return func(c Context) error {
		user := models.User{NickName: c.Param("nickname")}
		return c.JSON(m.uc.Get(&user))
	}
}

func (m UserHandlerManager) UpdateProfile() HandlerFunc {
	return func(c Context) error {
		user := models.User{NickName: c.Param("nickname")}

		if err := c.Bind(&user); err != nil {
			return c.JSON(retError(err))
		}
		return c.JSON(m.uc.Update(&user))
	}
}
