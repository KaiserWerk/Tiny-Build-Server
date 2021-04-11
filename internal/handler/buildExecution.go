package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"net/http"
	"strconv"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/gorilla/mux"
)

func BuildExecutionListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	//db := helper.GetDbConnection()
	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "buildexecution_list.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func BuildExecutionShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseService.New()
	defer ds.Quit()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("BuildExecutionShowHandler: could not parse entry ID: " + err.Error())
		w.WriteHeader(500)
		return
	}
	//var be entity.BuildExecution
	//row := db.QueryRow("SELECT id, build_definition_id, initiated_by, manual_run, action_log,"+
	//	" result, artifact_path, execution_time, executed_at FROM build_execution WHERE id = ?", vars["id"])
	//err = row.Scan(&be.Id, &be.BuildDefinitionId, &be.InitiatedBy, &be.ManualRun, &be.ActionLog,
	//	&be.Result, &be.ArtifactPath, &be.ExecutionTime, &be.ExecutedAt)
	buildExecution, err := ds.GetBuildExecutionById(id)
	if err != nil {
		helper.WriteToConsole("could not scan buildExecution in buildExecutionShowHandler")
		w.WriteHeader(500)
		return
	}

	//be.ActionLog = strings.ReplaceAll(be.ActionLog, "\n", "<br>")

	//var bd entity.BuildDefinition
	//row = db.QueryRow("SELECT id, caption FROM build_definition WHERE id = ?", buildExecution.BuildDefinitionId)
	//err = row.Scan(&bd.Id, &bd.Caption)
	buildDefinition, err := ds.GetBuildDefinitionById(buildExecution.BuildDefinitionId)
	if err != nil {
		helper.WriteToConsole("could not scan buildDefinition in buildExecutionShowHandler")
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
