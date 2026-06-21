package validation

import (
	"errors"
	"strings"
)

func ValidateNewUserData(email string, username string, password string) error {
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
