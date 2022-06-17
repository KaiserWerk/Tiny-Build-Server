package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/KaiserWerk/sessionstore/v2"
	"golang.org/x/crypto/bcrypt"
)

// GenerateToken generates a token with a given length
func GenerateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// HashString returns the bcrypt hash for a given string
func HashString(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// DoesHashMatch checks whether a given password and hash match
func DoesHashMatch(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CheckLogin checks if a valid sessions exists in an *http.Request
func CheckLogin(sessMgr *sessionstore.SessionManager, r *http.Request) (*sessionstore.Session, error) {
	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		return nil, fmt.Errorf("could not get cookie: %s", err.Error())
	}

	session, err := sessMgr.GetSession(sessId)
	if err != nil {
		return nil, fmt.Errorf("could not get session: %s", err.Error())
	}

	return session, nil
}
