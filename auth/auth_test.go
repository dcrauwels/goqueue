package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestPassword(t *testing.T) {
	// preliminary setup
	passwordQp := "qqpp1001"
	hashedPasswordQp, err := HashPassword(passwordQp)
	if err != nil {
		t.Errorf(`HashPassword("qqpp1001") = %s, %v; expected hash, nil`, hashedPasswordQp, err)
	}

	passwordZa := "zasxzasx"
	hashedPasswordZa, err := HashPassword(passwordZa)
	if err != nil {
		t.Errorf(`HashPassword("zasxzasx") = %s, %v; expected hash, nil`, hashedPasswordZa, err)
	}

	// now we check the passwords
	// straight matches
	if CheckPasswordHash(hashedPasswordQp, passwordQp) != nil {
		t.Errorf(`CheckPasswordHash(hashedPasswordQp, "qqpp1001") = %v, expected nil`, err)
	}
	if CheckPasswordHash(hashedPasswordZa, passwordZa) != nil {
		t.Errorf(`CheckPasswordHash(hashedPasswordZa, "zasxzasx") = %v; expected nil`, err)
	}
	// corss
	if CheckPasswordHash(hashedPasswordQp, passwordZa) == nil {
		t.Errorf(`CheckPasswordHash(hashedPasswordQp, "zasxzasx") = nil; expected err`)
	}

}

func TestUserJWT(t *testing.T) {
	// arguments
	userID := uuid.New()
	tokenSecret := "qqpp1001"

	// make jwt
	jwt, err := MakeJWT(userID, tokenSecret, 60)
	if err != nil {
		t.Errorf(`MakeJWT(userID, "qqpp1001", time.Second) = %s, %v; expected token, nil`, jwt, err)
	}

	// straight validation
	validatedID, err := ValidateJWT(jwt, tokenSecret)
	if err != nil {
		t.Errorf(`ValidateJWT(jwt, "qqpp1001") = %s, %v; expected UUID, nil`, validatedID.String(), err)
	}
	if userID.String() != validatedID.String() {
		t.Errorf(`userID.String() == validatedID.String() returns false; expected true`)
	}

	// wrong tokenSecret
	wrongID, err := ValidateJWT(jwt, "zasxzasx")
	if err == nil {
		t.Errorf(`ValidateJWT(jwt, "zasxzasx") = %v, %v; expected uuid.Nil, err`, wrongID, err)
	}
}
