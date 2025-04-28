package dbservice

import (
	"database/sql"
	"fmt"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type IDBService interface {
	AutoMigrate() error
	Quit()
	RowExists(query string, args ...any) bool
	GetNewestBuildDefinitions(limit int) ([]entity.BuildDefinition, error)
	GetAllBuildDefinitions() ([]entity.BuildDefinition, error)
	FindBuildDefinition(cond string, args ...any) (entity.BuildDefinition, error)
	GetBuildDefinitionById(id uint) (entity.BuildDefinition, error)
	GetBuildDefCaption(id uint) (string, error)
	DeleteBuildDefinition(bd *entity.BuildDefinition) error
	AddBuildDefinition(bd *entity.BuildDefinition) (uint, error)
	UpdateBuildDefinition(bd *entity.BuildDefinition) error

	GetNewestBuildExecutions(limit int, query string, args ...any) ([]entity.BuildExecution, error)
	GetBuildExecutionById(id int) (entity.BuildExecution, error)
	FindBuildExecutions(query any, args ...any) ([]entity.BuildExecution, error)
	AddBuildExecution(be *entity.BuildExecution) error
	UpdateBuildExecution(be *entity.BuildExecution) error

	GetAllSettings() (map[string]string, error)
	SetSetting(name, value string) error

	GetAllUsers() ([]entity.User, error)
	GetUsernameById(id int) string
	GetUserByEmail(email string) (entity.User, error)
	GetUserById(id uint) (entity.User, error)
	AddUser(user entity.User) (uint, error)
	UpdateUser(user entity.User) error
	FindUser(cond string, args ...any) (entity.User, error)
	DeleteUser(id uint) error

	InsertUserAction(userId uint, purpose, token string, validity sql.NullTime) error
	GetUserActionByToken(token string) (entity.UserAction, error)
	InvalidatePasswordResets(userId uint) error
	AddUserAction(action entity.UserAction) error
	UpdateUserAction(userAction entity.UserAction) error

	GetAvailableVariablesForUser(userId uint) ([]entity.UserVariable, error)
	AddVariable(userVar entity.UserVariable) (uint, error)
	GetVariable(id int) (entity.UserVariable, error)
	FindVariable(cond string, args ...any) (entity.UserVariable, error)
	UpdateVariable(userVar entity.UserVariable) error
	DeleteVariable(id uint) error
}

type DBService struct {
	db *gorm.DB
}

func New(cfg *configuration.AppConfig) IDBService {
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

}

// RowExists takes an SQL query and return true, if at least one entry
// exists for the given query
func (ds *DBService) RowExists(query string, args ...interface{}) bool {
	result := ds.db.Exec(fmt.Sprintf("SELECT exists (%s)", query), args...)
	return result.Error == nil
}
