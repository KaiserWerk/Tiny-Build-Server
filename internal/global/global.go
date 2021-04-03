package global

import (
	"database/sql"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/sessionstore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

var (
	sessMgr           *sessionstore.SessionManager
	config            *entity.Configuration
	configFile        = "config/app.yaml"
	loadConfOnce      sync.Once
	createSessMgrOnce sync.Once
	db                *sql.DB
	dbOnce            sync.Once
)

func GetDbConnection() *sql.DB {
	dbOnce.Do(func() {
		config := GetConfiguration()
		handle, err := sql.Open(config.Database.Driver, config.Database.DSN)
		if err != nil {
			panic(err.Error())
		}
		db = handle
	})
	return db
}

func Cleanup() {
	// close db connection
	db := GetDbConnection()
	err := db.Close()
	if err != nil {
		helper.WriteToConsole("could not close db connection: " + err.Error())
	}
	// flush log writer
}

func GetSessionManager() *sessionstore.SessionManager {
	createSessMgrOnce.Do(func() {
		sessMgr = sessionstore.NewManager("tbs_sessid")
	})

	return sessMgr
}

func SetConfigurationFile(f string) {
	configFile = f
}

func GetConfiguration() *entity.Configuration {
	loadConfOnce.Do(func() {
		if !helper.FileExists(configFile) {
			panic("configuration file '" + configFile + "' does not exist!")
		}
		cont, err := ioutil.ReadFile(configFile)
		if err != nil {
			panic("Could not read configuration file '" + configFile + "': " + err.Error())
		}

		var cfg *entity.Configuration
		err = yaml.Unmarshal(cont, &cfg)
		if err != nil {
			panic("could not parse configuration file content: " + err.Error())
		}

		config = cfg
	})

	return config
}
