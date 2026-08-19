package models

import (
	"database/sql"
	"errors"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type Conversation struct {
	ID              int64 `json:"id"`
	UserAID         int64 `json:"user_a_id"`
	UserBID         int64 `json:"user_b_id"`
	UserALastReadID int64 `json:"user_a_last_read_id"`
	UserBLastReadID int64 `json:"user_b_last_read_id"`
}

const MAX_SHOWN_CONVERSATION = 10

func (c *Conversation) Save() error {
	query := `INSERT INTO conversations(user_a_id, user_b_id) VALUES (?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(c.UserAID, c.UserBID)
	if err != nil {
		return err
	}

	c.ID, err = result.LastInsertId()
	return err
}

func CheckIfConversationExists(aID, bID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT id FROM conversations WHERE user_a_id = ? AND user_b_id = ? LIMIT 1)`
	row := databases.DB.QueryRow(query, aID, bID)

	var res int
	if err := row.Scan(&res); err != nil {
		return true, err
	}

	if res == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func GetConversationIDByUserIDs(aID, bID int64) (int64, error) {
	query := `SELECT id FROM conversations WHERE user_a_id = ? AND user_b_id = ?`
	row := databases.DB.QueryRow(query, aID, bID)

	var cID int64
	if err := row.Scan(&cID); err != nil {
		return 0, err
	}

	return cID, nil
}

func IsConversationMember(cID, uID int64) bool {
	query := `SELECT EXISTS(SELECT 1 FROM conversations WHERE id = ? AND (user_a_id = ? OR user_b_id = ?))`
	row := databases.DB.QueryRow(query, cID, uID, uID)

	var res int
	if err := row.Scan(&res); err != nil {
		return false
	}

	return res == 1
}

func GetOtherConversationParticipant(cID, uID int64) (int64, error) {
	query := `SELECT 
	CASE 
		WHEN user_a_id = ? THEN user_b_id ELSE user_a_id 
	END FROM conversations WHERE id = ? AND (user_a_id = ? OR user_b_id = ?)`
	row := databases.DB.QueryRow(query, uID, cID, uID, uID)

	var otherID int64
	err := row.Scan(&otherID)
	if err != nil {
		return 0, err
	}

	return otherID, nil
}

func CheckUserPositionInConversation(cID, uID int64) (string, error) {
	query := `SELECT
	CASE
		WHEN user_a_id = ? THEN 'a'
		ELSE 'b'
	END
	FROM conversations WHERE id = ? AND (user_a_id = ? OR user_b_id = ?)`

	var pos string
	err := databases.DB.QueryRow(query, uID, cID, uID, uID).Scan(&pos)
	if err != nil {
		return "", err
	}

	return pos, nil
}

func GetConversationsByUID(uID int64, offset int) (*[]ViewConversationDTO, error) {
	query := `SELECT
		c.id,
		c.user_a_id,
		c.user_b_id,
		u.username,
		u.avatar_url
	FROM conversations c
	JOIN users u ON u.id = CASE
		WHEN c.user_a_id = ? THEN c.user_b_id
		ELSE c.user_a_id
	END
	WHERE c.user_a_id = ? OR c.user_b_id = ?
	LIMIT ? OFFSET ?`
	rows, err := databases.DB.Query(query, uID, uID, uID, MAX_SHOWN_CONVERSATION, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []ViewConversationDTO
	for rows.Next() {
		var c ViewConversationDTO
		if err := rows.Scan(&c.ID, &c.UserAID, &c.UserBID, &c.Username, &c.AvatarUrl); err != nil {
			return nil, err
		}

		query := `SELECT content, created_at FROM messages WHERE conversation_id = ? ORDER BY id DESC LIMIT 1`
		row := databases.DB.QueryRow(query, c.ID)
		err = row.Scan(&c.LastMessage, &c.LastMessageAt)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		c.HasUnread, err = CheckIfConversationHasUnreadByUID(uID, c.ID)
		if err != nil {
			return nil, err
		}

		data = append(data, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &data, nil
}

func GetConversation(cID, uID int64) (*ViewConversationDTO, error) {
	query := `SELECT
		c.id,
		c.user_a_id,
		c.user_b_id,
		c.user_a_last_read_id,
		c.user_b_last_read_id,
		u.username,
		u.avatar_url
	FROM conversations c
	JOIN users u ON u.id = CASE
		WHEN c.user_a_id = ? THEN c.user_b_id
		ELSE c.user_a_id
	END
	WHERE c.id = ?`
	row := databases.DB.QueryRow(query, uID, cID)

	var c ViewConversationDTO
	if err := row.Scan(&c.ID, &c.UserAID, &c.UserBID, &c.UserALastReadID,
		&c.UserBLastReadID, &c.Username, &c.AvatarUrl); err != nil {
		return nil, err
	}

	return &c, nil
}

func GetTotalConversationByUID(uID int64) int {
	query := `SELECT COUNT(*) FROM conversations
	WHERE CASE WHEN user_a_id = ? THEN user_a_id ELSE user_b_id END = ?`
	var total int

	err := databases.DB.QueryRow(query, uID, uID).Scan(&total)
	if err != nil {
		return 0
	}

	return total
}

func CheckIfConversationHasUnreadByUID(uID, cID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT c.id 
	FROM conversations c
	JOIN messages m ON c.id = m.conversation_id AND 
	CASE 
		WHEN c.user_a_id = ? THEN c.user_b_id 
		ELSE c.user_a_id 
	END = m.sender_id
	WHERE CASE
		WHEN c.user_a_id = ? THEN c.user_a_last_read_id
		ELSE c.user_b_last_read_id
	END < m.id AND c.id = ?)`
	row := databases.DB.QueryRow(query, uID, uID, cID)

	var res int
	if err := row.Scan(&res); err != nil {
		return false, err
	}

	if res == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func SetReadMessage(cID, uID int64, position string) error {
	query := `UPDATE conversations SET`
	switch position {
	case "a":
		query += ` user_a_last_read_id =`
	case "b":
		query += ` user_b_last_read_id =`
	default:
		return errors.New("Invalid user")
	}

	query += ` (SELECT COALESCE(MAX(id), 0) FROM messages WHERE conversation_id = ? AND sender_id != ?)
	WHERE id = ?`
	_, err := databases.DB.Exec(query, cID, uID, cID)
	if err != nil {
		return err
	}

	return nil
}
