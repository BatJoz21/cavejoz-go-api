package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func toggleLike(context *gin.Context) {
	// Initialize struct and get parameter value
	var like models.Like
	var err error
	like.PostID, err = strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid postID"})
		return
	}
	like.UserID = context.GetInt64("uID")

	// Check if user is authorize to like the post
	if !canViewPost(like.PostID, like.UserID) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
		return
	}

	// Check if user has liked the post
	isExists, id := models.CheckIfLikeExists(like.PostID, like.UserID)
	message := ""
	isLiked := true

	if isExists {
		// Delete like data from database (unlike)
		err = models.DeleteLike(id)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to unlike the post"})
			return
		}
		message = "Post unlike"
		isLiked = false
	} else {
		// Save like data to database (like)
		err = like.SaveLike()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to like the post"})
			return
		}
		message = "Post liked"

		// Like notification
		userID, err := models.GetPostOwnerID(like.PostID)
		if err == nil {
			NotifyandPush(userID, like.UserID, like.PostID, "like")
		} else {
			log.Println("Post liked, but failed to create notification")
		}
	}

	// Get current total like on the post
	total := models.TotalLikeofAPost(like.PostID)

	context.JSON(http.StatusOK, gin.H{"message": message, "liked": isLiked, "count": total})
}

func getTotalLike(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid postID"})
		return
	}

	// Check post availability
	if !canViewPost(id, context.GetInt64("uID")) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
		return
	}

	// Get the total
	total := models.TotalLikeofAPost(id)
	context.JSON(http.StatusOK, total)
}
