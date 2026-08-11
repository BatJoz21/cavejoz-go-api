package routes

import (
	"net/http"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func getNewWSTicket(context *gin.Context) {
	// Get web socket ticket
	ticket, err := models.IssueWSTicket(context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to issue ws ticket"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"ticket": ticket})
}
