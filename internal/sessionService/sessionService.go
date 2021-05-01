package sessionService

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/sessionstore"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

// GetUserFromSession returns a user from a given session, if possible
func GetUserFromSession(s sessionstore.Session) (entity.User, error) {
	userIdStr, ok := s.GetVar("user_id")
	if !ok {
		return entity.User{}, fmt.Errorf("variable key user_id not found")
	}

	userId, _ := strconv.Atoi(userIdStr)

	ds := databaseService.New()
	//defer ds.Quit()
	user, err := ds.GetUserById(userId)
	return user, err
}
