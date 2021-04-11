package databaseService

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/gorm"
)

func (ds databaseService) GetNewestBuildExecutions(limit int) ([]entity.BuildExecution, error) {
	var beList []entity.BuildExecution

	var result *gorm.DB

	// TODO: ORDERING!
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

func (ds databaseService) GetBuildExecutionById(id int) (entity.BuildExecution, error) {
	var be entity.BuildExecution
	result := ds.db.First(&be, id)
	if result.Error != nil {
		return entity.BuildExecution{}, result.Error
	}
	
	return be, nil
}

func (ds databaseService) FindBuildExecutions(query interface{}, args ...interface{}) ([]entity.BuildExecution, error) {
	executions := make([]entity.BuildExecution, 0)
	result := ds.db.Where(query, args).Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}

	return executions, nil
}
