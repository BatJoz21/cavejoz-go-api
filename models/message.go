package models

import (
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type Message struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

const MAX_SHOWN_MESSAGE = 20

func (m *Message) Save() error {
	query := `INSERT INTO messages(conversation_id, sender_id, content) VALUES (?, ?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(m.ConversationID, m.SenderID, m.Content)
	if err != nil {
		return err
	}

	m.ID, err = result.LastInsertId()
	return err
}

func GetMessagesByConversationID(cID, cursor int64) (*[]Message, error) {
	query := `SELECT
		id,
		sender_id,
		content,
		created_at
	FROM messages
	WHERE conversation_id = ? AND (? = 0 OR id < ?)
	ORDER BY id DESC LIMIT ?`
	rows, err := databases.DB.Query(query, cID, cursor, cursor, MAX_SHOWN_MESSAGE)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}

		msgs = append(msgs, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &msgs, nil
}
