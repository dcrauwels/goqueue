package strutils

import (
	"database/sql"
	"errors"
	"net/mail"
	"os"
	"strconv"
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

func InitNullString(s string) sql.NullString {
	// shortform for initializing a nullstring. relevant when querying a DB with a nullable string as parameter
	r := sql.NullString{
		String: s,
		Valid:  true,
	}
	return r
}

func GetIntegerEnvironmentVariable(s string) (int, error) {
	/* Retrieves an environment value using os.LookupEnv, then converts it to an integer. Uses os.LookupEnv and */
	var r int
	ErrNoValueFound := errors.New("no environment value found for this key")
	ErrValueNotNumeric := errors.New("the environment for this key cannot be converted to an integer")

	envVar, ok := os.LookupEnv(s)
	if !ok { // this means there was no value found for keystring s
		return r, ErrNoValueFound
	}

	r, err := strconv.Atoi(envVar)
	if err != nil { // this means the value passed into Atoi cannot be converted into an integer - i.e. it contains non-numeric characters
		return r, ErrValueNotNumeric
	}

	return r, nil
}
