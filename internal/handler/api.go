package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/builder"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/deploymentservice"

	"github.com/sirupsen/logrus"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/common"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/network"
)

type job struct {
	local  *entity.LocalDeployment
	email  *entity.EmailDeployment
	remote *entity.RemoteDeployment
	err    error
}

func (j job) nil() bool {
	return j.local == nil && j.email == nil && j.remote == nil
}

const errMsg = "failed %s deployment: %s"

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
	build := builder.NewBuild(bd, h.BuildService.GetBasePath())

	// set up directory structure for build
	if err := build.Setup(ctx); err != nil {
		logger.Error("failed to set up build: " + err.Error())
		build.AddReportEntryf("could not create build directory structure: %s", err.Error())
		be.Status = entity.StatusFailed
		h.saveReport(build, be)
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
	data := bd.Data
	repositoryUrl, err := h.BuildService.GetRepositoryUrl(ctx, &data, withCredentials)
	if err != nil {
		build.AddReportEntryf("could not determine repository url: %s", err.Error())
		be.Status = entity.StatusFailed
		h.saveReport(build, be)
		return
	}

	err = h.BuildService.CloneRepository(ctx, bd.Data.Repository.Branch, repositoryUrl, build.GetCloneDir())
	if err != nil {
		build.AddReportEntryf("could not clone repository: %s", err.Error())
		be.Status = entity.StatusFailed
		h.saveReport(build, be)
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
		be.Status = entity.StatusFailed
		h.saveReport(build, be)
		return
	}

	stepErrors := make([]error, 0)

	steps := bdc.GetSteps()
	for _, step := range steps {
		step = strings.Trim(step, "[]")
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
			step = strings.ReplaceAll(step, `\`, `\\`)
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

			build.AddReportEntry(string(output))
		}
	}

	if len(stepErrors) > 0 {
		build.AddReportEntryf("build steps produced %d errors", len(stepErrors))
		logger.Errorf("build steps produced %d errors", len(stepErrors))
		be.Status = entity.StatusFailed
		h.saveReport(build, be)
		return
	} else {
		build.AddReportEntry("build steps produced no errors")
	}

	// build is to be considered successful, if buildDir is not empty
	if num, err := os.ReadDir(build.GetBuildDir()); err != nil || len(num) == 0 {
		logger.Error("build did not produce any output")
		build.AddReportEntryf("build did not produce any output")
		be.Status = entity.StatusPartiallySucceeded
		h.saveReport(build, be)
		return
	}

	//prepare artifact (zip the build folder contents)
	if err = build.Pack(ctx); err != nil {
		logger.WithField("error", err.Error()).Error("build could not be packed")
		build.AddReportEntryf("build could not be packed: " + err.Error())
		be.Status = entity.StatusFailed
		h.saveReport(build, be)
		return
	}

	be.ArtifactPath = build.GetArtifact()

	logger.Trace("build succeeded")
	build.AddReportEntry("build succeeded")
	be.Status = entity.StatusSucceeded
	h.saveReport(build, be)

	var numJobs = len(bdc.Deployments.LocalDeployments) + len(bdc.Deployments.EmailDeployments) + len(bdc.Deployments.RemoteDeployments)

	jobs := make(chan job, numJobs)
	results := make(chan job, numJobs)
	for i := 0; i < 3; i++ { // 3 workers
		go func(jobs <-chan job, results chan<- job, l *logrus.Entry, b *builder.Build) {
			for j := range jobs {
				if j.local != nil {
					l.Trace("processing local deployment")
					if err := h.DeployService.DoLocalDeployment(ctx, j.local, b); err != nil && !errors.Is(err, deploymentservice.ErrDisabled) {
						j.err = err
					}
				} else if j.email != nil {
					l.Trace("processing email deployment")
					if err := h.DeployService.DoEmailDeployment(ctx, j.email, bdc.Repository.Name, b); err != nil && !errors.Is(err, deploymentservice.ErrDisabled) {
						j.err = err
					}
				} else if j.remote != nil {
					l.Trace("processing remote deployment")
					if err := h.DeployService.DoRemoteDeployment(ctx, j.remote, b); err != nil && !errors.Is(err, deploymentservice.ErrDisabled) {
						j.err = err
					}
				}
				// process job and put it into the result channel
				results <- j
			}
		}(jobs, results, h.ContextLogger("deployWorkerPool"), build)
	}

	for _, l := range bdc.Deployments.LocalDeployments {
		jobs <- job{local: &l}
		logger.Trace("adding local deployment")
	}
	for _, e := range bdc.Deployments.EmailDeployments {
		jobs <- job{email: &e}
		logger.Trace("adding email deployment")
	}
	for _, r := range bdc.Deployments.RemoteDeployments {
		jobs <- job{remote: &r}
		logger.Trace("adding remote deployment")
	}
	close(jobs)

	for i := 0; i < numJobs; i++ {
		r, ok := <-results
		if !ok {
			continue
		}
		if r.nil() {
			continue
		}
		if r.err != nil {
			var msg string
			if r.local != nil {
				logger.Trace("local deployment failed")
				msg = fmt.Sprintf(errMsg, "local", r.err)
			} else if r.email != nil {
				logger.Trace("email deployment failed")
				msg = fmt.Sprintf(errMsg, "email", r.err)
			} else if r.remote != nil {
				logger.Trace("remote deployment failed")
				msg = fmt.Sprintf(errMsg, "remote", r.err)
			}
			build.AddReportEntry(msg)
		}
	}

	build.AddReportEntry("all deployments finished")
	h.saveReport(build, be)
}

func (h *HttpHandler) saveReport(build *builder.Build, be *entity.BuildExecution) {
	be.ActionLog = build.GetReport()
	be.ExecutionTime = (time.Now().Sub(be.ExecutedAt)).Seconds()
	if err := h.DBService.UpdateBuildExecution(be); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"ID":    be.ID,
			"error": err.Error(),
		}).Error("failed to update build execution")
	}
}
