package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
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
		return BuildDefinition{}, buildDefinitionNotFound{Id: id}
	}

	if _, err := os.Stat(bdFile); os.IsNotExist(err) {
		fmt.Printf("config file for build definition with id %v not found\n", id)
		return BuildDefinition{}, buildDefinitionConfigFileNotFound{Id: id}
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

		arch = window_amd64, darwin_amd32, raspi3, ...
	*/

	repoName := strings.Split(definition.Repository.FullName, "/")[1]

	fileExt := make(map[string]string)
	fileExt["windows"] = ".exe"
	fileExt["linux"] = ""
	fileExt["darwin"] = ".osx"

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
		return
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
		outputFile := ""
		for _, v := range definition.Actions {
			switch true {
			case strings.Contains(v, "restore"):
				// restore dependencies
				err = exec.Command(sysConf.GolangExecutable, "get", "./...").Run()
				if err != nil {
					fmt.Println("could not restore dependencies: " + err.Error())
					return
				}
			case strings.Contains(v, "test"):
				// tests and bench tests don't really matter for now
				err = exec.Command(sysConf.GolangExecutable, "test").Run()
				if err != nil {
					fmt.Println("could not restore dependencies: " + err.Error())
					return
				}
			case strings.Contains(v, "test bench"):
				err = exec.Command(sysConf.GolangExecutable, "test", "-bench=.").Run()
				if err != nil {
					fmt.Println("could not restore dependencies: " + err.Error())
					return
				}
			case strings.Contains(v, "build"):
				var (
					targetOS   string
					targetArch string
					targetArm  string
				)

				osArch := strings.Split(v, " ")[1]

				// its sth like raspi
				if !strings.Contains(osArch, "_") {
					switch osArch {
					case "raspi3":
						targetOS = "linux"
						targetArch = "arm"
						targetArm = "5"
					case "raspi4":
						targetOS = "linux"
						targetArch = "arm"
						targetArm = "6"
					}
				} else {
					parts := strings.Split(osArch, "_")
					targetOS = parts[0]
					targetArch = parts[1]
				}

				_ = os.Setenv("GOOS", targetOS)
				_ = os.Setenv("GOARCH", targetArch)
				_ = os.Setenv("GOARM", targetArm)
				outputFile = fmt.Sprintf("../build/%s%s", repoName, fileExt[targetOS])
				err = exec.Command(sysConf.GolangExecutable, "build", "-o", outputFile).Run()
				if err != nil {
					fmt.Println("could not build: " + err.Error())
					return
				}
			}
		}

		if definition.DeploymentEnabled {
			deployToHost(outputFile, definition)
		}
		// @TODO reset cwd?

	case "cs":
	case "csharp":
		// sysconfig!
		// @TODO
	}

	fmt.Println("build completed!")
}

func deployToHost(outputFile string, definition BuildDefinition) {
	for _, v := range definition.Deployments {
		// first, the pre deployment actions
		sshConfig := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			User:            v.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(v.Password),
			},
		}
		sshClient, err := ssh.Dial("tcp", v.Host, sshConfig)
		if err != nil {
			fmt.Println("could not establish ssh connection:", err.Error())
			return
		}

		if len(v.PreDeploymentActions) > 0 {
			for _, action := range v.PreDeploymentActions {
				session, err := sshClient.NewSession()
				if err != nil {
					fmt.Printf("Failed to create session: %s\n", err.Error())
					return
				}
				outp, err := session.Output(action)
				if err != nil {
					fmt.Println("could not execute pre deployment command: " + action)
					//return
				} else {
					outpDisplay := string(outp)
					if outpDisplay != "" {
						fmt.Println("output from pre remote command:", outpDisplay)
					}
				}
				_ = session.Close()
			}
		}

		session, err := sshClient.NewSession()
		if err != nil {
			fmt.Printf("Failed to create session: %s\n", err.Error())
			return
		}

		// then, the actual deployment
		sftpClient, err := sftp.NewClient(sshClient)
		if err != nil {
			fmt.Println("could not create sftp client instance:", err.Error())
			return
		}

		// create destination file
		// @TODO really necessary?
		dstFile, err := sftpClient.Create(v.WorkingDirectory)
		if err != nil {
			fmt.Println("failed to create remote file:", err.Error())
			return
		}

		// create source file
		srcFile, err := os.Open(outputFile)
		if err != nil {
			fmt.Println("failed to open source file:", err.Error())
			return
		}

		// copy source file to destination file
		bytes, err := io.Copy(dstFile, srcFile)
		if err != nil {
			fmt.Println("failed to copy file:", err.Error())
			return
		}
		_ = dstFile.Close()
		_ = srcFile.Close()
		fmt.Printf("%d bytes copied\n", bytes)
		_ = session.Close()

		// then, the post deployment actions
		if len(v.PostDeploymentActions) > 0 {
			for _, action := range v.PostDeploymentActions {
				session, err := sshClient.NewSession()
				if err != nil {
					fmt.Printf("Failed to create session: %s\n", err.Error())
					return
				}
				outp, err := session.Output(action)
				if err != nil {
					fmt.Println("could not execute post deployment command:", action, "cause:", err.Error())
					//return
				} else {
					outpDisplay := string(outp)
					if outpDisplay != "" {
						fmt.Println("output from post remote command:", outpDisplay)
					}
				}
				_ = session.Close()
			}
		}

		if sshClient != nil {
			err = sshClient.Close()
			if err != nil {
				fmt.Println("could not close ssh client connection")
				return
			}
		}

		_ = sftpClient.Close()
	}
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
		return "http://" + url // just http
	}
}
