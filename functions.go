package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

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
