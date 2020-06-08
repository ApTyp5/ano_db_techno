package usecase

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

type UserUseCase interface {
	Create(user *[]*models.User, err *error) int // /user/{nickname}/create
	Update(user *models.User, err *error) int    // /user/{nickname}/profile
	Get(user *models.User, err *error) int       // /user/{nickname}/profile
}

type RDBUserUseCase struct {
	us store.UserStore
}

func CreateRDBUserUseCase(db *pgx.ConnPool) UserUseCase {
	return RDBUserUseCase{
		us: store.CreatePSQLUserStore(db),
	}
}

func (uc RDBUserUseCase) Create(users *[]*models.User, err *error) int {
	prefix := "RDB users use case create"

	if *err = errors.Wrap(uc.us.Insert((*users)[0]), prefix); *err == nil {
		return 201
	}
	if *err = errors.Wrap(uc.us.SelectByNickNameOrEmail(users), prefix); *err == nil {
		*err = errors.New("Users with this nick or email exists")
		return 409
	}

	return 600
}

func (uc RDBUserUseCase) Update(user *models.User, err *error) int {
	prefix := "RDB user use case update"

	if *err = errors.Wrap(uc.us.UpdateByNickname(user), prefix); *err != nil {
		logs.Info("ERROR: ", (*err).Error())
		if *err = errors.Wrap(uc.us.SelectByNickname(user), prefix); *err != nil {
			return 404
		}
		return 409
	}

	return 200
}

func (uc RDBUserUseCase) Get(user *models.User, err *error) int {
	prefix := "RDB user use case get"

	if *err = errors.Wrap(uc.us.SelectByNickname(user), prefix); *err != nil {
		return 404
	}

	return 200
}
