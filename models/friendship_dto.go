package models

type AddFriendDTO struct {
	AddresseeID string `json:"addressee_id" binding:"required"`
}

type FriendDTO struct {
	FriendshipID int64   `json:"friendship_id"`
	FriendUID    int64   `json:"friend_u_id"`
	Username     string  `json:"username"`
	AvatarUrl    *string `json:"avatar_url"`
}

type FriendProfileDTO struct {
	FriendshipID int64   `json:"friendship_id"`
	FriendUID    int64   `json:"friend_u_id"`
	Username     string  `json:"username"`
	FullName     string  `json:"full_name"`
	Bio          *string `json:"bio"`
	AvatarUrl    *string `json:"avatar_url"`
}
