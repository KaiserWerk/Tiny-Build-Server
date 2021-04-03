package databaseService

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

func (ds databaseService) GetUsernameById(id int) string {
	var u entity.User

	result := ds.db.First(&u, id)
	if result.Error != nil {
		return ""
	}

	return u.Displayname
}

func (ds databaseService) GetUserByEmail(email string) (entity.User, error) {
	var u entity.User
	result := ds.db.First(&u, "email = ?", email)
	if result.Error != nil {
		return entity.User{}, result.Error
	}
	return u, nil
}

func (ds databaseService) GetUserById(id int) (entity.User, error) {
	var u entity.User
	result := ds.db.First(&u, id)
	if result.Error != nil {
		return u, result.Error
	}

	return u, nil
}
