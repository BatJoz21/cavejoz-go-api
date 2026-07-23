package main

import (
	"log"

	"github.com/BatJoz21/cavejoz-go-api/databases"
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

	routes.RegisteredRoutes(server)

	server.Run(":8080")
}