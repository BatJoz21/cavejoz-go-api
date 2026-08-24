package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func sendMessage(context *gin.Context) {
	// Get conversation id from url parameter
	cID, err := strconv.ParseInt(context.Param("cID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Conversation not exists"})
		return
	}

	// Check conversation membership
	senderID := context.GetInt64("uID")
	if !models.IsConversationMember(cID, senderID) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
		return
	}
	_, err = models.GetOtherConversationParticipant(cID, senderID)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Conversation participant not found"})
		return
	}

	// Get user's input
	var dto models.SendMessageDTO
	if err := context.ShouldBindBodyWithJSON(&dto); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	// Check if message's content is an empty string
	if strings.TrimSpace(dto.Content) == "" {
		context.JSON(http.StatusNoContent, gin.H{"message": "No message"})
		return
	}

	// Save the message to database
	m := models.Message{
		ConversationID: cID,
		SenderID:       senderID,
		Content:        strings.TrimRight(dto.Content, " \t\n\r"),
	}
	if err := m.Save(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send message"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Message has been sent"})
}

func getConversationMessage(context *gin.Context) {
	// Get conversation id from url parameter
	cID, err := strconv.ParseInt(context.Param("cID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid conversation ID"})
		return
	}

	// Check conversation ownership
	if !models.IsConversationMember(cID, context.GetInt64("uID")) {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
		return
	}

	// Get cursor value from url query
	cursor, err := strconv.ParseInt(context.DefaultQuery("cursor", "0"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cursor"})
		return
	}

	// Get messages from database
	msgs, err := models.GetMessagesByConversationID(cID, cursor)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusOK, msgs)
			return
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch messages"})
			return
		}
	}

	// Count next cursor value
	nextCursor := int64(0)
	if len(*msgs) == models.MAX_SHOWN_MESSAGE {
		nextCursor = (*msgs)[len(*msgs)-1].ID
	}

	// Rearrange the message order
	for i, j := 0, len(*msgs)-1; i < j; i, j = i+1, j-1 {
		(*msgs)[i], (*msgs)[j] = (*msgs)[j], (*msgs)[i]
	}

	context.JSON(http.StatusOK, gin.H{
		"messages":    msgs,
		"next_cursor": nextCursor,
	})
}
