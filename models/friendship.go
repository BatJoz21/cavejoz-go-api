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

const FriendshipListDataMax = 10

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

func GetPendingFriends(uID int64, offset int) (*[]FriendDTO, error) {
	query := `SELECT f.id, u.id, u.username, u.full_name, u.avatar_url
	FROM friendships f
	JOIN users u ON u.id = f.requester_id
	WHERE f.status = 'pending' AND f.addressee_id = ?
	LIMIT ? OFFSET ?`
	rows, err := databases.DB.Query(query, uID, FriendshipListDataMax, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []FriendDTO
	for rows.Next() {
		var friend FriendDTO
		err = rows.Scan(&friend.FriendshipID, &friend.FriendUID,
			&friend.Username, &friend.FullName, &friend.AvatarUrl)
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

func GetTotalPending(uID int64) int {
	query := `SELECT COUNT(*)
	FROM friendships f
	JOIN users u ON u.id = f.requester_id
	WHERE f.status = 'pending' AND f.addressee_id = ?`
	row := databases.DB.QueryRow(query, uID)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0
	}

	return total
}

func GetFriendshipList(uID int64, status string, offset int) (*[]FriendDTO, error) {
	query := `SELECT f.id, u.id, u.username, u.full_name, u.avatar_url
	FROM friendships f
	JOIN users u ON u.id = CASE
		WHEN f.requester_id = ? THEN f.addressee_id
		ELSE f.requester_id
	END
	WHERE f.status = ? AND (f.requester_id = ? OR f.addressee_id = ?)
	LIMIT ? OFFSET ?`
	rows, err := databases.DB.Query(query, uID, status, uID, uID, FriendshipListDataMax, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []FriendDTO
	for rows.Next() {
		var friend FriendDTO
		err = rows.Scan(&friend.FriendshipID, &friend.FriendUID,
			&friend.Username, &friend.FullName, &friend.AvatarUrl)
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

func GetFriendsID(uID int64) (*[]int64, error) {
	query := `SELECT u.id
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

	var ids []int64
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ids, nil
}

func GetTotalFriendshipDataByUID(uID int64, status string) (int, error) {
	query := `SELECT COUNT(*)
	FROM friendships f
	JOIN users u ON u.id = CASE
		WHEN f.requester_id = ? THEN f.addressee_id
		ELSE f.requester_id
	END
	WHERE f.status = ? AND (f.requester_id = ? OR f.addressee_id = ?)`
	row := databases.DB.QueryRow(query, uID, status, uID, uID)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func GetFriendshipStatus(uID, targetUID int64) string {
	query := `SELECT f.status
	FROM friendships f
	JOIN users u ON u.id = CASE
		WHEN f.requester_id = ? THEN f.addressee_id
		ELSE f.requester_id
	END
	WHERE
		(f.requester_id = ? AND f.addressee_id = ?) OR (f.requester_id = ? AND f.addressee_id = ?)`
	row := databases.DB.QueryRow(query, uID, uID, targetUID, targetUID, uID)

	var status string
	err := row.Scan(&status)
	if err == nil {
		return status
	}

	return ""
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

func UpdateFriendshipStatusToAccepted(friendshipID, AddresseeID int64) error {
	query := `UPDATE friendships SET status = 'accepted' 
	WHERE id = ? AND addressee_id = ? AND status = 'pending'`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(friendshipID, AddresseeID)
	if err != nil {
		return err
	}

	return nil
}

func UpdateFriendshipStatusToBlocked(friendshipID int64) error {
	query := `UPDATE friendships SET status = 'blocked' WHERE id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(friendshipID)
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
