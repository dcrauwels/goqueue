package strutils

import (
	"database/sql"
	"errors"
	"net/http"
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
	ErrValueNegative := errors.New("negative or null environment values are not allowed")

	envVar, ok := os.LookupEnv(s)
	if !ok { // this means there was no value found for keystring s
		return r, ErrNoValueFound
	}

	r, err := strconv.Atoi(envVar)
	if err != nil { // this means the value passed into Atoi cannot be converted into an integer - i.e. it contains non-numeric characters
		return r, ErrValueNotNumeric
	} else if r <= 0 {
		return r, ErrValueNegative
	}

	return r, nil
}

func GetPublicIDFromPathValue(path string, publicIDLength int, r *http.Request) (string, error) {
	// used for retrieving public IDs from path values
	// e.g. the value for 'user_public_id' in GET /api/users/{user_public_id}
	// checks if the public ID provided in the path is of the correct length as specified in the .env config file

	ErrIncorrectPublicIDLength := errors.New("path value public ID has incorrect length")

	result := r.PathValue(path)
	if len(result) != publicIDLength {
		return "", ErrIncorrectPublicIDLength
	} else {
		return result, nil
	}
}

func QueryParameterToNullString(s string) sql.NullString {
	// used to convert strings retrieved from query parameters (e.g. through r.URL.Query().Get()) to sql.NullStrings.
	// Empty query parameters are returned by r.URL.Query().Get() as "", which is why the Valid field uses the logic below.
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
