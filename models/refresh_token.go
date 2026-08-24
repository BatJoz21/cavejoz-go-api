package models

import (
	"errors"
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

// ErrRefreshTokenReuse means the presented token was already revoked by the
// time we tried to rotate it. Either it was stolen and replayed, or a client
// fired two refreshes at once. Both are handled the same way: end all sessions.
var ErrRefreshTokenReuse = errors.New("refresh token already used")

// RotateRefreshToken revokes the presented token and stores its replacement in
// one transaction, so a refresh can never leave both tokens live or neither.
//
// The UPDATE carries "revoked_at IS NULL" and the RowsAffected check acts as a
// compare-and-swap: if two concurrent refreshes present the same token, only
// one can win, and the loser gets ErrRefreshTokenReuse instead of a second
// valid session.
func RotateRefreshToken(oldHash, newHash string, userID int64, deviceName *string, expiresAt time.Time) error {
	tx, err := databases.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`UPDATE refresh_tokens SET revoked_at = ?
		WHERE token_hash = ? AND user_id = ? AND revoked_at IS NULL`,
		time.Now(), oldHash, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrRefreshTokenReuse
	}

	_, err = tx.Exec(`INSERT INTO refresh_tokens(user_id, device_name, token_hash, expires_at)
		VALUES (?, ?, ?, ?)`,
		userID, deviceName, newHash, expiresAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RevokeAllRefreshTokensByUserID ends every live session for a user. Called
// when a revoked token is presented again, since at that point we cannot tell
// which of the outstanding tokens is the attacker's.
func RevokeAllRefreshTokensByUserID(userID int64) error {
	_, err := databases.DB.Exec(`UPDATE refresh_tokens SET revoked_at = ?
		WHERE user_id = ? AND revoked_at IS NULL`,
		time.Now(), userID)

	return err
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
