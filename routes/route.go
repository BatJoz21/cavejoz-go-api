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
	authGroup.GET("/profile", getMyUserProfile)
	authGroup.PUT("/profile", updateUserProfile)
	authGroup.GET("/profile/:username", getUserProfile)
	authGroup.GET("/avatar/:uID", getUserAvatar)

	authGroup.POST("/friends", addFriend)
	authGroup.GET("/friends/pending", getPendingFriendList)
	authGroup.PUT("/friends/pending/:frId", acceptFriendRequest)
	authGroup.GET("/friends", getFriendsList)
	authGroup.DELETE("/friends/delete/:frId", deleteOrRejectFriendship)
	authGroup.POST("/block", blockAUser)
	authGroup.GET("/friends/posts/:uID", getAllPostsOfAUser)

	authGroup.POST("/posts", createPost)
	authGroup.GET("/posts/:postID", viewAPost)
	authGroup.PUT("/posts/:postID", editAPost)
	authGroup.DELETE("/posts/:postID", deleteAPost)
}
