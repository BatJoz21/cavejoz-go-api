package routes

import (
	"github.com/BatJoz21/cavejoz-go-api/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisteredRoutes(server *gin.Engine) {
	server.POST("/register", registerNewUser)
	server.POST("/login", login)
	server.POST("/refresh", refreshAccessToken)
	server.POST("/logout", logout)

	authGroup := server.Group("")
	authGroup.Use(middlewares.Authenticate)
	authGroup.GET("/profile", getUserProfile)
	authGroup.PUT("/profile", updateUserProfile)
}
