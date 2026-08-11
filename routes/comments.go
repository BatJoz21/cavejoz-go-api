package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func createNewComment(context *gin.Context) {
	// Get user's input
	var dto models.NewCommentDTO
	err := context.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get post's ID
	postID, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Create struct
	comment := models.Comment{
		PostID:  postID,
		UserID:  context.GetInt64("uID"),
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

	// Get page number
	page, err := strconv.Atoi(context.DefaultQuery("page", "1"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
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

	// Search comment data in the database
	c, err := models.GetCommentByID(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Delete data from database
	err = c.Delete()
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
