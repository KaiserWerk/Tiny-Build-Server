package databaseService

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

func (ds databaseService) GetAllSettings() (map[string]string, error) {
	settings := make(map[string]string)

	var s []entity.AdminSetting

	result := ds.db.Find(&s)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, v := range s {
		settings[v.Name] = v.Value
	}

	return settings, nil
}

func (ds databaseService) SetSetting(name, value string) error {

	result := ds.db.Model(&entity.AdminSetting{}).Where(name + " = ?", name).Update("value", value)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
