package databaseService

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/gorm"
)

func (ds databaseService) GetNewestBuildExecutions(limit int) ([]entity.BuildExecution, error) {
	var beList []entity.BuildExecution

	var result *gorm.DB

	if limit > 0 {
		result = ds.db.Limit(limit).Find(&beList)
	} else {
		result = ds.db.Find(&beList)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return beList, nil
}
