package databaseService

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/gorm"
)

func (ds databaseService) GetNewestBuildExecutions(limit int) ([]entity.BuildExecution, error) {
	beList := make([]entity.BuildExecution, 0)
	var result *gorm.DB
	if limit > 0 {
		result = ds.db.Limit(limit).Find(&beList).Order("executed_at DESC")
	} else {
		result = ds.db.Find(&beList).Order("executed_at DESC")
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

func (ds databaseService) AddBuildExecution(be entity.BuildExecution) error {
	result := ds.db.Create(&be)
	if result.Error != nil {
		return result.Error
	}
	return nil
}