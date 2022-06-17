package deploymentservice

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/fixtures"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func DoLocalDeployment(deployment *entity.LocalDeployment, artifact entity.Artifact) error {
	if !deployment.Enabled {
		return fmt.Errorf("skipping disabled deployment")
	}

	fileBytes, err := ioutil.ReadFile(artifact.FullPath())
	if err != nil {
		return fmt.Errorf("could not read artifact (%s): %s", artifact, err.Error())
	}

	_ = os.MkdirAll(filepath.Dir(deployment.Path), 0744)

	err = os.WriteFile(deployment.Path, fileBytes, 0744)
	if err != nil {
		return fmt.Errorf("could not write artifact (%s) to target (%s): %s", artifact, deployment.Path, err.Error())
	}

	return nil
}

func DoEmailDeployment(deployment *entity.EmailDeployment, repoName string, settings map[string]string, zipArchiveName string) error {
	if !deployment.Enabled {
		return fmt.Errorf("skipping disabled deployment")
	}

	data := struct {
		Version string
		Title   string
	}{
		Version: "n/a", // TODO
		Title:   repoName,
	}

	emailBody, err := templateservice.ParseEmailTemplate(string(fixtures.DeploymentEmail), data)
	if err != nil {
		return fmt.Errorf("could not parse deployment email template: %s", err.Error())
	}
	err = helper.SendEmail(
		settings,
		emailBody,
		fixtures.EmailSubjects[fixtures.DeploymentEmail],
		[]string{deployment.Address},
		[]string{zipArchiveName},
	)
	if err != nil {
		return fmt.Errorf("could not send out deployment email to %s: %s", deployment.Address, err.Error())
	}

	return nil
}

func DoRemoteDeployment(deployment *entity.RemoteDeployment, artifact entity.Artifact) error {
	if !deployment.Enabled {
		return fmt.Errorf("skipping disabled deployment")
	}
	// first, the pre deployment actions
	sshConfig := &ssh.ClientConfig{
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User: deployment.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(deployment.Password),
		},
	}
	sshClient, err := ssh.Dial("tcp", deployment.Host, sshConfig)
	if err != nil {
		return err
	}

	if len(deployment.PreDeploymentSteps) > 0 {
		for _, action := range deployment.PreDeploymentSteps {
			session, err := sshClient.NewSession()
			if err != nil {
				return err
			}
			if err = session.Run(action); err != nil {
				return err
			}
			_ = session.Close()
		}
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}

	// then, the actual deployment
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}

	// create destination file
	dstFile, err := sftpClient.Create(deployment.WorkingDirectory)
	if err != nil {
		return err
	}

	// create source file
	srcFile, err := os.Open(artifact.FullPath())
	if err != nil {
		return err
	}

	// copy source file to destination file
	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	_ = dstFile.Close()
	_ = srcFile.Close()
	fmt.Printf("%d bytes copied\n", bytes)
	_ = session.Close()

	// then, the post deployment actions
	if len(deployment.PostDeploymentSteps) > 0 {
		for _, action := range deployment.PostDeploymentSteps {
			session, err := sshClient.NewSession()
			if err != nil {
				return err
			}
			if err = session.Run(action); err != nil {
				return err
			}
			_ = session.Close()
		}
	}

	_ = sftpClient.Close()
	_ = sshClient.Close()

	return nil
}
