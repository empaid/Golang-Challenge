package main

import (
	"log"
	"main/handlers"
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
	r.GET("/users", userHandler.GetUsers)
	r.POST("/users", userHandler.CreateUser)
	r.POST("/users/:id/messages", userHandler.CreateUserMessage)
	r.PATCH("/users/:id", userHandler.PatchUser)
	r.Run()
}
