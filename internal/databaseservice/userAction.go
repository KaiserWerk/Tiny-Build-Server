package databaseservice

import (
	"database/sql"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
)

// InsertUserAction adds a new user action
func (ds databaseService) InsertUserAction(userId int, purpose, token string, validity sql.NullTime) error {

	// 1. set all timely invalid actions to a null token
	result := ds.db.Exec("UPDATE user_actions SET token = NULL WHERE user_id = ? AND validity < NOW()", userId)
	if result.Error != nil {
		return result.Error
	}

	// 2. insert new token
	newAction := entity.UserAction{
		UserId:   userId,
		Purpose:  purpose,
		Token:    token,
		Validity: validity,
	}

	result = ds.db.Create(&newAction)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetUserActionByToken retrieves a specific user action by token
func (ds databaseService) GetUserActionByToken(token string) (entity.UserAction, error) {
	userAction := entity.UserAction{}
	result := ds.db.Find(&userAction, "token = ?", token)
	if result.Error != nil {
		return entity.UserAction{}, result.Error
	}

	return userAction, nil
}

// InvalidatePasswordResets invalidates all user action of type 'password_reset'
func (ds databaseService) InvalidatePasswordResets(userId int) error {
	result := ds.db.Exec("UPDATE user_action SET validity = ? WHERE purpose = 'password_reset' AND user_id = ?",
		sql.NullTime{}, userId)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// AddUserAction creates a given user action
func (ds databaseService) AddUserAction(action entity.UserAction) error {
	result := ds.db.Create(&action)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// UpdateUserAction updates a given user action
func (ds databaseService) UpdateUserAction(userAction entity.UserAction) error {
	result := ds.db.Save(&userAction)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
