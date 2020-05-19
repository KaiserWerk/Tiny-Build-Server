package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
)

func getHeaderIfSet(r *http.Request, key string) (string, error) {
	header := r.Header.Get(key)
	if header == "" {
		return "", errors.New("header is not set or empty")
	}
	return header, nil
}
func loadSysConfig() (SysConfig, error) {
	cont, err := ioutil.ReadFile("config/app.yaml")
	if err != nil {
		return SysConfig{}, errors.New("could not read config/app.yaml file")
	}
	var config SysConfig
	err = yaml.Unmarshal(cont, &config)
	if err != nil {
		return SysConfig{}, errors.New("could not parse config/app.yaml file")
	}

	return config, nil
}

func loadBuildDefinition(id string) (BuildDefinition, error) {
	bdDir := "build_definitions/build_" + id
	bdFile := bdDir + "/build.yaml"

	if _, err := os.Stat(bdDir); os.IsNotExist(err) {
		fmt.Printf("build definition with id %v not found\n", id)
		return BuildDefinition{}, BuildDefinitionNotFound{Id: id}
	}

	if _, err := os.Stat(bdFile); os.IsNotExist(err) {
		fmt.Printf("config file for build definition with id %v not found\n", id)
		return BuildDefinition{}, BuildDefinitionConfigFileNotFound{Id: id}
	}

	cont, err := ioutil.ReadFile(bdFile)
	if err != nil {
		fmt.Println("could not read build definition config file")
		return BuildDefinition{}, errors.New("could not read build definition config file")
	}
	var bd BuildDefinition
	err = yaml.Unmarshal(cont, &bd)
	if err != nil {
		fmt.Println("could not unmarshal yaml")
		return BuildDefinition{}, errors.New("could not unmarshal yaml")
	}

	return bd, nil
}
