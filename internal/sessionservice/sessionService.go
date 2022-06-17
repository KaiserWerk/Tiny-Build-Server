package sessionservice

import (
	"fmt"
	"strconv"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

	"github.com/KaiserWerk/sessionstore/v2"
	_ "github.com/go-sql-driver/mysql"
)

// NewSessionManager creates a new session manager
func NewSessionManager(name string) *sessionstore.SessionManager {
	return sessionstore.NewManager(name)
}

// GetUserFromSession returns a user from a given session, if possible
func GetUserFromSession(ds *databaseservice.DatabaseService, s *sessionstore.Session) (entity.User, error) {
	userIdStr, ok := s.GetVar("user_id")
	if !ok {
		return entity.User{}, fmt.Errorf("variable key user_id not found")
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return entity.User{}, err
	}

	//defer ds.Quit()
	user, err := ds.GetUserById(uint(userId))
	return user, err
}
