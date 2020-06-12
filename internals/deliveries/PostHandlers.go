package deliveries

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecases"
	"github.com/jackc/pgx"
	. "github.com/labstack/echo"
	"strings"
)

type PostHandlerManager struct {
	uc usecases.PostUseCase
}

func CreatePostHandlerManager(db *pgx.ConnPool) PostHandlerManager {
	return PostHandlerManager{uc: usecases.CreateRDBPostUseCase(db)}
}

// /post/{id}/details
func (m PostHandlerManager) Details() HandlerFunc {
	return func(c Context) error {
		postFull := models.PostFull{Post: &models.Post{}}
		postFull.Post.Id = PathNatural(c, "id")
		related := c.QueryParam("related")

		return c.JSON(m.uc.Details(&postFull, strings.Split(related, ",")))
	}
}

// /post/{id}/details
func (m PostHandlerManager) Edit() HandlerFunc {
	return func(c Context) error {
		post := models.Post{Id: PathNatural(c, "id")}
		if err := c.Bind(&post); err != nil {
			return c.JSON(retError(err))
		}
		return c.JSON(m.uc.Edit(&post))
	}
}
