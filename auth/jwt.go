package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ClaimsWithUserType struct {
	jwt.RegisteredClaims
	UserType string `json:"usertype"`
}

var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

func MakeJWT(publicID string, userType string, tokenSecret string, expirationMinutes int) (string, error) { // returns JWT as string and error
	expiresIn := time.Duration(expirationMinutes) * time.Minute
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, ClaimsWithUserType{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "goqueue",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   publicID,
		},
		UserType: userType,
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (string, string, error) { //returns public ID, type (visitor/user) and error
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
		return "", "", err
	}

	// token checks
	// check if token is valid
	if !token.Valid {
		return "", "", fmt.Errorf("token is invalid")
	}
	// check if token is expired
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return "", "", jwt.ErrTokenExpired
	}
	// check if token is issued in the future
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(time.Now()) {
		return "", "", jwt.ErrTokenUsedBeforeIssued
	}
	// check if token is issued by the correct issuer
	if claims.Issuer != "goqueue" {
		return "", "", jwt.ErrTokenInvalidIssuer
	}

	// return ID
	return claims.Subject, claims.UserType, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return auth, fmt.Errorf("no authorization header value found")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil

}
