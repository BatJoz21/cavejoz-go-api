package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretkey = os.Getenv("secretkey")

func GenerateAccessToken(uID int64, username, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uID":      uID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
	})

	return token.SignedString([]byte(secretkey))
}

func GenerateRefreshToken(uID int64) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uID": uID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	return refreshToken.SignedString([]byte(secretkey))
}

func VerifyToken(token string) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return []byte(secretkey), nil
	})
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New("Invalid token")
	}

	return parsedToken, nil
}

func GetClaimsData(parsedToken *jwt.Token) (int64, string, string, error) {
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", "", errors.New("Invalid token claims")
	}

	uID := int64(claims["uID"].(float64))
	username := claims["username"].(string)
	role := claims["role"].(string)

	return uID, username, role, nil
}
