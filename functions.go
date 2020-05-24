package main

import (
	"bufio"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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

func readConsoleInput(externalShutdownCh chan bool) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		if err != nil {
			fmt.Printf("  could not process input %v\n", input)
			continue
		}

		switch string(input) {
		case "moo":
			moo := `                 (__)
                 (oo)
           /------\/
          / |    ||
         *  /\---/\
            ~~   ~~
..."Have you mooed today?"...`
			fmt.Println(moo)
		case "shutdown":
			close(externalShutdownCh)
		default:
			fmt.Printf("  unrecognized command: %v\n", string(input))
		}
	}
}

func startBuildProcess(id string, definition BuildDefinition) {
	/*
	* clone
	* restore
	* test
	* test bench
	* build arch

	arch = window_amd64, darwin_amd32, raspi_arm5, ...
	 */

	// noch zwischen go und c# unterscheiden

	baseDir := "build_definitions/build_" + id
	cloneDir := baseDir + "/clone"
	//buildDir := baseDir + "/build"
	// remove the clone directory possibly remaining
	// from previous build processes
	os.RemoveAll(cloneDir)

	// clone the repository
	repositoryUrl := getRepositoryUrl(definition, true)
	cmd := exec.Command("git", "clone", repositoryUrl, cloneDir)
	err := cmd.Run()
	if err != nil {
		fmt.Println("could not clone repository; aborting: " + err.Error())
		return
	}

	sysConf, err := loadSysConfig()
	if err != nil {
		fmt.Println("could not load system config")
	}
	// change dir
	err = os.Chdir(cloneDir)
	if err != nil {
		fmt.Println("could not change dir to clone: " + err.Error())
		return
	}

	switch definition.ProjectType {
	case "go":
	case "golang":
		// restore dependencies
		err = exec.Command(sysConf.GolangExecutable, "get", "./...").Run()
		if err != nil {
			fmt.Println("could not restore dependencies: " + err.Error())
			return
		}
		// tests and bench tests don't really matter for now
		err = exec.Command(sysConf.GolangExecutable, "test").Run()
		if err != nil {
			fmt.Println("could not restore dependencies: " + err.Error())
			return
		}
		err = exec.Command(sysConf.GolangExecutable, "test", "-bench=.").Run()
		if err != nil {
			fmt.Println("could not restore dependencies: " + err.Error())
			return
		}
		err = exec.Command(sysConf.GolangExecutable, "build", "-o", "../build/binary").Run()
		if err != nil {
			fmt.Println("could not build: " + err.Error())
			return
		}
	case "cs":
	case "csharp":

	}


	fmt.Println("build completed!")
}

func getRepositoryUrl(d BuildDefinition, withCredentials bool) string {
	var url string

	switch d.Repository.Host {
	case "bitbucket":
		url = "bitbucket.org/" + d.Repository.FullName
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.Repository.Username, d.Repository.Secret, url)
		}
		return "https://" + url
	case "github":
		url = "github.com/" + d.Repository.FullName
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.Repository.Username, d.Repository.Secret, url)
		}
		return "https://" + url
	case "gitlab":
		url = "gitlab.com/" + d.Repository.FullName
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.Repository.Username, d.Repository.Secret, url)
		}
		return "https://" + url
	case "gitea":
		url = d.Repository.HostUrl + "/" + d.Repository.FullName
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.Repository.Username, d.Repository.Secret, url)
		}
		return "https://" + url

	default:
		url = d.Repository.HostUrl + "/" + d.Repository.FullName
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.Repository.Username, d.Repository.Secret, url)
		}
		return "http://" + url
	}
}