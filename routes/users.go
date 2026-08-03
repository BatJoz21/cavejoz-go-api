package routes

import (
	"net/http"
	"os"

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
	// Get filename from param
	filename := context.Param("filename")

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
			err = utils.RemoveImage(user.AvatarUrl, "profile")
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
