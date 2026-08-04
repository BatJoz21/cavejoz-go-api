package models

import "time"

type NewCommentDTO struct {
	Content string `json:"content" binding:"required"`
}

type ViewCommentDTO struct {
	ID        int64      `json:"id"`
	PostID    int64      `json:"post_id"`
	UserID    int64      `json:"user_id"`
	Content   string     `json:"content"`
	CreatedAt *time.Time `json:"created_at"`
	Username  string     `json:"username"`
	AvatarUrl *string    `json:"avatar_url"`
}
