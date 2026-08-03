package models

type NewCommentDTO struct {
	Content string `json:"content" binding:"required"`
}
