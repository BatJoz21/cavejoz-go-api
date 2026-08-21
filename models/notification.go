package models

import (
	"errors"
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type Notification struct {
	ID          int64      `json:"id"`
	RecipientID int64      `json:"recipient_id"`
	ActorID     int64      `json:"actor_id"`
	Type        string     `json:"type"`
	ReferenceID int64      `json:"reference_id"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   *time.Time `json:"created_at"`
}

func (n *Notification) Save() error {
	// Check recipient and actor
	if n.RecipientID == n.ActorID {
		return errors.New("Cannot send notification to yourself")
	}

	query := `INSERT INTO notifications(recipient_id, actor_id, type, reference_id, is_read)
		VALUES(?, ?, ?, ?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}

	result, err := stmt.Exec(n.RecipientID, n.ActorID, n.Type, n.ReferenceID, n.IsRead)
	if err != nil {
		return err
	}

	n.ID, err = result.LastInsertId()
	return err
}

func GetNotificationsByRecipientID(recID int64, limit int) (*[]ViewNotifDTO, error) {
	query := `SELECT
		n.id,
		n.actor_id,
		u.username,
		u.avatar_url,
		n.type,
		n.reference_id,
		n.is_read,
		n.created_at
	FROM notifications n
	JOIN users u ON n.actor_id = u.id
	WHERE n.recipient_id = ? AND n.is_read = 0
	ORDER BY n.created_at DESC`

	var args []any
	args = append(args, recID)

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := databases.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var notifs []ViewNotifDTO
	for rows.Next() {
		var n ViewNotifDTO
		err = rows.Scan(&n.ID, &n.ActorID, &n.Username, &n.AvatarUrl, &n.Type, &n.ReferenceID, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, err
		}

		switch n.Type {
		case "like":
			n.Preview = n.Username + " has liked your post"
		case "comment":
			n.Preview = n.Username + " has commented on your post"
		case "friend_request":
			n.Preview = n.Username + " has sent you a friend request"
		case "friend_accept":
			n.Preview = n.Username + " has accepted your friend request"
		}

		notifs = append(notifs, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &notifs, nil
}

func GetViewNotificationByID(id int64) (*ViewNotifDTO, error) {
	query := `SELECT
		n.id,
		n.actor_id,
		u.username,
		u.avatar_url,
		n.type,
		n.reference_id,
		n.is_read,
		n.created_at
	FROM notifications n
	JOIN users u ON n.actor_id = u.id
	WHERE n.id = ? AND n.is_read = 0
	ORDER BY n.created_at DESC`

	row := databases.DB.QueryRow(query, id)

	var n ViewNotifDTO
	err := row.Scan(&n.ID, &n.ActorID, &n.Username, &n.AvatarUrl, &n.Type, &n.ReferenceID, &n.IsRead, &n.CreatedAt)
	if err != nil {
		return nil, err
	}

	switch n.Type {
	case "like":
		n.Preview = n.Username + " has liked your post"
	case "comment":
		n.Preview = n.Username + " has commented on your post"
	case "friend_request":
		n.Preview = n.Username + " has sent you a friend request"
	case "friend_accept":
		n.Preview = n.Username + " has accepted your friend request"
	}

	return &n, nil
}

func GetNotificationByID(notifID int64) (*Notification, error) {
	query := `SELECT
		n.id,
		n.recipient_id,
		n.actor_id,
		n.type,
		n.reference_id,
		n.is_read,
		n.created_at
	FROM notifications n WHERE n.id = ?`
	row := databases.DB.QueryRow(query, notifID)

	var n Notification
	err := row.Scan(&n.ID, &n.RecipientID, &n.ActorID, &n.Type, &n.ReferenceID, &n.IsRead, &n.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (n *Notification) UpdateIsRead(isRead int) error {
	query := `UPDATE notifications SET is_read = ? WHERE id = ? AND recipient_id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(isRead, n.ID, n.RecipientID)
	if err != nil {
		return err
	}

	return nil
}

func UpdateIsReadByRecipientID(isRead int, recipientID int64) error {
	query := `UPDATE notifications SET is_read = ? WHERE recipient_id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(isRead, recipientID)
	if err != nil {
		return err
	}

	return nil
}
