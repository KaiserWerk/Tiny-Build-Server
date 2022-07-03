package helper

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

	"gopkg.in/yaml.v3"
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

// UnmarshalBuildDefinitionContent unmarshals a build definition content
func UnmarshalBuildDefinitionContent(content string, bdContent *entity.BuildDefinitionContent) error {
	return yaml.Unmarshal([]byte(content), &bdContent)
}

func ReplaceVariables(content *string, variables []entity.UserVariable) {
	for _, v := range variables {
		*content = strings.ReplaceAll(*content, fmt.Sprintf("${%s}", v.Variable), v.Value)
	}
}
