package models

import "time"

type ViewConversationDTO struct {
	ID              int64      `json:"id"`
	UserAID         int64      `json:"user_a_id"`
	UserBID         int64      `json:"user_b_id"`
	UserALastReadID int64      `json:"user_a_last_read_id"`
	UserBLastReadID int64      `json:"user_b_last_read_id"`
	Username        string     `json:"username"`
	AvatarUrl       *string    `json:"avatar_url"`
	LastMessage     *string    `json:"last_message"`
	LastMessageAt   *time.Time `json:"last_message_at"`
	HasUnread       bool       `json:"has_unread"`
}
