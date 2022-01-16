package databaseservice

import (
	"fmt"
	"sync"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DatabaseService struct {
	db *gorm.DB
}

var (
	dbs  DatabaseService
	once sync.Once
)

// Get returns a new database service connection
func Get() *DatabaseService {
	once.Do(func() {
		config := global.GetConfiguration()

		db, err := gorm.Open(mysql.Open(config.Database.DSN), &gorm.Config{
			PrepareStmt: true,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   false,
			},
		})
		if err != nil {
			panic("gorm connection error: " + err.Error())
		}

		conn, err := db.DB()
		if err != nil {
			panic("form DB() error: " + err.Error())
		}

		conn.SetMaxIdleConns(10)
		conn.SetMaxOpenConns(50)
		dbs = DatabaseService{db: db}
	})

	return &dbs
}

// AutoMigrate makes sure the database tables exist, corresponding
// to the supplied structs
func (ds DatabaseService) AutoMigrate() error {
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
func (ds DatabaseService) Quit() {
	ds.Quit()
}

// RowExists takes an SQL query and return true, if at least one entry
// exists for the given query
func (ds DatabaseService) RowExists(query string, args ...interface{}) bool {
	exists := true

	result := ds.db.Exec(fmt.Sprintf("SELECT exists (%s)", query), args...)
	if result.Error == nil {
		exists = false
	}

	return exists
}
