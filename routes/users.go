package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/BatJoz21/cavejoz-go-api/utils"
	"github.com/gin-gonic/gin"
)

func getMyUserProfile(context *gin.Context) {
	// Get profile data using getUserDataForOperation function
	user := getMyUserData(context)

	context.JSON(http.StatusOK, user)
}

func getUserProfile(context *gin.Context) {
	// Get user's username from stored data in context
	username := context.Param("username")

	// Get profile data from database
	user, err := models.GetUserDataByUsername(username)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, user)
}

func getUserAvatar(context *gin.Context) {
	// Get user's ID from param
	uID, err := strconv.ParseInt(context.Param("uID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get user data from database
	avatarUrl, err := models.GetStoredAvatarPath(uID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Check if avatar url is null
	if avatarUrl == nil {
		context.Status(http.StatusNoContent)
		return
	}

	context.File(*avatarUrl)
}

func updateUserProfile(context *gin.Context) {
	// Get profile data using getUserDataForOperation function
	user := getMyUserData(context)

	// Get user's input
	if context.PostForm("username") != "" {
		user.Username = context.PostForm("username")
	}
	if context.PostForm("full_name") != "" {
		user.FullName = context.PostForm("full_name")
	}
	userBio := context.PostForm("bio")
	user.Bio = &userBio

	// Handle image upload
	file, err := context.FormFile("avatar")
	if err == nil {
		// Save new avatar
		avatar, err := utils.SaveAvatar(file, user.Username, context)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Delete old avatar if exists
		if user.AvatarUrl != nil {
			err = utils.RemoveImage(user.AvatarUrl)
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
		}

		// Assign new avatar to struct
		user.AvatarUrl = avatar
	}

	// Update user data in database
	err = user.Update()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return nil
	}

	return user
}
