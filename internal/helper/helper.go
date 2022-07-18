package helper

import (
	"errors"
	"net/http"
	"os"
	"time"
)

// FileExists checks whether a given file exists or not
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return !info.IsDir()
}

// GetHeaderIfSet returns the value of a given headers, if it exists
func GetHeaderIfSet(r *http.Request, key string) (string, error) {
	header := r.Header.Get(key)
	if header == "" {
		return "", errors.New("header is not set or empty")
	}
	return header, nil
}

// FormatDate formats a time.Time into a string
// To be used in templates
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
