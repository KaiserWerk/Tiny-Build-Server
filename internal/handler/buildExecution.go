package handler

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"

	"github.com/gorilla/mux"
)

// BuildExecutionListHandler lists all build executions in in descending order
func (h *HttpHandler) BuildExecutionListHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger = h.ContextLogger("BuildExecutionListHandler")
	)

	buildExecutions, err := h.Ds.GetNewestBuildExecutions(0, "")
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get build executions")
		return
	}

	data := struct {
		CurrentUser     entity.User
		BuildExecutions []entity.BuildExecution
	}{
		CurrentUser:     currentUser,
		BuildExecutions: buildExecutions,
	}

	if err := templateservice.ExecuteTemplate(w, "buildexecution_list.html", data); err != nil {
		w.WriteHeader(404)
	}
}

// BuildExecutionShowHandler shows details of a specific build execution
func (h *HttpHandler) BuildExecutionShowHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger = h.ContextLogger("BuildExecutionShowHandler")
		vars = mux.Vars(r)
	)


	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id": id,
		}).Error("could not parse entry ID")
		w.WriteHeader(500)
		return
	}
	buildExecution, err := h.Ds.GetBuildExecutionById(id)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id": id,
		}).Error("could not scan buildExecution")
		w.WriteHeader(500)
		return
	}

	buildDefinition, err := h.Ds.GetBuildDefinitionById(buildExecution.BuildDefinitionId)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"buildDefinitionId": buildExecution.BuildDefinitionId,
		}).Error("could not scan buildDefinition")
		w.WriteHeader(500)
		return
	}

	data := struct {
		CurrentUser     entity.User
		BuildExecution  entity.BuildExecution
		BuildDefinition entity.BuildDefinition
	}{
		CurrentUser:     currentUser,
		BuildExecution:  buildExecution,
		BuildDefinition: buildDefinition,
	}

	if err = templateservice.ExecuteTemplate(w, "buildexecution_show.html", data); err != nil {
		w.WriteHeader(404)
	}
}
