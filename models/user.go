package models

import "github.com/BatJoz21/cavejoz-go-api/databases"

type User struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"-"`
	FullName     string  `json:"full_name"`
	Bio          *string `json:"bio"`
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

func GetUserDataForOps(uID int64) (*User, error) {
	query := `SELECT
		username,
		email,
		full_name,
		bio,
		avatar_url
	FROM users WHERE id = ?`
	row := databases.DB.QueryRow(query, uID)

	var u User
	err := row.Scan(&u.Username, &u.Email, &u.FullName, &u.Bio, &u.AvatarUrl)
	if err != nil {
		return nil, err
	}

	u.ID = uID
	return &u, nil
}

func GetUserDataByUsername(username string) (*UserProfileDTO, error) {
	query := `SELECT
		id,
		username,
		full_name,
		bio,
		avatar_url
	FROM users WHERE username = ?`
	row := databases.DB.QueryRow(query, username)

	var u UserProfileDTO
	err := row.Scan(&u.ID, &u.Username, &u.FullName, &u.Bio, &u.AvatarUrl)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func GetStoredAvatarPath(uID int64) (*string, error) {
	query := `SELECT avatar_url FROM users WHERE id = ?`
	row := databases.DB.QueryRow(query, uID)

	var path string
	err := row.Scan(&path)
	if err != nil {
		return nil, err
	}

	return &path, nil
}

func (u *User) Update() error {
	query := `UPDATE users SET
		username = ?,
		full_name = ?,
		bio = ?,
		avatar_url = ?
	WHERE id = ?`
	_, err := databases.DB.Exec(query, u.Username, u.FullName, u.Bio, u.AvatarUrl, u.ID)

	if err != nil {
		return err
	}

	return nil
}
