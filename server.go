package main

import (
	"github.com/ApTyp5/new_db_techno/database"
	"github.com/ApTyp5/new_db_techno/internals/delivery"
	"github.com/ApTyp5/new_db_techno/logs"
	mv "github.com/ApTyp5/new_db_techno/middleware"
	fasthttpRouter "github.com/fasthttp/router"

	_ "github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

func main() {
	init := fasthttpRouter.New()
	router := init.Group("/api")
	connStr := "user=docker database=docker host=0.0.0.0 port=5432 password=docker sslmode=disable"

	db := database.Connect(connStr, 70) // panic
	defer db.Close()                    // panic
	defer func() { database.DropTables(db) }()

	forumHandlers := delivery.CreateForumHandlerManager(db)
	postHandlers := delivery.CreatePostHandlerManager(db)
	threadHandlers := delivery.CreateThreadHandlerManager(db)
	userHandlers := delivery.CreateUserHandlerManager(db)
	serviceHandlers := delivery.CreateServiceHandlerManager(db)
	{ // forum handlers
		forumRouter := router.Group("/forum")
		forumRouter.POST("/create",
			logs.AccessLog(mv.ContentTypeAppJson(forumHandlers.Create())))
		forumRouter.POST("/{slug}/create",
			logs.AccessLog(mv.ContentTypeAppJson(forumHandlers.CreateThread())))
		forumRouter.GET("/{slug}/details",
			logs.AccessLog(mv.ContentTypeAppJson(forumHandlers.Details())))
		forumRouter.GET("/{slug}/threads",
			logs.AccessLog(mv.ContentTypeAppJson(forumHandlers.Threads())))
		forumRouter.GET("/{slug}/users",
			logs.AccessLog(mv.ContentTypeAppJson(forumHandlers.Users())))
	}
	{ // post handlers
		postRouter := router.Group("/post")
		postRouter.GET("/{id}/details",
			logs.AccessLog(mv.ContentTypeAppJson(postHandlers.Details())))
		postRouter.POST("/{id}/details",
			logs.AccessLog(mv.ContentTypeAppJson(postHandlers.Edit())))
	}
	{ // service handlers
		serviceRouter := router.Group("/service")
		serviceRouter.POST("/clear",
			logs.AccessLog(mv.ContentTypeAppJson(serviceHandlers.Clear())))
		serviceRouter.GET("/status",
			logs.AccessLog(mv.ContentTypeAppJson(serviceHandlers.Status())))
	}
	{ // thread handlers
		threadRouter := router.Group("/thread")
		threadRouter.POST("/{slug_or_id}/create",
			logs.AccessLog(mv.ContentTypeAppJson(threadHandlers.AddPosts())))
		threadRouter.GET("/{slug_or_id}/details",
			logs.AccessLog(mv.ContentTypeAppJson(threadHandlers.Details())))
		threadRouter.POST("/{slug_or_id}/details",
			logs.AccessLog(mv.ContentTypeAppJson(threadHandlers.Edit())))
		threadRouter.GET("/{slug_or_id}/posts",
			logs.AccessLog(mv.ContentTypeAppJson(threadHandlers.Posts())))
		threadRouter.POST("/{slug_or_id}/vote",
			logs.AccessLog(mv.ContentTypeAppJson(threadHandlers.Vote())))
	}
	{ // user handlers
		userRouter := router.Group("/user")
		userRouter.POST("/{nickname}/create",
			logs.AccessLog(mv.ContentTypeAppJson(userHandlers.Create())))
		userRouter.GET("/{nickname}/profile",
			logs.AccessLog(mv.ContentTypeAppJson(userHandlers.Profile())))
		userRouter.POST("/{nickname}/profile",
			logs.AccessLog(mv.ContentTypeAppJson(userHandlers.UpdateProfile())))
	}

	logs.Info("server started on 5000")
	logs.Fatal(fasthttp.ListenAndServe(":5000", init.Handler))
}
