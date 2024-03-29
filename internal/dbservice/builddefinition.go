package dbservice

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/gorm"
)

// GetNewestBuildDefinitions fetches the most recently edited or added build definitions
func (ds *DBService) GetNewestBuildDefinitions(limit int) ([]entity.BuildDefinition, error) {

	var bdList []entity.BuildDefinition

	var result *gorm.DB
	if limit > 0 {
		result = ds.db.Limit(limit).Find(&bdList, "deleted = 0")
	} else {
		result = ds.db.Find(&bdList, "deleted = 0")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return bdList, nil
}

// GetAllBuildDefinitions fetches all build definitions
func (ds *DBService) GetAllBuildDefinitions() ([]entity.BuildDefinition, error) {
	bds := make([]entity.BuildDefinition, 0)
	result := ds.db.Find(&bds, "deleted = 0")
	if result.Error != nil {
		return nil, result.Error
	}

	return bds, nil
}

// FindBuildDefinition looks for a specific build definition
func (ds *DBService) FindBuildDefinition(cond string, args ...interface{}) (entity.BuildDefinition, error) {
	var bd entity.BuildDefinition
	result := ds.db.Where(cond+" AND deleted = 0", args...).First(&bd)
	if result.Error != nil {
		return bd, result.Error
	}

	return bd, nil
}

// GetBuildDefinitionById fetches a build definition by Id
func (ds *DBService) GetBuildDefinitionById(id uint) (entity.BuildDefinition, error) {
	var buildDefinition entity.BuildDefinition
	result := ds.db.First(&buildDefinition, id)
	if result.Error != nil {
		return entity.BuildDefinition{}, result.Error
	}
	return buildDefinition, nil
}

// GetBuildDefCaption fetches the caption of a given build definition id
// It is to be used in templates
func (ds *DBService) GetBuildDefCaption(id uint) (string, error) {
	var bd entity.BuildDefinition
	result := ds.db.First(&bd, id)
	if result.Error != nil {
		return "", result.Error
	}
	return bd.Caption, nil
}

// DeleteBuildDefinition removes a build definition
func (ds *DBService) DeleteBuildDefinition(bd *entity.BuildDefinition) error {
	bd.Deleted = true
	result := ds.db.Updates(bd)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// AddBuildDefinition adds a new build definition
func (ds *DBService) AddBuildDefinition(bd *entity.BuildDefinition) (uint, error) {
	bd.Deleted = false
	result := ds.db.Create(bd)
	if result.Error != nil {
		return 0, result.Error
	}

	return bd.ID, nil
}

// UpdateBuildDefinition updates a build definition
func (ds *DBService) UpdateBuildDefinition(bd *entity.BuildDefinition) error {
	result := ds.db.Updates(bd)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
