package dbservice

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DBService struct {
	db *gorm.DB
}

func New(cfg *configuration.AppConfig) *DBService {
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{
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
	return &DBService{db: db}
}

// AutoMigrate makes sure the database tables exist, corresponding
// to the supplied structs
func (ds *DBService) AutoMigrate() error {
	return ds.db.AutoMigrate(
		&entity.AdminSetting{},
		&entity.BuildDefinition{},
		&entity.BuildExecution{},
		&entity.User{},
		&entity.UserAction{},
		&entity.UserVariable{},
	)
}

// Quit ends the database connection
func (ds *DBService) Quit() {
	ds.Quit()
}

// RowExists takes an SQL query and return true, if at least one entry
// exists for the given query
func (ds *DBService) RowExists(query string, args ...interface{}) bool {
	result := ds.db.Exec(fmt.Sprintf("SELECT exists (%s)", query), args...)
	return result.Error == nil
}
