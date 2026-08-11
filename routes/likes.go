package routes

import (
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
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	like.UserID = context.GetInt64("uID")

	// Check if user has liked the post
	isExists, id := models.CheckIfLikeExists(like.PostID, like.UserID)
	message := ""
	isLiked := true

	if isExists {
		// Delete like data from database (unlike)
		err = models.DeleteLike(id)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		message = "Post unlike"
		isLiked = false
	} else {
		// Save like data to database (like)
		err = like.SaveLike()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		message = "Post liked"

		// Like notification
		userID, err := models.GetPostOwnerID(like.PostID)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		NotifyandPush(userID, like.UserID, like.PostID, "like")
	}

	// Get current total like on the post
	total := models.TotalLikeofAPost(like.PostID)

	context.JSON(http.StatusOK, gin.H{"message": message, "liked": isLiked, "count": total})
}

func getTotalLike(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get the total
	total := models.TotalLikeofAPost(id)
	context.JSON(http.StatusOK, total)
}
