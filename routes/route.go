package routes

import "github.com/gin-gonic/gin"

func RegisteredRoutes(server *gin.Engine) {
	server.POST("/register", registerNewUser)
	server.POST("/login", login)
	server.POST("/refresh", refreshAccessToken)
	server.POST("/logout", logout)
}
