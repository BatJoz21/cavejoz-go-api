package middlewares

import (
	"net/http"
	"strings"

	"github.com/BatJoz21/cavejoz-go-api/utils"
	"github.com/gin-gonic/gin"
)

func Authenticate(context *gin.Context) {
	// Get token from authorization header
	token := context.Request.Header.Get("Authorization")
	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorize"})
		return
	}

	// Trim token's prefix
	token = strings.TrimPrefix(token, "Bearer ")

	// Verify token
	jwtToken, err := utils.VerifyToken(token)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	// Get claims data
	uID, username, role, err := utils.GetClaimsData(jwtToken)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	// Set data in context
	context.Set("uID", uID)
	context.Set("username", username)
	context.Set("role", role)

	context.Next()
}
