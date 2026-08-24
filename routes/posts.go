package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"path/filepath"
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
	content, err = utils.SavePostContent(file, context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save post content"})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload post"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Post uploaded"})
}

func getPostsForFeeds(context *gin.Context) {
	// Get friend's IDs
	friendIDs, err := models.GetFriendsID(context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch friends"})
		return
	}

	// Get offset for pagination
	offset := utils.GetOffsetForPagination(models.FeedsLimit, context)

	// Get posts
	posts, err := models.GetPostsForFeedByUIDs(*friendIDs, offset)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch feeds"})
		return
	}

	context.JSON(http.StatusOK, posts)
}

func getAllPostsOfAUser(context *gin.Context) {
	// Get user ID
	id, err := strconv.ParseInt(context.Param("uID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid uID"})
		return
	}

	// Get offset for pagination
	offset := utils.GetOffsetForPagination(models.PostsLimitPerPage, context)

	// Check friendship status
	visibility := ""
	if context.GetInt64("uID") != id {
		if !models.IsFriend(context.GetInt64("uID"), id) {
			visibility = "public"
		}
	}

	// Get posts
	posts, err := models.GetAllPostsofAUser(id, visibility, offset)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch posts"})
		return
	}

	context.JSON(http.StatusOK, posts)
}

func viewAPost(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid postID"})
		return
	}

	// Check post's visibility
	if !canViewPost(id, context.GetInt64("uID")) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorize"})
		return
	}

	// Get post data from database
	post, err := models.GetPostofAUser(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch post"})
		return
	}

	context.JSON(http.StatusOK, post)
}

func getPostContentImage(context *gin.Context) {
	// Get the filename
	filename := context.Param("filename")

	// Check filename
	if filename != filepath.Base(filename) {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid filename"})
		return
	}

	// Check if user can access this post image
	postID, err := models.GetPostIDByContentUrl(filename)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"message": "Image not found"})
			return
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch image"})
			return
		}
	}
	if !canViewPost(postID, context.GetInt64("uID")) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
		return
	}

	// Get content image url
	contentUrl := utils.GetImageContentPath(&filename)

	// Show the image
	context.File(contentUrl)
}

func getTotalUserPost(context *gin.Context) {
	// Get user ID
	uID, err := strconv.ParseInt(context.Param("uID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid uID"})
		return
	}

	// Get the total
	total, err := models.GetTotalPostByUID(uID, models.IsFriend(context.GetInt64("uID"), uID))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch total post"})
		return
	}

	context.JSON(http.StatusOK, total)
}

func canViewPost(postID, uID int64) bool {
	// Check post visibility
	vis, pUID, err := models.GetPostVisibility(postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		} else {
			return false
		}
	}

	// Check if user is able to view, like, and comment on the post
	if uID != pUID && *vis == "friends" && !models.IsFriend(uID, pUID) {
		return false
	}

	return true
}

func editAPost(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid postID"})
		return
	}

	// Get the post data
	post, err := models.GetPostForOperation(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch post"})
		return
	}

	// Handle post's content file upload
	file, err := context.FormFile("content")
	if err == nil {
		content, err := utils.SavePostContent(file, context)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save post content"})
			return
		}

		err = utils.RemoveImage(&post.ContentUrl, "content")
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to remove old post content"})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to edit post"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Post edited"})
}

func deleteAPost(context *gin.Context) {
	// Get post ID
	id, err := strconv.ParseInt(context.Param("postID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid postID"})
		return
	}

	// Get the post data
	post, err := models.GetPostForOperation(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch post"})
		return
	}

	// Delete post's like
	err = models.DeleteAllLikeByPostID(post.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete post's likes"})
		return
	}

	// Delete post's comment
	err = models.DeleteAllCommentByPostID(post.ID, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete post's comments"})
		return
	}

	// Delete post data
	err = models.DeletePost(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete post"})
		return
	}

	// Delete post's content
	err = utils.RemoveImage(&post.ContentUrl, "content")
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete post content"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}
