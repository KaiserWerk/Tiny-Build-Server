package handler

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

// DownloadNewestArtifactHandler downloads the most recently created version
// of a produced artifact
func (h *HttpHandler) DownloadNewestArtifactHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		vars = mux.Vars(r)
		logger = h.ContextLogger("DownloadNewestArtifactHandler")
	)

	beList, err := h.Ds.GetNewestBuildExecutions(1, "build_definition_id = ?", vars["id"])
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"buildDefinitionId": vars["id"],
		}).Error("could not fetch build executions by definition id")
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	if len(beList) < 1 {
		logger.WithField("error", err.Error()).Info("could not find any build executions for definition")
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	artifact, err := filepath.Abs(beList[0].ArtifactPath)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"artifactPath": beList[0].ArtifactPath,
		}).Debug("could not determine absolute path of file: ")
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	artifact += ".zip"

	//fmt.Printf("file to serve: %s\n", artifact)

	cont, err := ioutil.ReadFile(artifact)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"artifactFile": artifact,
		}).Debug("could not read artifact file")
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filepath.Base(artifact)))
	w.Write(cont)
}

// DownloadSpecificArtifactHandler downloads a artifact produced by a specific
// build execution
func (h *HttpHandler) DownloadSpecificArtifactHandler(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
		logger = logging.New(logrus.InfoLevel, "DownloadNewestArtifactHandler", true)
	)
	id, _ := strconv.Atoi(vars["id"])
	be, err := h.Ds.GetBuildExecutionById(id)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"buildExecutionId": id,
		}).Error("could not fetch build execution by ID")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	artifact, err := filepath.Abs(be.ArtifactPath)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"artifcatPath": be.ArtifactPath,
		}).Info("could not determine absolute path of file")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	artifact += ".zip"
	cont, err := ioutil.ReadFile(artifact)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"artifactFile": artifact,
		}).Info("could not read artifact file")
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filepath.Base(artifact)))
	w.Write(cont)
}
