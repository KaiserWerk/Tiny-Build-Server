package dbservice

import (
	"fmt"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
)

// GetAllUsers fetches a list of all users
func (ds *DBService) GetAllUsers() ([]entity.User, error) {
	users := make([]entity.User, 0)
	result := ds.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetUsernameById fetches a username by Id
// Supposed to be used in templates
func (ds *DBService) GetUsernameById(id int) string {
	var u entity.User

	result := ds.db.First(&u, id)
	if result.Error != nil {
		return ""
	}

	return u.DisplayName
}

// GetUserByEmail fetches a user by email address
func (ds *DBService) GetUserByEmail(email string) (entity.User, error) {
	var u entity.User
	result := ds.db.First(&u, "email = ?", email)
	if result.Error != nil {
		return entity.User{}, result.Error
	}
	return u, nil
}

// GetUserById fetches a user by Id
func (ds *DBService) GetUserById(id uint) (entity.User, error) {
	var u entity.User
	result := ds.db.First(&u, id)
	if result.Error != nil {
		return u, result.Error
	}

	return u, nil
}

// AddUser adds a new user
func (ds *DBService) AddUser(user entity.User) (uint, error) {
	result := ds.db.Create(&user)
	if result.Error != nil {
		return 0, result.Error
	}

	return user.ID, nil
}

// UpdateUser alters an existing user
func (ds *DBService) UpdateUser(user entity.User) error {
	result := ds.db.Save(&user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// FindUser finds a user by the supplied criteria
func (ds *DBService) FindUser(cond string, args ...interface{}) (entity.User, error) {
	var user entity.User
	result := ds.db.Where(cond, args...).Find(&user)
	if result.Error != nil {
		return entity.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return entity.User{}, fmt.Errorf("no user found")
	}

	return user, nil
}

// DeleteUser removes a user by the given Id
func (ds *DBService) DeleteUser(id uint) error {
	result := ds.db.Delete(&entity.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
