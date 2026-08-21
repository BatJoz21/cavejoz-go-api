package utils

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretkey = sync.OnceValue(func() []byte {
	s := os.Getenv("secretkey")
	if len(s) < 32 {
		log.Fatal("Secret key is missing or shorter than 32 bytes")
	}
	return []byte(s)
})

// Token lifetimes. RefreshTokenTTL is used both for the JWT's own exp claim
// and for the expires_at column, so the two can never drift apart.
const (
	AccessTokenTTL  = time.Hour
	RefreshTokenTTL = time.Hour * 24 * 7

	// A spent refresh token presented again within this window is treated as a
	// client race (two requests refreshing at the same moment) rather than a
	// stolen token. Presented later than this, reuse means someone kept a copy.
	// Shrinking it tightens replay detection at the cost of spurious logouts
	// for chatty clients; widening it does the reverse.
	RefreshReuseGrace = 10 * time.Second
)

func GenerateAccessToken(uID int64, username, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uID":      uID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(AccessTokenTTL).Unix(),
	})

	return token.SignedString(secretkey())
}

func GenerateRefreshToken(uID int64) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uID": uID,
		"exp": time.Now().Add(RefreshTokenTTL).Unix(),
	})

	return refreshToken.SignedString(secretkey())
}

func VerifyToken(token string) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return secretkey(), nil
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

	uIDf, ok1 := claims["uID"].(float64)
	username, ok2 := claims["username"].(string)
	role, ok3 := claims["role"].(string)
	if !ok1 || !ok2 || !ok3 {
		return 0, "", "", errors.New("Invalid token claims")
	}

	uID := int64(uIDf)

	return uID, username, role, nil
}

// GetRefreshClaims reads the only claim a refresh token carries. It is
// separate from GetClaimsData because refresh tokens have no username or role,
// so asking for those would reject every valid refresh token.
func GetRefreshClaims(parsedToken *jwt.Token) (int64, error) {
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("Invalid token claims")
	}

	uIDf, ok := claims["uID"].(float64)
	if !ok {
		return 0, errors.New("Invalid token claims")
	}

	return int64(uIDf), nil
}
