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

const COMMENT_LIMIT_PER_PAGE = 10

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

func GetAllCommentByPostID(postID int64, offset int) (*[]ViewCommentDTO, error) {
	query := `SELECT
		c.id,
		c.post_id,
		c.user_id,
		c.content,
		c.created_at,
		u.username,
		u.avatar_url
	FROM comments c
	JOIN users u ON c.user_id = u.id
	WHERE c.post_id = ?
	ORDER BY c.created_at DESC
	LIMIT ? OFFSET ?`
	rows, err := databases.DB.Query(query, postID, COMMENT_LIMIT_PER_PAGE, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []ViewCommentDTO
	for rows.Next() {
		var c ViewCommentDTO
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.Username, &c.AvatarUrl); err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &comments, nil
}

func GetTotalCommentByPostID(postID int64) int {
	query := `SELECT COUNT(*) FROM comments WHERE post_id = ?`
	row := databases.DB.QueryRow(query, postID)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0
	}

	return total
}

func GetCommentByID(commentID int64) (*Comment, error) {
	query := `SELECT id, user_id, post_id, content FROM comments WHERE id = ?`
	row := databases.DB.QueryRow(query, commentID)

	var c Comment
	err := row.Scan(&c.ID, &c.UserID, &c.PostID, &c.Content)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Comment) Delete() error {
	query := `DELETE FROM comments WHERE id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.ID)
	if err != nil {
		return err
	}

	return nil
}
