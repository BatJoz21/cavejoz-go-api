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
	authGroup.GET("/avatar/:filename", getUserAvatar)

	authGroup.POST("/friends", addFriend)
	authGroup.GET("/friends/:uID/total", getUserTotalFriend)
	authGroup.GET("/friends/pending", getPendingFriendList)
	authGroup.PUT("/friends/pending/:frId", acceptFriendRequest)
	authGroup.GET("/friends", getFriendsList)
	authGroup.GET("/friends/status/:targetUID", getFriendshipStatus)
	authGroup.DELETE("/friends/delete/:frId", deleteOrRejectFriendship)
	authGroup.GET("/block", getBlockedList)
	authGroup.POST("/block", blockAUser)

	authGroup.GET("/feeds", getPostsForFeeds)

	authGroup.POST("/posts", createPost)
	authGroup.GET("/posts/:postID", viewAPost)
	authGroup.GET("/content/image/:filename", getPostContentImage)
	authGroup.PUT("/posts/:postID", editAPost)
	authGroup.DELETE("/posts/:postID", deleteAPost)

	authGroup.POST("/posts/:postID/like", toggleLike)
	authGroup.GET("/posts/:postID/like", getTotalLike)

	authGroup.POST("/posts/:postID/comment", createNewComment)
	authGroup.GET("/posts/:postID/comment", getAllCommentOfAPost)
	authGroup.DELETE("/posts/:postID/comment/:commentID", deleteComment)

	authGroup.GET("/notifications", getAllNotification)
	authGroup.GET("/notifications/:limit", getNotificationWithLimit)
	authGroup.PUT("/notifications/markAllRead", markAllNotificationRead)
	authGroup.PUT("/notifications/:notifID", markReadNotification)

	authGroup.GET("/users/posts/:uID", getAllPostsOfAUser)
	authGroup.GET("/users/posts/:uID/total", getTotalUserPost)
}
