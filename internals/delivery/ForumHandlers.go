package delivery

import (
	_const "github.com/ApTyp5/new_db_techno/const"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecase"
	"github.com/jackc/pgx"
	. "github.com/labstack/echo"
)

type ForumHandlerManager struct {
	uc usecase.ForumUseCase
}

func CreateForumHandlerManager(db *pgx.ConnPool) ForumHandlerManager {
	return ForumHandlerManager{
		uc: usecase.CreateRDBForumUseCase(db),
	}
}

// /forum/create
func (m ForumHandlerManager) Create() HandlerFunc {
	return func(c Context) error {
		forum := models.Forum{}

		if err := c.Bind(&forum); err != nil {
			return c.JSON(retError(err))
		}

		return c.JSON(m.uc.Create(&forum))
	}
}

// /forum/{slug}/create
func (m ForumHandlerManager) CreateThread() HandlerFunc {
	return func(c Context) error {
		thread := models.Thread{}

		if err := c.Bind(&thread); err != nil {
			return c.JSON(retError(err))
		}

		return c.JSON(m.uc.CreateThread(&thread))
	}
}

// /forum/{slug}/details
func (m ForumHandlerManager) Details() HandlerFunc {
	return func(c Context) error {
		forum := models.Forum{Slug: c.Param("slug")}
		return c.JSON(m.uc.Details(&forum))
	}
}

// /forum/{slug}/threads
func (m ForumHandlerManager) Threads() HandlerFunc {
	return func(c Context) error {
		threads := make([]*models.Thread, 0, _const.BuffSize)
		forum := models.Forum{Slug: c.Param("slug")}
		limit := QueryNatural(c, "limit")
		since := c.QueryParam("since")
		desc := QueryBool(c, "desc")

		return c.JSON(m.uc.Threads(&threads, &forum, limit, since, desc))
	}
}

// /forum/{slug}/users
func (m ForumHandlerManager) Users() HandlerFunc {
	return func(c Context) error {
		forum := models.Forum{Slug: c.Param("slug")}
		users := make([]*models.User, 0, _const.BuffSize)
		limit := QueryNatural(c, "limit")
		since := c.QueryParam("since")
		desc := QueryBool(c, "desc")

		return c.JSON(m.uc.Users(&users, &forum, limit, since, desc))
	}
}
