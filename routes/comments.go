package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func createNewComment(context *gin.Context) {
	// Get post's ID and user's id
	postID, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	uID := context.GetInt64("uID")

	// Check if user is authorize to comment the post
	if !canViewPost(postID, uID) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorize access"})
		return
	}

	// Get user's input
	var dto models.NewCommentDTO
	err = context.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Create struct
	comment := models.Comment{
		PostID:  postID,
		UserID:  uID,
		Content: dto.Content,
	}

	// Insert data into database
	err = comment.Save()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Create notification
	userID, err := models.GetPostOwnerID(postID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	NotifyandPush(userID, comment.UserID, comment.PostID, "comment")

	context.JSON(http.StatusOK, gin.H{"message": "New comment uploaded"})
}

func getAllCommentOfAPost(context *gin.Context) {
	// Get post's ID
	postID, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Check if user is authorize to view post's comment
	if !canViewPost(postID, context.GetInt64("uID")) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorize access"})
		return
	}

	// Get page number
	page, err := strconv.Atoi(context.DefaultQuery("page", "1"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if page < 1 {
		page = 1
	}
	offset := models.COMMENT_LIMIT_PER_PAGE * (page - 1)

	// Get data from database
	comments, err := models.GetAllCommentByPostID(postID, offset)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get total comment on a post
	total := models.GetTotalCommentByPostID(postID)

	context.JSON(http.StatusOK, gin.H{"comments": comments, "total": total})
}

func deleteComment(context *gin.Context) {
	// Get comment's ID
	id, err := strconv.ParseInt(context.Param("commentID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Delete data from database
	count, err := models.DeleteCommentByID(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if count < 1 {
		context.JSON(http.StatusNotFound, gin.H{"message": "No rows affected"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
