package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func toggleLike(context *gin.Context) {
	var like models.Like
	var err error
	like.PostID, err = strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	like.UserID = context.GetInt64("uID")

	isExists, id := models.CheckIfLikeExists(like.PostID, like.UserID)
	message := ""
	if isExists {
		err = models.DeleteLike(id)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		message = "Post unlike"
	} else {
		err = like.SaveLike()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		message = "Post liked"
	}

	context.JSON(http.StatusOK, gin.H{"message": message})
}

func getTotalLike(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	total := models.TotalLikeofAPost(id)
	context.JSON(http.StatusOK, total)
}
