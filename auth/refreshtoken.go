package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
)

var ErrRefreshTokenInvalid = errors.New("could not revoke old refresh token as it was not found in database")

func MakeRefreshToken() (string, error) {
	// get the hex
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	// hex to string
	encodedKey := hex.EncodeToString(key)
	return encodedKey, nil
}

func RotateRefreshToken(db databaseQueryer, w http.ResponseWriter, r *http.Request, oldRefreshTokenCookie *http.Cookie) (database.RefreshToken, error) {
	/*
		Function to rotate (revoke, then create anew) a refresh token.
	*/

	// 1. init
	result := database.RefreshToken{}

	// 2. get and revoke old refresh token from cookie
	oldRefreshToken, err := db.RevokeRefreshTokenByToken(r.Context(), oldRefreshTokenCookie.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, ErrRefreshTokenInvalid, "unable to revoke token: not found in database.")
			return result, err
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database(RevokeRefreshTokenByToken in RotateRefreshToken)")
			return result, err
		}
	}

	// 3. make new refresh token
	newRefreshToken, err := MakeRefreshToken()
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error creating refresh token (in MakeRefreshToken in RotateRefreshToken)")
		return result, err
	}

	// 4. query database to create new refresh token
	rtParams := database.CreateRefreshTokenParams{
		Token:        newRefreshToken,
		UserPublicID: oldRefreshToken.UserPublicID,
	}
	result, err = db.CreateRefreshToken(r.Context(), rtParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (CreateRefreshToken in makeAuthMiddleware)")
		return result, err
	}

	//5. return
	return result, nil
}
