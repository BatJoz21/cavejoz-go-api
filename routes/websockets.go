package routes

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/BatJoz21/cavejoz-go-api/hub"
	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(h *hub.Hub) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Read the ticket from the url query
		ticket := context.Query("ticket")
		if ticket == "" {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorize"})
			return
		}

		// Get user's id from consuming ticket
		uID, ok := models.ConsumeWSTicket(ticket)
		if !ok {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorize"})
			return
		}

		// Upgrade
		conn, err := upgrader.Upgrade(context.Writer, context.Request, nil)
		if err != nil {
			return
		}

		// Add new connection
		client := hub.NewClient(uID, conn)
		h.Add(uID, client)

		go client.WritePump()
		client.ReadPump(h, handleSocketMessage)
	}
}

func handleSocketMessage(uID int64, msgType string, raw []byte) {
	switch msgType {
	case "send_message":
		handleSendMessage(uID, raw)
	case "typing":
		handleTyping(uID, raw)
	default:
	}
}

func handleSendMessage(uID int64, raw []byte) {
	var p models.SendMessagePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}

	content := strings.TrimSpace(p.Content)
	if content == "" || len(content) > 2000 {
		return
	}
	if p.ConversationID == 0 {
		return
	}

	// Authorize + get recipient in one query
	recipientID, err := models.GetOtherConversationParticipant(p.ConversationID, uID)
	if err != nil {
		return
	}

	// Insert
	msg := models.Message{
		ConversationID: p.ConversationID,
		SenderID:       uID,
		Content:        p.Content,
	}
	if err := msg.Save(); err != nil {
		return
	}

	// Re-fetch so the payload matches the REST shape exactly
	viewMsg, err := models.GetMessageByID(msg.ID)
	if err != nil {
		return
	}

	payLoad := gin.H{
		"type":    "message",
		"message": viewMsg,
	}

	appHub.Send(recipientID, payLoad)
	appHub.Send(uID, payLoad)
}
