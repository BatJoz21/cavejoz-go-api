package routes

import (
	"net/http"

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
		client.ReadPump(h)
	}
}
