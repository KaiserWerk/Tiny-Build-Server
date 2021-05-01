package databaseservice

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
)

// GetAllSettings fetches all settings
func (ds databaseService) GetAllSettings() (map[string]string, error) {
	settings := make(map[string]string)

	var s []entity.AdminSetting

	result := ds.db.Find(&s)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, v := range s {
		settings[v.Name] = v.Value
	} // das geht schöner

	return settings, nil
}

// SetSetting sets a new value for a given setting
func (ds databaseService) SetSetting(name, value string) error {
	//result := ds.db.Model(&entity.AdminSetting{}).Where("name = ?", name).Update("value", value)
	var setting entity.AdminSetting
	found := ds.db.Find(&setting, "name = ?", name)
	if found.Error != nil {
		return found.Error
	}

	setting = entity.AdminSetting{
		Name:  name,
		Value: value,
	}

	if found.RowsAffected == 0 {
		// not found, inserting
		insertResult := ds.db.Create(&setting)
		if insertResult.Error != nil {
			return insertResult.Error
		}
	} else {
		// found, updating
		updateResult := ds.db.Model(&entity.AdminSetting{}).Where("name = ?", name).Update("value", value)
		if updateResult.Error != nil {
			return updateResult.Error
		}
	}

	return nil
}
