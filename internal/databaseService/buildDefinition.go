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
