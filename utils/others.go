package utils

import (
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

func IsEmailValid(email string) bool {
	if email == "" {
		return false
	}

	addr, err := mail.ParseAddress(email)

	return err == nil && addr.Address == email
}

func IsPasswordValid(password string) error {
	// Check if password is empty
	if password == "" {
		return errors.New("empty password")
	}

	// Check password length
	if len(password) < 8 || len(password) > 16 {
		return errors.New("missmatch length")
	}

	// Check if password has uppercase
	if !hasUpperCase(password) {
		return errors.New("must have an uppercase letter")
	}

	// Check if password has lowercase
	if !hasLowerCase(password) {
		return errors.New("must have an lowercase letter")
	}

	// Check if password has number
	if !hasDigit(password) {
		return errors.New("must have a number")
	}

	// Check if password has symbol
	if !hasSymbol(password) {
		return errors.New("must have a symbol")
	}

	return nil
}

func IsFullNameValid(fullName string) bool {
	if fullName == "" || strings.TrimSpace(fullName) == "" {
		return false
	}

	return true
}

func GetOffsetForPagination(max int, context *gin.Context) int {
	// Get page
	page, err := strconv.Atoi(context.DefaultQuery("page", "1"))
	if err != nil {
		return 0
	}
	if page < 1 {
		page = 1
	}

	// Count offset
	offset := max * (page - 1)
	return offset
}

func hasUpperCase(input string) bool {
	for _, r := range input {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}

func hasLowerCase(input string) bool {
	for _, r := range input {
		if unicode.IsLower(r) {
			return true
		}
	}

	return false
}

func hasDigit(input string) bool {
	for _, r := range input {
		if unicode.IsDigit(r) {
			return true
		}
	}

	return false
}

func hasSymbol(input string) bool {
	for _, r := range input {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return true
		}
	}

	return false
}
