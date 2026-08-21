package routes

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/BatJoz21/cavejoz-go-api/utils"
	"github.com/gin-gonic/gin"
)

var UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

func registerNewUser(context *gin.Context) {
	// Upload user's avatar
	var avatar *string
	file, err := context.FormFile("avatar")
	if err == nil {
		avatar, err = utils.SaveAvatar(file, context)
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

	// Checking username
	username := context.PostForm("username")
	if !UsernameRegex.MatchString(username) {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid username format. Must be 3-30 alphanumeric characters or underscores."})
		return
	}

	// Store user's data into database
	user := models.User{
		Username:     username,
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
	// Get user's input. ShouldBindBodyWithJSON reads the copy cached by
	// ThrottleLoginByAccount, which has already consumed the raw body.
	var dto models.UserLoginDTO
	err := context.ShouldBindBodyWithJSON(&dto)
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
	expire := time.Now().Add(utils.RefreshTokenTTL)
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
	parsedToken, err := utils.VerifyToken(dto.RefreshToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}

	// Read the uID the token was signed with, so we can check it against the
	// stored row rather than trusting the row alone.
	claimUID, err := utils.GetRefreshClaims(parsedToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}

	// Hash refresh token
	hashed := utils.HashRefreshToken(dto.RefreshToken)

	// Check if refresh token exists and valid
	refreshTokenObj, err := models.GetRefreshTokenByHashedToken(hashed)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}
	if refreshTokenObj.UserID != claimUID {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}

	// A token we already spent is being presented again. Just after rotation
	// that is almost certainly a client firing two refreshes at once, so we
	// only reject it. Long after rotation it means someone kept a copy, and we
	// cannot tell the legitimate holder from a thief — so end every session.
	if refreshTokenObj.RevokedAt != nil {
		if time.Since(*refreshTokenObj.RevokedAt) > utils.RefreshReuseGrace {
			log.Printf("refresh token replay for user %d, revoking all sessions", refreshTokenObj.UserID)
			if err := models.RevokeAllRefreshTokensByUserID(refreshTokenObj.UserID); err != nil {
				log.Printf("failed to revoke sessions for user %d: %v", refreshTokenObj.UserID, err)
			}
		}
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}

	if refreshTokenObj.ExpiresAt == nil || time.Now().After(*refreshTokenObj.ExpiresAt) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}

	// Generate access token (JWT)
	userData, err := models.GetUserDataByUID(refreshTokenObj.UserID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to refresh session"})
		return
	}
	accessToken, err := utils.GenerateAccessToken(userData.ID, userData.Username, userData.Role)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to refresh session"})
		return
	}

	// Rotate the refresh token: the one just presented is spent.
	newRefreshToken, err := utils.GenerateRefreshToken(userData.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to refresh session"})
		return
	}

	expire := time.Now().Add(utils.RefreshTokenTTL)
	err = models.RotateRefreshToken(
		hashed,
		utils.HashRefreshToken(newRefreshToken),
		refreshTokenObj.UserID,
		refreshTokenObj.DeviceName,
		expire,
	)
	if errors.Is(err, models.ErrRefreshTokenReuse) {
		// Lost the compare-and-swap: another request spent this token in the
		// moment between our check above and this update. That is a race, not
		// theft, so reject this one and leave the winner's session alone.
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to refresh session"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to refresh session"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message":       "session refreshed",
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
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
