package global

import (
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
)

func Cleanup() {
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
