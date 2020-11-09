package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/KaiserWerk/Tiny-Build-Server/internal"
	"github.com/KaiserWerk/sessionstore"
	"golang.org/x/crypto/bcrypt"
)

func GenerateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func HashString(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func DoesHashMatch(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckLogin(r *http.Request) (sessionstore.Session, error) {
	sessMgr := internal.GetSessionManager()
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
