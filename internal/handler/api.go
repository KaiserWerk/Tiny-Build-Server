package handler

import (
	"context"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/common"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/network"
)

// PayloadReceiveHandler takes care of accepting the payload from the webhook HTTP call
// sent by a Git hoster
func (h *HttpHandler) PayloadReceiveHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger := h.ContextLogger("PayloadReceiveHandler")

	// get token
	token := r.URL.Query().Get("token")
	if token == "" {
		logger.Error("missing token")
		http.Error(w, "could not determine token", http.StatusBadRequest)
		return
	}

	// find build definition by token
	bd, err := h.DBService.FindBuildDefinition("token = ?", token)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"token": token,
		}).Error("could not find build definition for token")
		http.Error(w, fmt.Sprintf("could not find build definition for token %s: %s", token, err.Error()), http.StatusNotFound)
		return
	}

	if bd.Deleted {
		logger.WithFields(logrus.Fields{
			"id": bd.ID,
		}).Info("requested deleted build definition")
		http.Error(w, "requested deleted build definition", http.StatusNotFound)
		return
	}

	variables, err := h.DBService.GetAvailableVariablesForUser(bd.CreatedBy)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not determine variables for user")
		http.Error(w, fmt.Sprintf("could not determine variables for user: %s", err.Error()), http.StatusNotFound)
		return
	}

	// unmarshal the build definition content
	bdContent, err := common.UnmarshalBuildDefinition([]byte(bd.Raw), variables)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not unmarshal build definition")
		http.Error(w, "could not unmarshal build definition content: "+err.Error(), http.StatusNotFound)
		return
	}
	bd.Data = bdContent

	// check if the correct headers, depending on the hoster, are set and
	// have the correct values
	if err = network.CheckPayloadHeader(bdContent, r); err != nil {
		logger.WithField("error", err.Error()).Error("request headers are incorrect")
		http.Error(w, "request headers are incorrect", http.StatusBadRequest)
		return
	}

	logger.Debug("payload received")

	// insert new build execution
	be := entity.NewBuildExecution(bd.ID, 0)
	if err := h.DBService.AddBuildExecution(be); err != nil {
		logger.WithField("error", err.Error()).Error("failed to add build execution")
		http.Error(w, "failed to add build execution", http.StatusBadRequest)
		return
	}

	// start the actual build process
	go h.InitiateBuildProcess(&bd, be)
}

