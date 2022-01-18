package databaseservice

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
)

// GetAvailableVariablesForUser determines all available variables for a user by the given Id
func (ds *DatabaseService) GetAvailableVariablesForUser(userId uint) ([]entity.UserVariable, error) {
	variables := make([]entity.UserVariable, 0)
	result := ds.db.Find(&variables).Where("user_entry_id = ? OR public = 1", userId)
	if result.Error != nil {
		return nil, result.Error
	}

	return variables, nil
}

func (ds *DatabaseService) AddVariable(userVar entity.UserVariable) (uint, error) {
	result := ds.db.Create(&userVar)
	if result.Error != nil {
		return 0, result.Error
	}

	return userVar.ID, nil
}

func (ds *DatabaseService) GetVariable(id int) (entity.UserVariable, error) {
	var uv entity.UserVariable
	result := ds.db.First(&uv, id)
	if result.Error != nil {
		return uv, result.Error
	}

	return uv, nil
}

func (ds *DatabaseService) FindVariable(cond string, args ...interface{}) (entity.UserVariable, error) {
	var userVar entity.UserVariable
	result := ds.db.Where(cond, args...).Find(&userVar)
	if result.Error != nil {
		return entity.UserVariable{}, result.Error
	}

	if result.RowsAffected == 0 {
		return entity.UserVariable{}, fmt.Errorf("no variable found")
	}

	return userVar, nil
}

func (ds *DatabaseService) UpdateVariable(userVar entity.UserVariable) error {
	result := ds.db.Save(&userVar)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (ds *DatabaseService) DeleteVariable(id uint) error {
	result := ds.db.Delete(&entity.UserVariable{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
