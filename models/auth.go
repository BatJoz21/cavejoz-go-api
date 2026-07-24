package models

import (
	"errors"

	"github.com/BatJoz21/cavejoz-go-api/databases"
	"github.com/BatJoz21/cavejoz-go-api/utils"
)

func ValidateCredentials(u *UserLoginDTO) error {
	// Get password from database
	query := `SELECT password_hash FROM users WHERE email = ?`
	row := databases.DB.QueryRow(query, u.Email)
	var retreivedPassword string
	if err := row.Scan(&retreivedPassword); err != nil {
		return err
	}

	// Compare user input's password with password from database
	if !utils.CheckPasswordHash(u.Password, retreivedPassword) {
		return errors.New("Invalid credentials")
	}

	return nil
}

func GetUserDataByEmail(email string) (*SessionDataDTO, error) {
	// Get user data from database for login purpose
	query := `SELECT id, username, role FROM users WHERE email = ?`
	row := databases.DB.QueryRow(query, email)

	var data SessionDataDTO
	if err := row.Scan(&data.ID, &data.Username, &data.Role); err != nil {
		return nil, err
	}

	return &data, nil
}

func GetUserDataByUID(uID int64) (*SessionDataDTO, error) {
	// Get user data from database for refresh access token purpose
	query := `SELECT id, username, role FROM users WHERE id = ?`
	row := databases.DB.QueryRow(query, uID)

	var data SessionDataDTO
	if err := row.Scan(&data.ID, &data.Username, &data.Role); err != nil {
		return nil, err
	}

	return &data, nil
}
