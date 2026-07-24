package models

import "github.com/BatJoz21/cavejoz-go-api/databases"

type User struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"`
	FullName     string  `json:"full_name"`
	Bio          string  `json:"bio"`
	Role         string  `json:"role"`
	AvatarUrl    *string `json:"avatar_url"`
}

func (u *User) Save() error {
	query := `INSERT INTO users(username, email, password_hash, full_name, role, avatar_url)
		VALUES (?, ?, ?, ?, ?, ?)`
	stmt, err := databases.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(u.Username, u.Email, u.PasswordHash,
		u.FullName, u.Role, u.AvatarUrl)
	if err != nil {
		return err
	}

	u.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}
