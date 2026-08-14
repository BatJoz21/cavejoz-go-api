package models

type SendMessagePayload struct {
	ConversationID int64  `json:"conversation_id"`
	Content        string `json:"content"`
}
