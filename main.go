package main

import (
	"log"

	"github.com/BatJoz21/cavejoz-go-api/databases"
	"github.com/BatJoz21/cavejoz-go-api/hub"
	"github.com/BatJoz21/cavejoz-go-api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	databases.InitDB()

	server := gin.Default()

	h := hub.NewHub()

	routes.RegisteredRoutes(server, h)

	server.Run(":8080")
}
