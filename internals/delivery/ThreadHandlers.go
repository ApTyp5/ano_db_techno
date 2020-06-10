package delivery

import (
	_const "github.com/ApTyp5/new_db_techno/const"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/internals/usecase"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/jackc/pgx"
	. "github.com/labstack/echo"
)

type ThreadHandlerManager struct {
	uc usecase.ThreadUseCase
}

func CreateThreadHandlerManager(db *pgx.ConnPool) ThreadHandlerManager {
	return ThreadHandlerManager{
		uc: usecase.CreateRDBThreadUseCase(db),
	}
}

func (m ThreadHandlerManager) AddPosts() HandlerFunc {
	return func(c Context) error {
		var posts []models.Post
		thread := models.Thread{
			Id:   PathNatural(c, "slug_or_id"),
			Slug: c.Param("slug_or_id"),
		}

		if err := c.Bind(&posts); err != nil {
			return c.JSON(retError(err))
		}

		return c.JSON(m.uc.AddPosts(&thread, &posts))
	}
}

func (m ThreadHandlerManager) Details() HandlerFunc {
	return func(c Context) error {
		thread := models.Thread{
			Id:   PathNatural(c, "slug_or_id"),
			Slug: c.Param("slug_or_id"),
		}
		return c.JSON(m.uc.Details(&thread))
	}
}

func (m ThreadHandlerManager) Edit() HandlerFunc {
	return func(c Context) error {
		thread := models.Thread{
			Id:   PathNatural(c, "slug_or_id"),
			Slug: c.Param("slug_or_id"),
		}

		if err := c.Bind(&thread); err != nil {
			return c.JSON(retError(err))
		}

		return c.JSON(m.uc.Edit(&thread))
	}
}

func (m ThreadHandlerManager) Posts() HandlerFunc {
	return func(c Context) error {
		thread := models.Thread{
			Id:   PathNatural(c, "slug_or_id"),
			Slug: c.Param("slug_or_id"),
		}

		posts := make([]*models.Post, 0, _const.BuffSize)

		limit := QueryNatural(c, "limit")
		since := QueryNatural(c, "since")
		sort := c.QueryParam("sort")
		desc := QueryBool(c, "desc")

		return c.JSON(m.uc.Posts(&posts, &thread, limit, since, sort, desc))
	}
}

func (m ThreadHandlerManager) Vote() HandlerFunc {
	return func(c Context) error {
		thread := models.Thread{
			Id:   PathNatural(c, "slug_or_id"),
			Slug: c.Param("slug_or_id"),
		}

		logs.Info("id: ", thread.Id)
		vote := models.Vote{}

		if err := c.Bind(&vote); err != nil {
			return c.JSON(retError(err))
		}
		logs.Info("id: ", thread.Id)
		return c.JSON(m.uc.Vote(&thread, &vote))
	}
}
