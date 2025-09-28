package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ClaimsWithUserType struct {
	jwt.RegisteredClaims
	UserType string `json:"usertype"`
}

var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

func MakeJWT(ID uuid.UUID, userType string, tokenSecret string, expirationMinutes int) (string, error) {
	expiresIn := time.Duration(expirationMinutes) * time.Minute
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, ClaimsWithUserType{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "goqueue",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   ID.String(),
		},
		UserType: userType,
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, string, error) { //returns ID, type (visitor/user) and error
	// define claims to unpack into and keyfunc
	claims := &ClaimsWithUserType{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		//check if signed with HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(tokenSecret), nil
	}

	// parse the token
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return uuid.Nil, "", err
	}

	// token checks
	// check if token is valid
	if !token.Valid {
		return uuid.Nil, "", fmt.Errorf("token is invalid")
	}
	// check if token is expired
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return uuid.Nil, "", jwt.ErrTokenExpired
	}
	// check if token is issued in the future
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(time.Now()) {
		return uuid.Nil, "", jwt.ErrTokenUsedBeforeIssued
	}
	// check if token is issued by the correct issuer
	if claims.Issuer != "goqueue" {
		return uuid.Nil, "", jwt.ErrTokenInvalidIssuer
	}

	// get ID from token
	ID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", jwt.ErrTokenInvalidId
	}

	// return ID
	return ID, claims.UserType, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return auth, fmt.Errorf("no authorization header value found")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil

}
