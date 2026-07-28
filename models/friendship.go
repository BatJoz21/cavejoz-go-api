package models

import (
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type Friendship struct {
	ID          int64      `json:"id"`
	RequesterID int64      `json:"requester_id"`
	AddresseeID int64      `json:"addressee_id"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
}

func (f *Friendship) SaveAddFriendData() error {
	query := `INSERT INTO friendships(requester_id, addressee_id, status)
		VALUES (?, ?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(f.RequesterID, f.AddresseeID, f.Status)
	if err != nil {
		return err
	}

	f.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func GetPendingFriends(uID int64) (*[]FriendDTO, error) {
	query := `SELECT f.id, u.id, u.username, u.avatar_url
	FROM friendships f
	JOIN users u ON u.id = f.requester_id
	WHERE f.status = 'pending' AND f.addressee_id = ?`
	rows, err := databases.DB.Query(query, uID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []FriendDTO
	for rows.Next() {
		var friend FriendDTO
		err = rows.Scan(&friend.FriendshipID, &friend.FriendUID,
			&friend.Username, &friend.AvatarUrl)
		if err != nil {
			return nil, err
		}

		friends = append(friends, friend)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &friends, nil
}

func GetUserFriends(uID int64) (*[]FriendDTO, error) {
	query := `SELECT f.id, u.id, u.username, u.avatar_url
	FROM friendships f
	JOIN users u ON u.id = CASE
		WHEN f.requester_id = ? THEN f.addressee_id
		ELSE f.requester_id
	END
	WHERE f.status = 'accepted' AND (f.requester_id = ? OR f.addressee_id = ?)`
	rows, err := databases.DB.Query(query, uID, uID, uID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []FriendDTO
	for rows.Next() {
		var friend FriendDTO
		err = rows.Scan(&friend.FriendshipID, &friend.FriendUID,
			&friend.Username, &friend.AvatarUrl)
		if err != nil {
			return nil, err
		}

		friends = append(friends, friend)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &friends, nil
}

func GetFriendProfile(uID, frienUID int64) (*FriendProfileDTO, error) {
	query := `SELECT
		f.id,
		u.id,
		u.username,
		u.full_name,
		u.bio,
		u.avatar_url
	FROM friendships f
	JOIN users u ON u.id = CASE
		WHEN f.requester_id = ? THEN f.addressee_id
		ELSE f.requester_id
	END
	WHERE f.status = 'accepted' AND (f.requester_id = ? OR f.addressee_id = ?) AND u.id = ?`
	row := databases.DB.QueryRow(query, uID, uID, uID, frienUID)

	var profile FriendProfileDTO
	err := row.Scan(&profile.FriendshipID, &profile.FriendUID, &profile.Username,
		&profile.FullName, &profile.Bio, &profile.AvatarUrl)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func IsFriendshipDataExists(uID, targetUID int64) (bool, int64) {
	query := `SELECT f.id
	FROM friendships f
	JOIN users u ON u.id = CASE
		WHEN f.requester_id = ? THEN f.addressee_id
		ELSE f.requester_id
	END
	WHERE
		(f.requester_id = ? AND f.addressee_id = ?) OR (f.requester_id = ? AND f.addressee_id = ?)`
	row := databases.DB.QueryRow(query, uID, uID, targetUID, targetUID, uID)

	var profile FriendProfileDTO
	err := row.Scan(&profile.FriendshipID)
	if err == nil {
		return true, profile.FriendshipID
	}

	return false, 0
}

func IsFriend(uID, targetUID int64) bool {
	query := `SELECT f.id
	FROM friendships f
	JOIN users u ON u.id = CASE
		WHEN f.requester_id = ? THEN f.addressee_id
		ELSE f.requester_id
	END
	WHERE
		status = 'accepted' AND
		(f.requester_id = ? AND f.addressee_id = ?) OR (f.requester_id = ? AND f.addressee_id = ?)`
	row := databases.DB.QueryRow(query, uID, uID, targetUID, targetUID, uID)

	var profile FriendProfileDTO
	err := row.Scan(&profile.FriendshipID)
	if err == nil {
		return true
	}

	return false
}

func UpdateFriendshipStatus(status string, friendshipID, AddresseeID int64) error {
	query := `UPDATE friendships SET status = ? WHERE id = ? AND addressee_id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, friendshipID, AddresseeID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteFriendship(friendshipID int64) error {
	query := `DELETE FROM friendships WHERE id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(friendshipID)
	return err
}
