package usecase

import (
	"errors"
	"github.com/ApTyp5/new_db_techno/internals/models"
)

func wrapError(err error) interface{} {
	return &models.Error{Message: err.Error()}
}

func wrapStrError(str string) interface{} {
	return &models.Error{Message: str}
}

func unknownError() (int, interface{}) {
	return 600, errors.New("Unknown error")
}
