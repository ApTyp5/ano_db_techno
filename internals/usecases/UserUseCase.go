package usecases

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/repositories"
	"github.com/jackc/pgx"
	"net/http"
)

type UserUseCase interface {
	Create(users []models.User, user *models.User) (int, interface{}) // /user/{nickname}/create
	Update(user *models.User) (int, interface{})                      // /user/{nickname}/profile
	Get(user *models.User) (int, interface{})                         // /user/{nickname}/profile
}

type RDBUserUseCase struct {
	us repositories.UserRepo
}

func CreateRDBUserUseCase(db *pgx.ConnPool) UserUseCase {
	return RDBUserUseCase{
		us: repositories.CreatePSQLUserRepo(db),
	}
}

func (uc RDBUserUseCase) Create(users []models.User, user *models.User) (int, interface{}) {
	if err := uc.us.SelectByNickNameOrEmail(&users, user); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	if len(users) > 0 {
		return http.StatusConflict, users
	}

	if err := uc.us.Insert(user); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusCreated, user
}

func (uc RDBUserUseCase) Update(user *models.User) (int, interface{}) {
	users := make([]models.User, 0)
	if err := uc.us.SelectByNickNameOrEmail(&users, user); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("user with such nick not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}

	if len(users) == 0 {
		return http.StatusNotFound, wrapStrError("user with such nick not found")
	}

	if len(users) > 1 {
		return http.StatusConflict, wrapStrError("your data conflicts with other users")
	}

	if err := uc.us.UpdateByNickname(user); err != nil {
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusOK, user
}

func (uc RDBUserUseCase) Get(user *models.User) (int, interface{}) {
	if err := uc.us.SelectByNickname(user); err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, wrapStrError("user with such nick not found")
		}
		return http.StatusInternalServerError, wrapError(err)
	}

	return http.StatusOK, user
}
