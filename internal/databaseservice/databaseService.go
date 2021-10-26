package databaseservice

import (
	"fmt"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type databaseService struct {
	db *gorm.DB
}

// New returns a new database service connection
func New() *databaseService {
	config := global.GetConfiguration()

	var driver gorm.Dialector = mysql.Open(config.Database.DSN)
	if config.Database.Driver == "sqlite" {
		driver = sqlite.Open(config.Database.DSN)
	}

	db, err := gorm.Open(driver, &gorm.Config{
		PrepareStmt: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			NoLowerCase:   false,
		},
	})
	if err != nil {
		panic("gorm connection error: " + err.Error())
	}

	return &databaseService{db: db}
}

// AutoMigrate makes sure the database tables exist, corresponding
// to the supplied structs
func (ds databaseService) AutoMigrate() error {
	err := ds.db.AutoMigrate(
		&entity.AdminSetting{},
		&entity.BuildDefinition{},
		&entity.BuildExecution{},
		&entity.User{},
		&entity.UserAction{},
		&entity.UserVariable{})
	if err != nil {
		return err
	}

	return nil
}

// Quit ends the database connection
func (ds databaseService) Quit() {
	ds.Quit()
}

// RowExists takes an SQL query and return true, if at least one entry
// exists for the given query
func (ds databaseService) RowExists(query string, args ...interface{}) bool {
	exists := true

	result := ds.db.Exec(fmt.Sprintf("SELECT exists (%s)", query), args...)
	if result.Error == nil {
		exists = false
	}

	return exists
}
