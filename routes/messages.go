package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func sendMessage(context *gin.Context) {
	// Get conversation id from url parameter
	cID, err := strconv.ParseInt(context.Param("cID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "DM not exists"})
		return
	}

	// Get user's input
	var dto models.SendMessageDTO
	if err := context.ShouldBindBodyWithJSON(&dto); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Save the message to database
	m := models.Message{
		ConversationID: cID,
		SenderID:       context.GetInt64("uID"),
		Content:        dto.Content,
	}
	if err := m.Save(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Message has been sent"})
}
