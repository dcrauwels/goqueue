package strutils

import (
	"errors"
	"net/mail"
	"unicode"
)

func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 30 {
		return errors.New("password length invalid")
	}
	for _, r := range password {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return errors.New("password contains invalid (non-alphanumeric) characters")
		}
	}
	return nil
}
