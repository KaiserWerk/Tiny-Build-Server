package dbservice

import (
	"database/sql"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/gorm"
)

type DBServiceMock struct {
}

func (m *DBServiceMock) AutoMigrate() error {
	return nil
}

func (m *DBServiceMock) Quit() {}

func (m *DBServiceMock) RowExists(query string, args ...interface{}) bool {
	return true
}

func (m *DBServiceMock) GetNewestBuildDefinitions(limit int) ([]entity.BuildDefinition, error) {
	return []entity.BuildDefinition{}, nil
}
func (m *DBServiceMock) GetAllBuildDefinitions() ([]entity.BuildDefinition, error) {
	return []entity.BuildDefinition{}, nil
}
func (m *DBServiceMock) FindBuildDefinition(cond string, args ...any) (entity.BuildDefinition, error) {
	return entity.BuildDefinition{
		Model: gorm.Model{
			ID: 1,
		},
		Token:     "abc123",
		CreatedBy: 1,

		Raw: `project_type:
repository:
  hoster: github
  hoster_url:
  name: KaiserWerk/GitHub-Public-Golang-Test-Repo
  access_user:
  access_secret:
  branch: master
setup:
  -
test:
  -
pre_build:
  -
build:
  -
post_build:
  -
deployments:
  local_deployments:
    - enabled: false
      path: /your/target/path
  email_deployments:
    - enabled: false
      address:
  remote_deployments:
    - enabled: false
      host:
      port:
      connection_type: sftp
      username:
      password:
      working_directory:
      pre_deployment_steps:
      post_deployment_steps:
`,
	}, nil
}
func (m *DBServiceMock) GetBuildDefinitionById(id uint) (entity.BuildDefinition, error) {
	return entity.BuildDefinition{}, nil
}
func (m *DBServiceMock) GetBuildDefCaption(id uint) (string, error) {
	return "", nil
}
func (m *DBServiceMock) DeleteBuildDefinition(bd *entity.BuildDefinition) error {
	return nil
}
func (m *DBServiceMock) AddBuildDefinition(bd *entity.BuildDefinition) (uint, error) {
	return 0, nil
}
func (m *DBServiceMock) UpdateBuildDefinition(bd *entity.BuildDefinition) error {
	return nil
}
func (m *DBServiceMock) GetNewestBuildExecutions(limit int, query string, args ...any) ([]entity.BuildExecution, error) {
	return []entity.BuildExecution{}, nil
}
func (m *DBServiceMock) GetBuildExecutionById(id int) (entity.BuildExecution, error) {
	return entity.BuildExecution{}, nil
}
func (m *DBServiceMock) FindBuildExecutions(query any, args ...any) ([]entity.BuildExecution, error) {
	return []entity.BuildExecution{}, nil
}
func (m *DBServiceMock) AddBuildExecution(be *entity.BuildExecution) error {
	return nil
}
func (m *DBServiceMock) UpdateBuildExecution(be *entity.BuildExecution) error {
	return nil
}

func (m *DBServiceMock) GetAllSettings() (map[string]string, error) {
	return map[string]string{}, nil
}
func (m *DBServiceMock) SetSetting(name, value string) error {
	return nil
}

func (m *DBServiceMock) GetAllUsers() ([]entity.User, error) {
	return []entity.User{}, nil
}
func (m *DBServiceMock) GetUsernameById(id int) string {
	return ""
}
func (m *DBServiceMock) GetUserByEmail(email string) (entity.User, error) {
	return entity.User{}, nil
}
func (m *DBServiceMock) GetUserById(id uint) (entity.User, error) {
	return entity.User{}, nil
}
func (m *DBServiceMock) AddUser(user entity.User) (uint, error) {
	return 0, nil
}
func (m *DBServiceMock) UpdateUser(user entity.User) error {
	return nil
}
func (m *DBServiceMock) FindUser(cond string, args ...any) (entity.User, error) {
	return entity.User{}, nil
}
func (m *DBServiceMock) DeleteUser(id uint) error {
	return nil
}

func (m *DBServiceMock) InsertUserAction(userId uint, purpose, token string, validity sql.NullTime) error {
	return nil
}
func (m *DBServiceMock) GetUserActionByToken(token string) (entity.UserAction, error) {
	return entity.UserAction{}, nil
}
func (m *DBServiceMock) InvalidatePasswordResets(userId uint) error {
	return nil
}
func (m *DBServiceMock) AddUserAction(action entity.UserAction) error {
	return nil
}
func (m *DBServiceMock) UpdateUserAction(userAction entity.UserAction) error {
	return nil
}

func (m *DBServiceMock) GetAvailableVariablesForUser(userId uint) ([]entity.UserVariable, error) {
	return []entity.UserVariable{}, nil
}
func (m *DBServiceMock) AddVariable(userVar entity.UserVariable) (uint, error) {
	return 0, nil
}
func (m *DBServiceMock) GetVariable(id int) (entity.UserVariable, error) {
	return entity.UserVariable{}, nil
}
func (m *DBServiceMock) FindVariable(cond string, args ...any) (entity.UserVariable, error) {
	return entity.UserVariable{}, nil
}
func (m *DBServiceMock) UpdateVariable(userVar entity.UserVariable) error {
	return nil
}
func (m *DBServiceMock) DeleteVariable(id uint) error {
	return nil
}
