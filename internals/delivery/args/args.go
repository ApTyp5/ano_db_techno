package args

import (
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"strconv"
)

func GetBodyInterface(v interface{}, c echo.Context) error {
	return c.Bind(v)
}

func PathInt(name string, ctx echo.Context) (int, error) {
	var (
		prefix = "извлечение целого числа из пути"
		value  int64
		err    error
	)

	if value, err = strconv.ParseInt(ctx.Param(name), 10, 0); err != nil {
		return 0, errors.Wrap(err, prefix)
	}

	return int(value), nil
}

func PathString(name string, ctx echo.Context) (string, error) {
	return ctx.Param(name), nil
}

func QueryInt(name string, ctx echo.Context) int {
	if ctx.QueryParams().Get(name) != "" {
		strInt := ctx.QueryParam(name)
		num, err := strconv.ParseInt(strInt, 10, 0)
		if err != nil {
			logs.Error(err)
		}

		return int(num)
	}

	return -1
}

func QueryString(name string, ctx echo.Context) string {
	return ctx.QueryParam(name)
}

func QueryBool(name string, ctx echo.Context) bool {
	if ctx.QueryParams().Get(name) != "" {
		str := QueryString(name, ctx)
		return str == "true"
	}
	return false
}
