package models

import "time"

type ViewNotifDTO struct {
	ID          int64      `json:"id"`
	ActorID     int64      `json:"actor_id"`
	Username    string     `json:"username"`
	AvatarUrl   *string    `json:"avatar_url"`
	Type        string     `json:"type"`
	ReferenceID int64      `json:"reference_id"`
	Preview     string     `json:"preview"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   *time.Time `json:"created_at"`
}
