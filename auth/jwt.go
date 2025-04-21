package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	const expiresIn = time.Hour
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// define claims to unpack into and keyfunc
	claims := &jwt.RegisteredClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		//check if signed with HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	}

	// parse the token
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		fmt.Println(tokenString)
		return uuid.Nil, err
	}

	// token checks
	// check if token is valid
	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	// check if token is expired
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return uuid.Nil, fmt.Errorf("token expired")
	}
	// check if token is issued in the future
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(time.Now()) {
		return uuid.Nil, fmt.Errorf("token issued in the future")
	}
	// check if token is issued by the correct issuer
	if claims.Issuer != "chirpy" {
		return uuid.Nil, fmt.Errorf("token issued by incorrect issuer")
	}

	// get userID from token
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid userID in token: %v", err)
	}
	// return userID
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return auth, fmt.Errorf("no authorization header value found")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil

}
