package routes

import (
	"encoding/json"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

type typingPayload struct {
	ConversationID int64 `json:"conversation_id"`
}

func handleTyping(uID int64, raw []byte) {
	var p typingPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}

	recipientID, err := models.GetOtherConversationParticipant(p.ConversationID, uID)
	if err != nil {
		return
	}

	appHub.Send(recipientID, gin.H{
		"type":            "typing",
		"conversation_id": p.ConversationID,
		"user_id":         uID,
	})
}
