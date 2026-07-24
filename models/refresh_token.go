package models

import (
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type RefreshToken struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	DeviceName *string    `json:"device_name"`
	TokenHash  string     `json:"token_hash"`
	ExpiresAt  *time.Time `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	CreatedAt  *time.Time `json:"created_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func GetRefreshTokenByHashedToken(hashed string) (*RefreshToken, error) {
	query := `SELECT
		id,
		user_id,
		device_name,
		token_hash,
		expires_at,
		revoked_at,
		created_at
	FROM refresh_tokens WHERE token_hash = ?`
	row := databases.DB.QueryRow(query, hashed)

	var data RefreshToken
	err := row.Scan(&data.ID, &data.UserID, &data.DeviceName, &data.TokenHash,
		&data.ExpiresAt, &data.RevokedAt, &data.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *RefreshToken) StoreRefreshToken() error {
	query := `INSERT INTO refresh_tokens(user_id, device_name, token_hash, expires_at)
		VALUES (?, ?, ?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(r.UserID, r.DeviceName, r.TokenHash, r.ExpiresAt)
	if err != nil {
		return err
	}

	r.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (r *RefreshToken) RevokeRefreshToken() error {
	query := `UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ?`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), r.TokenHash)

	return err
}
