package handler

import (
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"

	"github.com/gorilla/mux"
)

// BuildExecutionListHandler lists all build executions in in descending order
func (h *HTTPHandler) BuildExecutionListHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("BuildExecutionListHandler")
	)

	buildExecutions, err := h.DBService.GetNewestBuildExecutions(0, "")
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get build executions")
		h.SessionService.AddMessage(w, "success", "Failed to fetch build executions")
		return
	}

	buildDefinitions, err := h.DBService.GetAllBuildDefinitions()
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get build definition")
		h.SessionService.AddMessage(w, "success", "Failed to fetch build definitions")
		return
	}

	users, err := h.DBService.GetAllUsers()
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get usersn")
		h.SessionService.AddMessage(w, "success", "Failed to fetch user list")
		return
	}

	data := struct {
		CurrentUser      entity.User
		BuildExecutions  []entity.BuildExecution
		BuildDefinitions []entity.BuildDefinition
		Users            []entity.User
	}{
		CurrentUser:      currentUser,
		BuildExecutions:  buildExecutions,
		BuildDefinitions: buildDefinitions,
		Users:            users,
	}

	if err := templateservice.ExecuteTemplate(h.Injector(), w, "buildexecution_list.html", data); err != nil {
		w.WriteHeader(404)
	}
}

// BuildExecutionShowHandler shows details of a specific build execution
func (h *HTTPHandler) BuildExecutionShowHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("BuildExecutionShowHandler")
		vars        = mux.Vars(r)
	)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    id,
		}).Error("could not parse entry ID")
		w.WriteHeader(500)
		return
	}
	buildExecution, err := h.DBService.GetBuildExecutionById(id)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    id,
		}).Error("could not scan buildExecution")
		w.WriteHeader(500)
		return
	}

	buildDefinition, err := h.DBService.GetBuildDefinitionById(buildExecution.BuildDefinitionID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error":             err.Error(),
			"buildDefinitionId": buildExecution.BuildDefinitionID,
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

	if err = templateservice.ExecuteTemplate(h.Injector(), w, "buildexecution_show.html", data); err != nil {
		w.WriteHeader(404)
	}
}
