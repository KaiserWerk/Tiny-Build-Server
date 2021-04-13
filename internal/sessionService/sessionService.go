package sessionService

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/sessionstore"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

func GetUserFromSession(s sessionstore.Session) (entity.User, error) {
	userIdStr, ok := s.GetVar("user_id")
	if !ok {
		return entity.User{}, nil
	}

	userId, _ := strconv.Atoi(userIdStr)

	ds := databaseService.New()
	//defer ds.Quit()
	user, err := ds.GetUserById(userId)
	return user, err
}

