package deploymentservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/builder"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/mailer"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

var (
	ErrDisabled = errors.New("deployment is disabled")
	ErrCanceled = errors.New("deploymentservice: canceled by context")
)

type DeploymentService struct {
	Mailer *mailer.Mailer
}

func (dpl *DeploymentService) DoLocalDeployment(ctx context.Context, deployment *entity.LocalDeployment, build *builder.Build) error {
	if !deployment.Enabled {
		return ErrDisabled
	}
	if ctx.Err() != nil {
		return ErrCanceled
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

func (dpl *DeploymentService) DoEmailDeployment(ctx context.Context, deployment *entity.EmailDeployment, repoName string, build *builder.Build) error {
	if !deployment.Enabled {
		return ErrDisabled
	}
	if ctx.Err() != nil {
		return ErrCanceled
	}

	data := struct {
		Version string
		Title   string
	}{
		Version: "n/a", // TODO
		Title:   repoName,
	}

	emailBody, err := templateservice.ParseEmailTemplate(mailer.GetTemplateFromSubject(mailer.SubjNewDeployment), data)
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

func (dpl *DeploymentService) DoRemoteDeployment(ctx context.Context, deployment *entity.RemoteDeployment, build *builder.Build) error {
	if !deployment.Enabled {
		return ErrDisabled
	}
	if ctx.Err() != nil {
		return ErrCanceled
	}

	// first, the pre deployment actions
	sshConfig := &ssh.ClientConfig{
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User: deployment.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(deployment.Password),
		},
	}
	sshConfig.HostKeyCallback = func(_ string, _ net.Addr, _ ssh.PublicKey) error {
		return nil
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", deployment.Host, deployment.Port), sshConfig)
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
	// TODO: für mehrere Dateien/entzippen überarbeiten
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
	dstFile, err := sftpClient.Create(deployment.WorkingDirectory + "/" + filepath.Base(build.GetArtifact()))
	if err != nil {
		return err
	}

	// create source file
	srcFile, err := os.Open(build.GetArtifact())
	if err != nil {
		return err
	}

	// copy source file to destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	_ = dstFile.Close()
	_ = srcFile.Close()

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
