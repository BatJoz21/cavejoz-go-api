package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func createNewComment(context *gin.Context) {
	// Get post's ID and user's id
	postID, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}
	uID := context.GetInt64("uID")

	// Check if user is authorize to comment the post
	if !canViewPost(postID, uID) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
		return
	}

	// Get user's input
	var dto models.NewCommentDTO
	err = context.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload comment"})
		return
	}

	// Create notification
	userID, err := models.GetPostOwnerID(postID)
	if err == nil {
		NotifyandPush(userID, comment.UserID, comment.PostID, "comment")
	} else {
		log.Println("Comment uploaded, but failed to create notification")
	}

	context.JSON(http.StatusOK, gin.H{"message": "New comment uploaded"})
}

func getAllCommentOfAPost(context *gin.Context) {
	// Get post's ID
	postID, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Check if user is authorize to view post's comment
	if !canViewPost(postID, context.GetInt64("uID")) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
		return
	}

	// Get page number
	page, err := strconv.Atoi(context.DefaultQuery("page", "1"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}
	if page < 1 {
		page = 1
	}
	offset := models.COMMENT_LIMIT_PER_PAGE * (page - 1)

	// Get data from database
	comments, err := models.GetAllCommentByPostID(postID, offset)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch comments"})
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
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Delete data from database
	count, err := models.DeleteCommentByID(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete comment"})
		return
	}
	if count < 1 {
		context.JSON(http.StatusNotFound, gin.H{"message": "Comment not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
