package databaseservice

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

	"gorm.io/gorm"
)

// GetNewestBuildExecutions fetches the newest build executions
func (ds *DatabaseService) GetNewestBuildExecutions(limit int, query string, args ...interface{}) ([]entity.BuildExecution, error) {
	beList := make([]entity.BuildExecution, 0)
	var result *gorm.DB
	if limit > 0 {
		if query != "" {
			result = ds.db.Where(query, args).Limit(limit).Order("executed_at desc").Find(&beList)
		} else {
			result = ds.db.Limit(limit).Order("executed_at desc").Find(&beList)
		}

	} else {
		if query != "" {
			result = ds.db.Where(query, args).Order("executed_at desc").Find(&beList)
		} else {
			result = ds.db.Order("executed_at desc").Find(&beList)
		}
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return beList, nil
}

// GetBuildExecutionById fetches a specific build execution by id
func (ds *DatabaseService) GetBuildExecutionById(id int) (entity.BuildExecution, error) {
	var be entity.BuildExecution
	result := ds.db.First(&be, id)
	if result.Error != nil {
		return entity.BuildExecution{}, result.Error
	}
	return be, nil
}

// FindBuildExecutions finds build executions by criteria
func (ds *DatabaseService) FindBuildExecutions(query interface{}, args ...interface{}) ([]entity.BuildExecution, error) {
	executions := make([]entity.BuildExecution, 0)
	result := ds.db.Where(query, args...).Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	return executions, nil
}

// AddBuildExecution adds a new build execution
func (ds *DatabaseService) AddBuildExecution(be entity.BuildExecution) error {
	result := ds.db.Create(&be)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
