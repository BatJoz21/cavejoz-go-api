package main

import (
	"log"
	"net/http"
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
	"github.com/BatJoz21/cavejoz-go-api/hub"
	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/BatJoz21/cavejoz-go-api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	databases.InitDB()

	models.StartWSTicketSweeper(10 * time.Minute)

	server := gin.Default()

	if err := server.SetTrustedProxies(nil); err != nil {
		log.Fatal("failed to set trusted proxies: ", err)
	}

	h := hub.NewHub()
	routes.RegisteredRoutes(server, h)

	srv := &http.Server{
		Addr: ":8080", Handler: server,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	srv.ListenAndServe()
}
