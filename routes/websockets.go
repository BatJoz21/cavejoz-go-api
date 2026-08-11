package routes

import (
	"log"
	"net/http"

	"github.com/BatJoz21/cavejoz-go-api/hub"
	"github.com/BatJoz21/cavejoz-go-api/utils"
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
		// Read the token from the url query
		authToken := context.Query("token")
		if authToken == "" {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorize"})
			return
		}

		// Verify the token
		jwtToken, err := utils.VerifyToken(authToken)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorize"})
			return
		}

		// Get user's id from token's claims
		uID, _, _, err := utils.GetClaimsData(jwtToken)
		if err != nil {
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
		client.ReadPump(h)

		log.Printf("User %d disconnected - total connections: %d", uID, h.Count(uID))
	}
}
