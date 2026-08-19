package models

import "time"

type SendMessageDTO struct {
	Content        string `json:"content" binding:"required"`
}

type ViewMessageDTO struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	Username       string    `json:"username"`
	AvatarUrl      *string   `json:"avatar_url"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}
