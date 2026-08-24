package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/BatJoz21/cavejoz-go-api/utils"
	"github.com/gin-gonic/gin"
)

func getMyUserProfile(context *gin.Context) {
	// Get profile data using getUserDataForOperation function
	user := getMyUserData(context)
	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	context.JSON(http.StatusOK, user)
}

func getUserProfile(context *gin.Context) {
	// Get user's username from stored data in context
	username := context.Param("username")

	// Get profile data from database
	user, err := models.GetUserDataByUsername(username)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch user profile"})
		return
	}

	context.JSON(http.StatusOK, user)
}

func getUserAvatar(context *gin.Context) {
	// Get filename from param
	filename := context.Param("filename")

	// Check filename
	if filename != filepath.Base(filename) {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid filename"})
		return
	}

	// Get avatar url
	avatarUrl := utils.GetAvatarPath(&filename)

	if _, err := os.Stat(avatarUrl); err != nil {
		if os.IsNotExist(err) {
			avatarUrl = utils.GetDefaultAvatar()
		}
	}

	// Show the image
	context.File(avatarUrl)
}

func updateUserProfile(context *gin.Context) {
	// Get profile data using getUserDataForOperation function
	user := getMyUserData(context)
	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Get user's input
	if context.PostForm("username") != "" {
		username := context.PostForm("username")
		if !UsernameRegex.MatchString(username) {
			context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid username format. Must be 3-30 alphanumeric characters or underscores."})
			return
		}

		user.Username = username
	}
	if utils.IsFullNameValid(context.PostForm("full_name")) {
		user.FullName = context.PostForm("full_name")
	}
	userBio := context.PostForm("bio")
	user.Bio = &userBio

	// Handle image upload
	file, err := context.FormFile("avatar")
	if err == nil {
		// Save new avatar
		avatar, err := utils.SaveAvatar(file, context)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"message": "Failed to save avatar"})
			return
		}

		// Delete old avatar if exists
		if user.AvatarUrl != nil {
			err = utils.RemoveImage(user.AvatarUrl, "profile")
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to remove old avatar"})
				return
			}
		}

		// Assign new avatar to struct
		user.AvatarUrl = avatar
	}

	// Update user data in database
	err = user.Update()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user profile"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Update successful"})
}

func getMyUserData(context *gin.Context) *models.User {
	// Get user's ID from stored data in context
	uID := context.GetInt64("uID")

	// Get profile data from database
	user, err := models.GetUserDataForOps(uID)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch user data"})
		return nil
	}

	return user
}
