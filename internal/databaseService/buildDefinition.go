package databaseService

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/gorm"
)

func (ds databaseService) GetNewestBuildDefinitions(limit int) ([]entity.BuildDefinition, error) {

	var bdList []entity.BuildDefinition

	var result *gorm.DB
	if limit > 0 {
		result = ds.db.Limit(limit).Find(&bdList)
	} else {
		result = ds.db.Find(&bdList)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return bdList, nil
}

func (ds databaseService) GetBuildDefinitionById(id int) (entity.BuildDefinition, error) {
	var buildDefinition entity.BuildDefinition
	result := ds.db.First(&buildDefinition, id)
	if result.Error != nil {
		return entity.BuildDefinition{}, result.Error
	}
	return buildDefinition, nil
}

func (ds databaseService) GetBuildDefCaption(id int) (string, error) {
	var bd entity.BuildDefinition
	result := ds.db.First(&bd, id)
	if result.Error != nil {
		return "", result.Error
	}
	return bd.Caption, nil
}