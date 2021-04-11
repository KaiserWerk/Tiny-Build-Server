package databaseService

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type databaseService struct {
	db *gorm.DB
}



func New() *databaseService {
	config := global.GetConfiguration()

	var driver gorm.Dialector = mysql.Open(config.Database.DSN)
	if config.Database.Driver == "sqlite" {
		driver = sqlite.Open(config.Database.DSN)
	}

	db, err := gorm.Open(driver, &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	err = db.AutoMigrate(
		&entity.AdminSetting{},
		&entity.BuildDefinition{},
		&entity.BuildExecution{},
		&entity.User{},
		&entity.UserAction{},
		&entity.UserVariable{})
	if err != nil {
		panic("AutoMigrate panic: " + err.Error())
	}

	return &databaseService{db: db}
}

func (ds databaseService) Quit() {
	ds.Quit()
}

func (ds databaseService) RowExists(query string, args ...interface{}) bool {
	exists := true

	result := ds.db.Exec(fmt.Sprintf("SELECT exists (%s)", query), args...)
	if result.Error != nil {
		exists = false
	}

	return exists
}


