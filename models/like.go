package models

import (
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type Like struct {
	ID        int64      `json:"id"`
	PostID    int64      `json:"post_id"`
	UserID    int64      `json:"user_id"`
	CreatedAt *time.Time `json:"created_at"`
}

func (l *Like) SaveLike() error {
	query := `INSERT INTO likes(post_id, user_id) VALUES(?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(l.PostID, l.UserID)
	if err != nil {
		return err
	}

	l.ID, err = result.LastInsertId()
	return err
}

func CheckIfLikeExists(postID, userID int64) (bool, int64) {
	query := `SELECT id FROM likes WHERE post_id = ? AND user_id = ?`
	row := databases.DB.QueryRow(query, postID, userID)

	var id int64
	err := row.Scan(&id)
	if err != nil {
		return false, 0
	}

	return true, id
}

func TotalLikeofAPost(postID int64) int {
	query := `SELECT COUNT(*) FROM likes WHERE post_id = ?`
	row := databases.DB.QueryRow(query, postID)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0
	}

	return total
}

func DeleteLike(likeID int64) error {
	query := `DELETE FROM likes WHERE id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(likeID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteAllLikeByPostID(postID int64) error {
	query := `DELETE FROM likes WHERE post_id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(postID)
	if err != nil {
		return err
	}

	return nil
}
