package configuration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

const (
	distFile = "app.dist.yaml"
)

func Setup(file string) (*entity.Configuration, bool, error) {
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

	var cfg entity.Configuration
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
