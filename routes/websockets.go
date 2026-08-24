package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"log"

	"github.com/BatJoz21/cavejoz-go-api/hub"
	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Origins allowed to open a WebSocket, from ALLOWED_ORIGINS (comma separated).
// Loaded lazily so godotenv.Load() in main has already run by the first upgrade.
var allowedWSOrigins = sync.OnceValue(func() map[string]bool {
	origins := map[string]bool{}
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins[strings.ToLower(o)] = true
		}
	}
	if len(origins) == 0 {
		log.Println("ALLOWED_ORIGINS is empty — WebSocket upgrades from browsers will be rejected")
	}
	return origins
})

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		// Non-browser clients (mobile app, CLI) send no Origin. Browsers always
		// send one on a handshake, so allowing the empty case opens no cross-site
		// path, and the single-use ticket is still required either way.
		if origin == "" {
			return true
		}

		if allowedWSOrigins()[strings.ToLower(origin)] {
			return true
		}

		log.Printf("rejected WebSocket upgrade from origin %q (%s)", origin, r.RemoteAddr)
		return false
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
