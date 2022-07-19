package configuration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

const (
	distFile = "app.dist.yaml"
)

// AppConfig contains the configuration, taken from a
// YAML configuration file
type AppConfig struct {
	Database struct {
		Driver string `yaml:"driver" envconfig:"db_driver"`
		DSN    string `yaml:"dsn" envconfig:"db_dsn"`
	} `yaml:"database"`
	Tls struct {
		CertFile string `yaml:"certfile" envconfig:"tls_certfile"`
		KeyFile  string `yaml:"keyfile" envconfig:"tls_keyfile"`
	}
	Build struct {
		BasePath string `yaml:"basepath" envconfig:"basepath"`
	}
}

func Setup(file string) (*AppConfig, bool, error) {
	var created bool
	if _, err := os.Stat(file); err != nil && errors.Is(err, os.ErrNotExist) {
		content, err := assets.GetConfig(distFile)
		if err != nil {
			return nil, false, fmt.Errorf("configuration dist file '%s' could not be read", distFile)
		}
		if err := ioutil.WriteFile(file, content, 0600); err != nil {
			return nil, false, fmt.Errorf("could not write config file: %s", err.Error())
		}
		created = true
	}
	cont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, false, fmt.Errorf("could not write configuration file '%s': %s", file, err.Error())
	}

	cfg := AppConfig{}
	setupDefaultValues(&cfg)

	err = yaml.Unmarshal(cont, &cfg)
	if err != nil {
		return nil, false, fmt.Errorf("could not unmarshal configuration file YAML content: %s", err.Error())
	}

	err = envconfig.Process("tbs", &cfg)
	if err != nil {
		return nil, false, fmt.Errorf("could not process env vars for configuration: %s", err.Error())
	}

	if s := os.Getenv("TBS_DB_DSN"); s != "" {
		cfg.Database.DSN = s
	}

	return &cfg, created, nil
}

func setupDefaultValues(a *AppConfig) {
	a.Database.Driver = "mysql"
	a.Database.DSN = "root:root@tcp(127.0.0.1:3306)/tinybuildserver?parseTime=true"
	a.Build.BasePath = "data"
}

type Settings map[string]string
