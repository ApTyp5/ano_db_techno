package main

import (
	"fmt"
	"github.com/ApTyp5/new_db_techno/database"
	"github.com/ApTyp5/new_db_techno/internals/deliveries"
	_ "github.com/jackc/pgx"
	"github.com/labstack/echo"
	"time"
)

func main() {
	e := echo.New()
	e.Use(Logs)
	group := e.Group("/api")

	connStr := "user=docker database=docker host=0.0.0.0 port=5432 password=docker sslmode=disable"

	db := database.Connect(connStr, 100) // panic
	defer db.Close()                     // panic
	defer func() { database.TruncTables(db) }()

	forumHandlers := deliveries.CreateForumHandlerManager(db)
	postHandlers := deliveries.CreatePostHandlerManager(db)
	threadHandlers := deliveries.CreateThreadHandlerManager(db)
	userHandlers := deliveries.CreateUserHandlerManager(db)
	serviceHandlers := deliveries.CreateServiceHandlerManager(db)

	{ // forum handlers
		forumRouter := group.Group("/forum")
		forumRouter.POST("/create", forumHandlers.Create())
		forumRouter.POST("/:forum/create", forumHandlers.CreateThread())
		forumRouter.GET("/:slug/details", forumHandlers.Details())
		forumRouter.GET("/:slug/threads", forumHandlers.Threads())
		forumRouter.GET("/:slug/users", forumHandlers.Users())
	}
	{ // post handlers
		postRouter := group.Group("/post")
		postRouter.GET("/:id/details", postHandlers.Details())
		postRouter.POST("/:id/details", postHandlers.Edit())
	}
	{ // service handlers
		serviceRouter := group.Group("/service")
		serviceRouter.POST("/clear", serviceHandlers.Clear())
		serviceRouter.GET("/status", serviceHandlers.Status())
	}
	{ // thread handlers
		threadRouter := group.Group("/thread")
		threadRouter.POST("/:slug_or_id/create", threadHandlers.AddPosts())
		threadRouter.GET("/:slug_or_id/details", threadHandlers.Details())
		threadRouter.POST("/:slug_or_id/details", threadHandlers.Edit())
		threadRouter.GET("/:slug_or_id/posts", threadHandlers.Posts())
		threadRouter.POST("/:slug_or_id/vote", threadHandlers.Vote())
	}
	{ // user handlers
		userRouter := group.Group("/user")
		userRouter.POST("/:nickname/create", userHandlers.Create())
		userRouter.GET("/:nickname/profile", userHandlers.Profile())
		userRouter.POST("/:nickname/profile", userHandlers.UpdateProfile())
	}

	e.Logger.Fatal(e.Start(":5000"))
}

func Logs(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var err error
		if ctx.Request().Method == "GET" {
			start := time.Now()
			err = next(ctx)
			respTime := time.Since(start)
			if respTime.Milliseconds() >= 400 {
				fmt.Println("\n\nMILLI SEC:", respTime.Milliseconds(), "\n PATH:", ctx.Request().URL.Path, "\n METHOD:", ctx.Request().Method)
				fmt.Println(ctx.QueryParams())
			}
		} else {
			err = next(ctx)
		}
		return err
	}
}
