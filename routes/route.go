package routes

import (
	"github.com/BatJoz21/cavejoz-go-api/hub"
	"github.com/BatJoz21/cavejoz-go-api/middlewares"
	"github.com/gin-gonic/gin"
)

var appHub *hub.Hub

func RegisteredRoutes(server *gin.Engine, h *hub.Hub) {
	appHub = h
	if appHub == nil {
		panic("hub not initialized")
	}

	server.GET("/health", func(context *gin.Context) {
		context.JSON(200, gin.H{"status": "ok"})
	})

	// The unauthenticated endpoints each get their own rate-limit budget.
	server.POST("/register",
		middlewares.RateLimitIP(middlewares.RegisterMax, middlewares.RegisterWindow),
		registerNewUser)
	server.POST("/login",
		middlewares.RateLimitIP(middlewares.LoginIPMax, middlewares.LoginIPWindow),
		middlewares.ThrottleLoginByAccount(middlewares.LoginAccountMax, middlewares.LoginAccountWindow),
		login)
	server.POST("/refresh",
		middlewares.RateLimitIP(middlewares.SessionMax, middlewares.SessionWindow),
		refreshAccessToken)
	server.POST("/logout",
		middlewares.RateLimitIP(middlewares.SessionMax, middlewares.SessionWindow),
		logout)

	server.GET("/goapi/ws", WebSocketHandler(appHub))

	authGroup := server.Group("")
	authGroup.Use(middlewares.Authenticate)

	authGroup.GET("/ws-ticket", getNewWSTicket)

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

	authGroup.POST("/dm", createChatRoom)
	authGroup.GET("/dm", getConversations)
	authGroup.GET("/dmID", getConversationID)
	authGroup.GET("/dm/:cID", getConversation)
	authGroup.GET("/dm/:cID/message", getConversationMessage)
	authGroup.POST("/dm/:cID/message", sendMessage)
	authGroup.PUT("/dm/:cID/read", updateReadRecordOnConversation)

	authGroup.GET("/users/posts/:uID", getAllPostsOfAUser)
	authGroup.GET("/users/posts/:uID/total", getTotalUserPost)
}
