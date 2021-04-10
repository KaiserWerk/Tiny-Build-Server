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

func (ds databaseService) AddUser(user entity.User) (int, error) {
	result := ds.db.Create(&user)
	if result.Error != nil {
		return 0, result.Error
	}

	return user.Id, nil
}

func (ds databaseService) UpdateUser(user entity.User) error {
	result := ds.db.Save(&user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
