package models

import (
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type Comment struct {
	ID        int64      `json:"id"`
	PostID    int64      `json:"post_id"`
	UserID    int64      `json:"user_id"`
	Content   string     `json:"content"`
	CreatedAt *time.Time `json:"created_at"`
}

func (c *Comment) Save() error {
	query := `INSERT INTO comments(post_id, user_id, content)
		VALUES(?, ?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(c.PostID, c.UserID, c.Content)
	if err != nil {
		return err
	}

	c.ID, err = result.LastInsertId()
	return err
}

func GetAllCommentByPostID(postID int64) (*[]Comment, error) {
	query := `SELECT
		id,
		post_id,
		user_id,
		content,
		created_at
	FROM comments WHERE post_id = ?`
	rows, err := databases.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(c.ID, c.PostID, c.UserID, c.Content, c.CreatedAt); err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &comments, nil
}
