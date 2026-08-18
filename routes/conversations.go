package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func createChatRoom(context *gin.Context) {
	// Get both user's ids
	frID, err := strconv.ParseInt(context.PostForm("friendID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	uID := context.GetInt64("uID")

	// Check if both user isn't the same person
	if frID == uID {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Cannot chat with yourself"})
		return
	}

	// Check Friendship status
	status := models.GetFriendshipStatus(uID, frID)
	if status != "accepted" {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Users are not a friend"})
		return
	}

	// Check if conversation is exists
	isExists, err := models.CheckIfConversationExists(min(uID, frID), max(uID, frID))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if isExists {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Conversation already exists"})
		return
	}

	// Create new chat room save it in the database
	cnv := models.Conversation{
		UserAID: min(uID, frID),
		UserBID: max(uID, frID),
	}
	if err := cnv.Save(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Chat room created"})
}

func getConversationID(context *gin.Context) {
	// The caller is always one side of the conversation, only the other side
	// comes from the url query
	frID, err := strconv.ParseInt(context.Query("friendID"), 10, 64)
	if err != nil || frID <= 0 {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid friendID"})
		return
	}
	uID := context.GetInt64("uID")
	if frID == uID {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Cannot open a conversation with yourself"})
		return
	}

	// Check if conversation exists in database
	cID, err := models.GetConversationIDByUserIDs(min(uID, frID), max(uID, frID))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
		return
	}

	context.JSON(http.StatusOK, cID)
}

func getConversations(context *gin.Context) {
	// Get user's ID
	uID := context.GetInt64("uID")

	// Get current page and count offset
	page, err := strconv.Atoi(context.DefaultQuery("page", "1"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page"})
		return
	}
	offset := models.MAX_SHOWN_CONVERSATION * (page - 1)

	// Get conversations by user's ID
	data, err := models.GetConversationsByUID(uID, int(offset))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get conversations"})
		return
	}

	// Get total conversations of a user
	total := models.GetTotalConversationByUID(uID)

	context.JSON(http.StatusOK, gin.H{
		"conversations": data,
		"total":         total,
		"max":           models.MAX_SHOWN_CONVERSATION,
	})
}

func getConversation(context *gin.Context) {
	// Get conversations ID from url param and user's id
	cID, err := strconv.ParseInt(context.Param("cID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	uID := context.GetInt64("uID")

	// Check conversation ownership
	if !models.IsConversationMember(cID, uID) {
		context.JSON(http.StatusNotFound, gin.H{"message": "Conversation not found"})
		return
	}

	// Get conversation data from database
	c, err := models.GetConversation(cID, uID)
	// Check if result is empty
	if errors.Is(err, sql.ErrNoRows) {
		context.JSON(http.StatusNotFound, gin.H{"message": "Conversation not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get cursor value from url query
	cursor, err := strconv.ParseInt(context.DefaultQuery("cursor", "0"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get all messages data
	msgs, err := models.GetMessagesByConversationID(c.ID, cursor)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get next cursor value
	nextCursor := int64(0)
	if len(*msgs) == models.MAX_SHOWN_MESSAGE {
		nextCursor = (*msgs)[len(*msgs)-1].ID
	}

	context.JSON(http.StatusOK, gin.H{
		"conversation": c,
		"messages":     msgs,
		"next_cursor":  nextCursor,
	})
}

func updateReadRecordOnConversation(context *gin.Context) {
	// Get conversation ID and user ID
	cID, err := strconv.ParseInt(context.Param("cID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	uID := context.GetInt64("uID")

	// Get user's position in conversation (user a or user b)
	uPos, err := models.CheckUserPositionInConversation(cID, uID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	} else if uPos == "" {
		// If not a member, return an error
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user"})
		return
	}

	// Updated read messages data
	err = models.SetReadMessage(cID, uID, uPos)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Data updated"})
}
