package deploymentservice

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/mailer"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

var ErrDisabled = errors.New("deployment is disabled")

type DeploymentService struct {
	Mailer *mailer.Mailer
}

func (dpl *DeploymentService) DoLocalDeployment(deployment *entity.LocalDeployment, build *entity.Build) error {
	if !deployment.Enabled {
		return ErrDisabled
	}

	fileBytes, err := ioutil.ReadFile(build.GetArtifact())
	if err != nil {
		return fmt.Errorf("could not read artifact file '%s': %s", build.GetArtifact(), err.Error())
	}

	if err := os.MkdirAll(filepath.Dir(deployment.Path), 0744); err != nil {
		return err
	}

	if err = os.WriteFile(deployment.Path, fileBytes, 0744); err != nil {
		return fmt.Errorf("could not write artifact (%s) to target (%s): %s", build.GetArtifact(), deployment.Path, err.Error())
	}

	return nil
}

func (dpl *DeploymentService) DoEmailDeployment(deployment *entity.EmailDeployment, repoName string, build *entity.Build) error {
	if !deployment.Enabled {
		return ErrDisabled
	}

	data := struct {
		Version string
		Title   string
	}{
		Version: "n/a", // TODO
		Title:   repoName,
	}

	emailBody, err := templateservice.ParseEmailTemplate(string(mailer.SubjNewDeployment), data)
	if err != nil {
		return fmt.Errorf("could not parse deployment email template: %s", err.Error())
	}
	err = dpl.Mailer.SendEmail(
		emailBody,
		string(mailer.SubjNewDeployment),
		[]string{deployment.Address},
		[]string{build.GetArtifact()},
	)
	if err != nil {
		return fmt.Errorf("could not send out deployment email to %s: %s", deployment.Address, err.Error())
	}

	return nil
}

func (dpl *DeploymentService) DoRemoteDeployment(deployment *entity.RemoteDeployment, build *entity.Build) error {
	if !deployment.Enabled {
		return ErrDisabled
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
	srcFile, err := os.Open(build.GetArtifact())
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
