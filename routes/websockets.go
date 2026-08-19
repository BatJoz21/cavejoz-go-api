package routes

import (
	"encoding/json"
	"log"
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
			log.Println("upgrade failed:", err)
			return
		}

		// Add new connection
		client := hub.NewClient(uID, conn)
		h.Add(uID, client)
		log.Printf("User %d connected — total connections: %d", uID, h.Count(uID))

		go client.WritePump()
		client.ReadPump(h, handleSocketMessage)

		log.Printf("User %d disconnected - total connections: %d", uID, h.Count(uID))
	}
}

func handleSocketMessage(uID int64, msgType string, raw []byte) {
	switch msgType {
	case "send_message":
		handleSendMessage(uID, raw)
	case "typing":
		handleTyping(uID, raw)
	default:
		log.Printf("unknown frame type %q from user %d", msgType, uID)
	}
}

func handleSendMessage(uID int64, raw []byte) {
	log.Printf("[1] handleSendMessage entered, user %d", uID)

	var p models.SendMessagePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("bad send_message from user %d: %v", uID, err)
		return
	}
	log.Printf("[2] parsed: conv=%d content=%q", p.ConversationID, p.Content)

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
		log.Printf("An error occured: %v", err.Error())
		return
	}
	log.Printf("[3] authorized, recipient=%d", recipientID)

	// Insert
	msg := models.Message{
		ConversationID: p.ConversationID,
		SenderID:       uID,
		Content:        p.Content,
	}
	if err := msg.Save(); err != nil {
		log.Printf("Failed to send message: %v", err.Error())
		return
	}

	// Re-fetch so the payload matches the REST shape exactly
	viewMsg, err := models.GetMessageByID(msg.ID)
	if err != nil {
		log.Printf("Failed to fetch message: %v", err.Error())
		return
	}

	payLoad := gin.H{
		"type":    "message",
		"message": viewMsg,
	}

	log.Printf("pushing message %d to users %d and %d", viewMsg.ID, recipientID, uID)
	log.Printf("recipient %d has %d connections, sender %d has %d", recipientID, appHub.Count(recipientID), uID, appHub.Count(uID))
	appHub.Send(recipientID, payLoad)
	appHub.Send(uID, payLoad)
}
