package models

type UserProfileDTO struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	FullName  string  `json:"full_name"`
	Bio       *string `json:"bio"`
	AvatarUrl *string `json:"avatar_url"`
}
