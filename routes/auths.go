package routes

import (
	"net/http"
	"time"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/BatJoz21/cavejoz-go-api/utils"
	"github.com/gin-gonic/gin"
)

func registerNewUser(context *gin.Context) {
	// Upload user's avatar
	var avatar *string
	file, err := context.FormFile("avatar")
	if err == nil {
		avatar, err = utils.SaveAvatar(file, context.PostForm("username"), context)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"isRegistered": false,
				"message": err.Error()})
			return
		}
	}

	// Hashing password
	hashedPassword, err := utils.HashPassword(context.PostForm("password"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"isRegistered": false,
			"message": err.Error()})
		return
	}

	// Store user's data into database
	user := models.User{
		Username:     context.PostForm("username"),
		Email:        context.PostForm("email"),
		PasswordHash: hashedPassword,
		FullName:     context.PostForm("full_name"),
		Bio:          nil,
		Role:         "user",
		AvatarUrl:    avatar,
	}
	err = user.Save()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"isRegistered": false,
			"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"isRegistered": true,
		"message": "New user registered, you can now log in to your account"})
}

func login(context *gin.Context) {
	// Get user's input
	var dto models.UserLoginDTO
	err := context.ShouldBindJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Validate user's input (email and password)
	err = models.ValidateCredentials(&dto)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	// Generate access token (JWT)
	userData, err := models.GetUserDataByEmail(dto.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	accessToken, err := utils.GenerateAccessToken(userData.ID, userData.Username, userData.Role)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Generate refresh token and store it in the database
	refreshToken, err := utils.GenerateRefreshToken(userData.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	expire := time.Now().Add(time.Hour * 24 * 7)
	deviceName := dto.DeviceName
	if deviceName == "" {
		deviceName = "Unknown device"
	}
	refreshTokenObj := models.RefreshToken{
		UserID:     userData.ID,
		DeviceName: &deviceName,
		TokenHash:  utils.HashRefreshToken(refreshToken),
		ExpiresAt:  &expire,
	}
	err = refreshTokenObj.StoreRefreshToken()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          userData,
	})
}

func refreshAccessToken(context *gin.Context) {
	// Get user's input
	var dto models.RefreshTokenRequest
	err := context.ShouldBindJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Verify refresh token
	_, err = utils.VerifyToken(dto.RefreshToken)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Hash refresh token
	hashed := utils.HashRefreshToken(dto.RefreshToken)

	// Check if refresh token exists and valid
	refreshTokenObj, err := models.GetRefreshTokenByHashedToken(hashed)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if refreshTokenObj.RevokedAt != nil || time.Now().After(*refreshTokenObj.ExpiresAt) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}

	// Generate access token (JWT)
	userData, err := models.GetUserDataByUID(refreshTokenObj.UserID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	accessToken, err := utils.GenerateAccessToken(userData.ID, userData.Username, userData.Role)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message":      "session refreshed",
		"access_token": accessToken,
	})
}

func logout(context *gin.Context) {
	// Get user's input
	var dto models.RefreshTokenRequest
	err := context.ShouldBindJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"isLoggedOut": false,
			"message":     err.Error(),
		})
		return
	}

	// Verify refresh token
	_, err = utils.VerifyToken(dto.RefreshToken)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"isLoggedOut": false,
			"message":     err.Error(),
		})
		return
	}

	// Hash refresh token
	hashed := utils.HashRefreshToken(dto.RefreshToken)
	refreshTokenObj, err := models.GetRefreshTokenByHashedToken(hashed)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"isLoggedOut": false,
			"message":     err.Error(),
		})
		return
	}

	// Revoke refresh token
	err = refreshTokenObj.RevokeRefreshToken()
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"isLoggedOut": false,
			"message":     err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"isLoggedOut": true,
		"message":     "Log out successful",
	})
}
