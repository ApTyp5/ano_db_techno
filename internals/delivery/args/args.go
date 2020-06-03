package args

import (
	"encoding/json"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"strconv"
)

func GetBodyInterface(v interface{}, ctx *fasthttp.RequestCtx) error {
	return errors.Wrap(json.Unmarshal(ctx.PostBody(), v), "getBodyInterface")
}

func SetBodyInterface(v interface{}, ctx *fasthttp.RequestCtx) {
	bytes, err := json.Marshal(v)
	ctx.SetBody(bytes)
	if err != nil {
		logs.Error(err)
	}
}

func SetBodyError(err error, ctx *fasthttp.RequestCtx) {
	er := models.Error{
		Message: err.Error(),
	}

	SetBodyInterface(&er, ctx)
}

func PathInt(name string, ctx *fasthttp.RequestCtx) (int, error) {
	var (
		prefix = "извлечение целого числа из пути"
		value  int64
		err    error
	)

	if value, err = strconv.ParseInt(ctx.UserValue(name).(string), 10, 0); err != nil {
		return 0, errors.Wrap(err, prefix)
	}

	return int(value), nil
}

func PathString(name string, ctx *fasthttp.RequestCtx) (string, error) {
	prefix := "извлечение строки из пути"

	value, ok := ctx.UserValue(name).(string)
	if !ok {
		return "", errors.Wrap(
			errors.New("ошибка приведения занчения к string"),
			prefix,
		)
	}

	return value, nil
}

func QueryInt(name string, ctx *fasthttp.RequestCtx) int {
	if ctx.QueryArgs().Has(name) {
		strInt := string(ctx.QueryArgs().Peek(name))

		num, err := strconv.ParseInt(strInt, 10, 0)
		if err != nil {
			logs.Error(err)
		}

		return int(num)
	}

	return -1
}

func QueryString(name string, ctx *fasthttp.RequestCtx) string {
	if ctx.QueryArgs().Has(name) {
		return string(ctx.QueryArgs().Peek(name))
	}
	return ""
}

func QueryBool(name string, ctx *fasthttp.RequestCtx) bool {
	if ctx.QueryArgs().Has(name) {
		str := QueryString(name, ctx)
		return str == "true"
	}
	return false
}

func QueryStringSlice(ctx *fasthttp.RequestCtx) []string {
	result := make([]string, 0, 3)
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		result = append(result, string(value))
	})

	return result
}
