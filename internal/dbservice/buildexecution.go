package dbservice

import (
	"gorm.io/gorm"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
)

// GetNewestBuildExecutions fetches the newest build executions
func (ds *DBService) GetNewestBuildExecutions(limit int, query string, args ...interface{}) ([]entity.BuildExecution, error) {
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
func (ds *DBService) GetBuildExecutionById(id int) (entity.BuildExecution, error) {
	var be entity.BuildExecution
	result := ds.db.First(&be, id)
	if result.Error != nil {
		return entity.BuildExecution{}, result.Error
	}
	return be, nil
}

// FindBuildExecutions finds build executions by criteria
func (ds *DBService) FindBuildExecutions(query interface{}, args ...interface{}) ([]entity.BuildExecution, error) {
	executions := make([]entity.BuildExecution, 0)
	result := ds.db.Where(query, args...).Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	return executions, nil
}

// AddBuildExecution adds a new build execution
func (ds *DBService) AddBuildExecution(be *entity.BuildExecution) error {
	result := ds.db.Create(be)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateBuildExecution alters an existing user
func (ds *DBService) UpdateBuildExecution(be *entity.BuildExecution) error {
	return ds.db.Save(be).Error
}
