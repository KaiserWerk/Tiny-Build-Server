package global

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"sync"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"

	"github.com/KaiserWerk/sessionstore"
	"gopkg.in/yaml.v2"
)

var (
	Version           = "0.0.0-dev"
	VersionDate       = "0000-00-00 00:00:00 +00:00"
	sessMgr           *sessionstore.SessionManager
	config            *entity.Configuration
	configFile        = "app.yaml"
	loadConfOnce      sync.Once
	createSessMgrOnce sync.Once
)

// GetSessionManager fetches the current session manager
func GetSessionManager() *sessionstore.SessionManager {
	createSessMgrOnce.Do(func() {
		sessMgr = sessionstore.NewManager("tbs_sessid")
	})

	return sessMgr
}

// SetConfigurationFile sets the path to a different configuration file
func SetConfigurationFile(f string) {
	configFile = f
}

// GetConfigurationFile returns the name of the configuration file
func GetConfigurationFile() string {
	return configFile
}

// GetConfiguration fetches the configuration from a given file
func GetConfiguration() *entity.Configuration {
	loadConfOnce.Do(func() {
		if !helper.FileExists(configFile) {
			content, err := assets.GetConfig("app.dist.yaml")
			if err != nil {
				panic("configuration dist file 'app.dist.yaml' could not be read!")
			}
			err = ioutil.WriteFile(configFile, content, 0600)
			if err != nil {
				panic("could not write config file: " + err.Error())
			}
			panic("Configuration file created; exiting...")
		}
		cont, err := ioutil.ReadFile(configFile)
		if err != nil {
			panic("Could not write configuration file '" + configFile + "': " + err.Error())
		}

		var cfg *entity.Configuration
		err = yaml.Unmarshal(cont, &cfg)
		if err != nil {
			panic("could not parse configuration file content: " + err.Error())
		}

		err = envconfig.Process("tbs", cfg)
		if err != nil {
			panic("could not process env vars for configuration: " + err.Error())
		}

		config = cfg
	})

	return config
}
