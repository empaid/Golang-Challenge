package main

import (
	"log"
	"main/handlers"
	"main/middleware"
	"main/queries"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Server is starting...")

	r := gin.Default()

	dbConn := queries.GetConnection()

	userHandler := &handlers.UserHandler{
		UserDB: queries.NewUserDB(dbConn),
	}

	users := r.Group("/users")
	users.POST("", userHandler.CreateUser)
	users.Use(middleware.JWTAuth())
	{
		users.GET("", userHandler.GetUsers)
		users.POST("/:id/messages", userHandler.CreateUserMessage)
		users.PATCH("/:id", userHandler.PatchUser)

	}

	r.POST("/login", userHandler.Login)
	r.Run()
}
