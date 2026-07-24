package models

type UserLoginDTO struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SessionDataDTO struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username"`
	Role         string  `json:"role"`
}
