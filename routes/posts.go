package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/BatJoz21/cavejoz-go-api/utils"
	"github.com/gin-gonic/gin"
)

func createPost(context *gin.Context) {
	// Saving post's content file
	var content *string
	file, err := context.FormFile("content")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "No content"})
		return
	}
	content, err = utils.SavePostContent(file, context.GetInt64("uID"), context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Create post struct
	post := models.Post{
		UserID:     context.GetInt64("uID"),
		Caption:    context.PostForm("caption"),
		ContentUrl: *content,
		Visibility: context.PostForm("visibility"),
	}

	// Save post in the database
	if err := post.Save(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Post uploaded"})
}

func getAllPostsOfAUser(context *gin.Context) {
	// Get user ID
	id, err := strconv.ParseInt(context.Param("uID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Check friendship status
	visibility := ""
	if !models.IsFriend(context.GetInt64("uID"), id) {
		visibility = "public"
	}

	// Get posts
	posts, err := models.GetAllPostsofAUser(id, visibility)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, posts)
}

func viewAPost(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Check post's visibility
	vis, postUID, err := models.GetPostVisibility(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Check if user is owner or not
	if context.GetInt64("uID") != postUID {
		// Check if user is friend with post owner
		if *vis == "friends" {
			if !models.IsFriend(context.GetInt64("uID"), postUID) {
				context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorize"})
				return
			}
		}
	}

	// Get post data from database
	post, err := models.GetPostofAUser(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, post)
}

func editAPost(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get the post data
	post, err := models.GetPostForOperation(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Handle post's content file upload
	file, err := context.FormFile("content")
	if err == nil {
		content, err := utils.SavePostContent(file, context.GetInt64("uID"), context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		err = utils.RemoveImage(&post.ContentUrl)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		post.ContentUrl = *content
	}

	// Assign new value
	post.Caption = context.PostForm("caption")
	post.Visibility = context.PostForm("visibility")

	// Update the data on the database
	err = post.EditPost()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Post edited"})
}

func deleteAPost(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get the post data
	post, err := models.GetPostForOperation(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Delete post's content
	err = utils.RemoveImage(&post.ContentUrl)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Delete post data
	err = models.DeletePost(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}
