package databaseservice

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

// GetAvailableVariablesForUser determines all available variables for a user by the given Id
func (ds databaseService) GetAvailableVariablesForUser(userId int) ([]entity.UserVariable, error) {
	variables := make([]entity.UserVariable, 0)
	result := ds.db.Find(&variables).Where("user_entry_id = ? OR public = 1", userId)
	if result.Error != nil {
		return nil, result.Error
	}

	return variables, nil
}