func (h *HttpHandler) InitiateBuildProcess(bd *entity.BuildDefinition, be *entity.BuildExecution) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger := h.ContextLogger("InitiateBuildProcess")
	build := entity.NewBuild(bd, h.BuildService.GetBasePath())
	defer func() {
		if err := h.DBService.UpdateBuildExecution(be); err != nil {
			logger.WithFields(logrus.Fields{
				"ID":    be.ID,
				"error": err.Error(),
			}).Error("failed to update build execution")
		}
	}()

	// set up directory structure for build
	if err := build.Setup(ctx); err != nil {
		logger.Error("failed to set up build: " + err.Error())
		build.AddReportEntryf("could not create build directory structure: %s", err.Error())
		be.Status = entity.StatusFailed
		return
	}

	// clone the repository
	withCredentials := bd.Data.Repository.AccessSecret != ""
	if withCredentials && bd.Data.Repository.AccessUser == "" {
		bd.Data.Repository.AccessUser = "nobody"
	}

	// if no branch is set, use master as the default
	if bd.Data.Repository.Branch == "" {
		bd.Data.Repository.Branch = "master"
	}

	repositoryUrl, err := h.BuildService.GetRepositoryUrl(ctx, &(bd.Data), withCredentials)
	if err != nil {
		build.AddReportEntryf("could not determine repository url: %s", err.Error())
		be.Status = entity.StatusFailed
		return
	}

	err = h.BuildService.CloneRepository(ctx, bd.Data.Repository.Branch, repositoryUrl, build.GetCloneDir())
	if err != nil {
		build.AddReportEntryf("could not clone repository: %s", err.Error())
		be.Status = entity.StatusFailed
		return
	}

	// set up special variables
	vars := []entity.UserVariable{{
		Variable: "buildDir",
		Value:    build.GetBuildDir(),
	}, {
		Variable: "cloneDir",
		Value:    build.GetCloneDir(),
	}}

	// do the unmarshal again with updated variables
	bdc, err := buildservice.GetPreparedContent(ctx, bd, vars)
	if err != nil {
		build.AddReportEntryf("could not unmarshal build definition: %s", err.Error())
		return
	}

	stepErrors := make([]error, 0)

	steps := bdc.GetSteps()
	for _, step := range steps {
		build.AddReportEntryf("step: %s", step)
		switch true {
		case strings.HasPrefix(step, "setenv"):
			parts, err := common.SplitCommand(step)
			if err != nil {
				stepErrors = append(stepErrors, err)
				build.AddReportEntryf("could not prepare step command '%s': %s", step, err.Error())
				continue
			}
			if len(parts) != 3 {
				build.AddReportEntryf("step '%s' has an invalid format", step)
				continue
			}

			if err = os.Setenv(parts[1], parts[2]); err != nil {
				stepErrors = append(stepErrors, err)
				build.AddReportEntryf("step '%s' was not successful: '%s'", step, err.Error())
				continue
			}
		case strings.HasPrefix(step, "unsetenv"):
			parts, err := common.SplitCommand(step)
			if err != nil {
				stepErrors = append(stepErrors, err)
				build.AddReportEntryf("could not prepare step command '%s': %s", step, err.Error())
				continue
			}
			if len(parts) != 2 {
				build.AddReportEntryf("step '%s' has an invalid format", step)
				continue
			}
			err = os.Unsetenv(parts[1])
			if err != nil {
				stepErrors = append(stepErrors, err)
				build.AddReportEntryf("step '%s' was not successful: '%s'", step, err.Error())
				continue
			}
		default:
			parts, err := common.SplitCommand(step)
			if err != nil {
				stepErrors = append(stepErrors, err)
				build.AddReportEntryf("could not prepare step command '%s': %s", step, err.Error())
				continue
			}
			var cmd *exec.Cmd
			if len(parts) <= 0 { // :)
				build.AddReportEntry("empty step; skipping")
				continue
			} else if len(parts) == 1 {
				cmd = exec.CommandContext(ctx, parts[0])
			} else {
				cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
			}
			cmd.Dir = build.GetCloneDir()

			output, err := cmd.CombinedOutput()
			if err != nil {
				stepErrors = append(stepErrors, err)
				build.AddReportEntryf("could not execute command '%s': '%s' -> (%s)", cmd.String(), err.Error(), strings.TrimSpace(string(output)))
				continue
			}

			build.AddReportEntry(strings.TrimSpace(string(output)))
		}
	}

	// build is to be considered successful, if buildDir is not empty
	if num, err := os.ReadDir(build.GetBuildDir()); err != nil || len(num) == 0 {
		logger.Error("build did not produce any output")
		build.AddReportEntryf("build did not produce any output")
		build.SetStatus(entity.StatusPartiallySucceeded)
		return
	}

	//prepare artifact (zip the build folder contents)
	if err = build.Pack(ctx); err != nil {
		logger.WithField("error", err.Error()).Error("build could not be packed")
		build.AddReportEntryf("build could not be packed: " + err.Error())
		build.SetStatus(entity.StatusFailed)
		return
	}

	logger.Trace("build succeeded")
	build.SetStatus(entity.StatusSucceeded)
	build.AddReportEntry("build succeeded")

	// Local Deployments
	for _, ld := range bdc.Deployments.LocalDeployments {
		go func(l *logrus.Entry, d *entity.LocalDeployment, b *entity.Build) {
			if err := h.DeployService.DoLocalDeployment(ctx, d, b); err != nil {
				l.WithField("error", err.Error()).Error("failed to do local deployment")
				build.AddReportEntryf("failed to execute local deployment: %s", err.Error())
			}
		}(h.ContextLogger("LocalDeployment"), &ld, build)
	}

	// Email Deployments
	for _, ed := range bdc.Deployments.EmailDeployments {
		go func(l *logrus.Entry, d *entity.EmailDeployment, b *entity.Build, repo string) {
			if err := h.DeployService.DoEmailDeployment(ctx, d, repo, b); err != nil {
				l.WithField("error", err.Error()).Error("failed to do email deployment")
				build.AddReportEntryf("failed to execute email deployment: %s", err.Error())
			}
		}(h.ContextLogger("EmailDeployment"), &ed, build, bdc.Repository.Name)
	}

	// Remote Deployments
	for _, rd := range bdc.Deployments.RemoteDeployments {
		go func(l *logrus.Entry, d *entity.RemoteDeployment, b *entity.Build) {
			if err := h.DeployService.DoRemoteDeployment(ctx, d, b); err != nil {
				l.WithField("error", err.Error()).Error("failed to do remote deployment")
				build.AddReportEntryf("failed to execute remote deployment: %s", err.Error())
			}
		}(h.ContextLogger("RemoteDeployment"), &rd, build)
	}
}
