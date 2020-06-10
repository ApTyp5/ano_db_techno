package usecase

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/store"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
)

type UserUseCase interface {
	Create(user []*models.User) (int, interface{}) // /user/{nickname}/create
	Update(user *models.User) (int, interface{})   // /user/{nickname}/profile
	Get(user *models.User) (int, interface{})      // /user/{nickname}/profile
}

type RDBUserUseCase struct {
	us store.UserStore
}

func CreateRDBUserUseCase(db *pgx.ConnPool) UserUseCase {
	return RDBUserUseCase{
		us: store.CreatePSQLUserStore(db),
	}
}

func (uc RDBUserUseCase) Create(users []*models.User) (int, interface{}) {
	prefix := "RDB users use case create"

	if err := errors.Wrap(uc.us.Insert(users[0]), prefix); err == nil {
		return 201, users[0]
	}

	if err := errors.Wrap(uc.us.SelectByNickNameOrEmail(&users), prefix); err == nil {
		return 409, &users
	}

	return unknownError()
}

func (uc RDBUserUseCase) Update(user *models.User) (int, interface{}) {
	prefix := "RDB user use case update"

	if err := errors.Wrap(uc.us.UpdateByNickname(user), prefix); err != nil {
		if err := errors.Wrap(uc.us.SelectByNickname(user), prefix); err != nil {
			return http.StatusNotFound, wrapStrError("user with such nick not found")
		}
		return http.StatusConflict, wrapStrError("your data conflicts with existing users")
	}

	return http.StatusOK, user
}

func (uc RDBUserUseCase) Get(user *models.User) (int, interface{}) {
	prefix := "RDB user use case get"

	if err := errors.Wrap(uc.us.SelectByNickname(user), prefix); err != nil {
		return http.StatusNotFound, wrapStrError("user with such nick not found")
	}

	return http.StatusOK, user
}
