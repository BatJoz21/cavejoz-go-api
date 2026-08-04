package models

import "time"

type ViewPostDTO struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Caption    string     `json:"caption"`
	ContentUrl string     `json:"content_url"`
	Visibility string     `json:"visibility"`
	CreatedAt  *time.Time `json:"created_at"`
	Username   string     `json:"username"`
	AvatarUrl  *string    `json:"avatar_url"`
}
