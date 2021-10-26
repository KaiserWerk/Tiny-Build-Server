package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"

	"github.com/KaiserWerk/sessionstore"
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
func CheckLogin(r *http.Request) (sessionstore.Session, error) {
	sessMgr := global.GetSessionManager()
	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		return sessionstore.Session{}, errors.New("could not get cookie: " + err.Error())
	}

	session, err := sessMgr.GetSession(sessId)
	if err != nil {
		return sessionstore.Session{}, errors.New("could not get session: " + err.Error())
	}

	return session, nil
}
