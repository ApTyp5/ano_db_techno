package delivery

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	. "github.com/labstack/echo"
	"strconv"
)

func retError(err error) (int, interface{}) {
	er := &models.Error{Message: err.Error()}
	return 600, er
}

func PathNatural(c Context, name string) int {
	val := c.Param(name)
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return -1
	}

	return int(i)
}

func QueryNatural(c Context, name string) int {
	val := c.QueryParam(name)
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return -1
	}

	return int(i)
}

func QueryBool(c Context, name string) bool {
	return c.QueryParam(name) == "true"
}
