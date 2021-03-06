package deliveries

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecases"
	"github.com/jackc/pgx"
	. "github.com/labstack/echo"
)

type ForumHandlerManager struct {
	uc usecases.ForumUseCase
}

func CreateForumHandlerManager(db *pgx.ConnPool) ForumHandlerManager {
	return ForumHandlerManager{
		uc: usecases.CreateRDBForumUseCase(db),
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
		return c.JSON(m.uc.Details(c.Param("slug")))
	}
}

// /forum/{slug}/threads
func (m ForumHandlerManager) Threads() HandlerFunc {
	return func(c Context) error {
		slug := c.Param("slug")
		limit := QueryNatural(c, "limit")
		since := c.QueryParam("since")
		desc := QueryBool(c, "desc")

		return c.JSON(m.uc.Threads(slug, limit, since, desc))
	}
}

// /forum/{slug}/users
func (m ForumHandlerManager) Users() HandlerFunc {
	return func(c Context) error {
		slug := c.Param("slug")
		limit := QueryNatural(c, "limit")
		since := c.QueryParam("since")
		desc := QueryBool(c, "desc")

		return c.JSON(m.uc.Users(slug, limit, since, desc))
	}
}
