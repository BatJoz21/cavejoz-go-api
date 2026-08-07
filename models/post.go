package models

import (
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type Post struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Caption    string     `json:"caption"`
	ContentUrl string     `json:"content_url"`
	Visibility string     `json:"visibility"`
	CreatedAt  *time.Time `json:"created_at"`
}

const (
	PostsLimitPerPage = 5
	FeedsLimit        = 2
)

func (p *Post) Save() error {
	query := `INSERT INTO posts(user_id, caption, content_url, visibility)
		VALUES (?, ?, ?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}

	result, err := stmt.Exec(p.UserID, p.Caption, p.ContentUrl, p.Visibility)
	if err != nil {
		return err
	}

	p.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func GetPostsForFeedByUIDs(frID []int64, offset int) (*[]ViewPostDTO, error) {
	query := `SELECT
		p.id,
		p.user_id,
		p.caption,
		p.content_url,
		p.visibility,
		p.created_at,
		u.username,
		u.avatar_url
	FROM posts p
	JOIN users u ON p.user_id = u.id`

	if len(frID) == 0 {
		return nil, nil
	}

	var args []any
	for i := 0; i < len(frID); i++ {
		if i == 0 {
			query += ` WHERE p.user_id = ?`
			args = append(args, frID[i])
		} else {
			query += ` OR p.user_id = ?`
			args = append(args, frID[i])
		}
	}

	query += ` ORDER BY p.created_at DESC
	LIMIT ? OFFSET ?`
	args = append(args, FeedsLimit)
	args = append(args, offset)

	rows, err := databases.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var posts []ViewPostDTO
	for rows.Next() {
		var post ViewPostDTO
		if err := rows.Scan(&post.ID, &post.UserID, &post.Caption, &post.ContentUrl,
			&post.Visibility, &post.CreatedAt, &post.Username, &post.AvatarUrl); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &posts, nil
}

func GetAllPostsofAUser(uID int64, visibility string, offset int) (*[]ViewPostDTO, error) {
	query := `SELECT
		p.id,
		p.user_id,
		p.caption,
		p.content_url,
		p.visibility,
		p.created_at,
		u.username,
		u.avatar_url
	FROM posts p
	JOIN users u ON p.user_id = u.id
	WHERE user_id = ?`

	var args []any
	args = append(args, uID)

	if visibility == "public" || visibility == "friends" {
		query += ` AND visibility = ?`
		args = append(args, visibility)
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, PostsLimitPerPage)
	args = append(args, offset)

	rows, err := databases.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var posts []ViewPostDTO
	for rows.Next() {
		var post ViewPostDTO
		if err := rows.Scan(&post.ID, &post.UserID, &post.Caption, &post.ContentUrl,
			&post.Visibility, &post.CreatedAt, &post.Username, &post.AvatarUrl); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &posts, nil
}

func GetPostofAUser(postID int64) (*ViewPostDTO, error) {
	query := `SELECT
		p.id,
		p.user_id,
		p.caption,
		p.content_url,
		p.visibility,
		p.created_at,
		u.username,
		u.avatar_url
	FROM posts p
	JOIN users u ON p.user_id = u.id
	WHERE p.id = ?`
	row := databases.DB.QueryRow(query, postID)

	var post ViewPostDTO
	err := row.Scan(&post.ID, &post.UserID, &post.Caption, &post.ContentUrl,
		&post.Visibility, &post.CreatedAt, &post.Username, &post.AvatarUrl)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func GetPostForOperation(postID, uID int64) (*Post, error) {
	query := `SELECT
		id,
		user_id,
		caption,
		content_url,
		visibility
	FROM posts
	WHERE id = ? AND user_id = ?`
	row := databases.DB.QueryRow(query, postID, uID)

	var post Post
	err := row.Scan(&post.ID, &post.UserID, &post.Caption, &post.ContentUrl, &post.Visibility)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func GetPostOwnerID(postID int64) (int64, error) {
	query := `SELECT user_id FROM posts WHERE id = ?`
	row := databases.DB.QueryRow(query, postID)

	var userID int64
	if err := row.Scan(&userID); err != nil {
		return 0, err
	}

	return userID, nil
}

func GetPostVisibility(postID int64) (*string, int64, error) {
	query := `SELECT user_id, visibility FROM posts WHERE id = ?`
	row := databases.DB.QueryRow(query, postID)

	var visibility string
	var postUserID int64
	err := row.Scan(&postUserID, &visibility)
	if err != nil {
		return nil, 0, err
	}

	return &visibility, postUserID, nil
}

func GetTotalPostByUID(uID int64) (int, error) {
	query := `SELECT COUNT(*) FROM posts WHERE user_id = ?`
	row := databases.DB.QueryRow(query, uID)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (p *Post) EditPost() error {
	query := `UPDATE posts SET
		caption = ?,
		content_url = ?,
		visibility = ?
	WHERE id = ? AND user_id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(p.Caption, p.ContentUrl, p.Visibility, p.ID, p.UserID)
	if err != nil {
		return err
	}
	return nil
}

func DeletePost(postID, uID int64) error {
	query := `DELETE FROM posts WHERE id = ? AND user_id = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(postID, uID)
	if err != nil {
		return err
	}
	return nil
}
