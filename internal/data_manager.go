package internal

import (
	"Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/sessionstore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var sessMgr *sessionstore.SessionManager
var config *entity.Configuration

var configFile string = "config/app.yaml"

func GetSessionManager() *sessionstore.SessionManager {
	if sessMgr == nil {
		sessMgr = sessionstore.NewManager("tbs_sessid")
	}

	return sessMgr
}

func SetConfigurationFile(f string) {
	configFile = f
}

func GetConfiguration() *entity.Configuration {
	if config == nil {
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
	}

	return config
}
