package validation

import (
	"errors"
	"strings"
)

func ValidateRegister(email string, username string, password string) error {
	email = strings.TrimSpace(email)
	username = strings.TrimSpace(username)

	if email == "" {
		return errors.New("email is required")
	}

	if username == "" {
		return errors.New("username is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	return nil
}

func ValidateLogin(email string, password string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return errors.New("email is required")
	}

	if password == "" {
		return errors.New("password is required")
	}

	return nil
}

func ValidateRefresh(token string) error {
	if strings.TrimSpace(token) == "" {
		return errors.New("refresh token is required")
	}

	return nil
}
