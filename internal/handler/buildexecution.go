package handler

import (
	"Tiny-Build-Server/internal/entity"
	"Tiny-Build-Server/internal/helper"
	"Tiny-Build-Server/internal/security"
	"Tiny-Build-Server/internal/templates"
	"github.com/gorilla/mux"
	"net/http"
)

func BuildExecutionListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := helper.GetUserFromSession(session)
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

	if err := templates.ExecuteTemplate(w, "buildexecution_list.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func BuildExecutionShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := helper.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	db := helper.GetDbConnection()
	vars := mux.Vars(r)

	var be entity.BuildExecution
	row := db.QueryRow("SELECT id, build_definition_id, initiated_by, manual_run, action_log,"+
		" result, artifact_path, execution_time, executed_at FROM build_execution WHERE id = ?", vars["id"])
	err = row.Scan(&be.Id, &be.BuildDefinitionId, &be.InitiatedBy, &be.ManualRun, &be.ActionLog,
		&be.Result, &be.ArtifactPath, &be.ExecutionTime, &be.ExecutedAt)
	if err != nil {
		helper.WriteToConsole("could not scan buildExecution in buildExecutionShowHandler")
		w.WriteHeader(500)
		return
	}

	//be.ActionLog = strings.ReplaceAll(be.ActionLog, "\n", "<br>")

	var bd entity.BuildDefinition
	row = db.QueryRow("SELECT id, caption FROM build_definition WHERE id = ?", be.BuildDefinitionId)
	err = row.Scan(&bd.Id, &bd.Caption)
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
		BuildExecution:  be,
		BuildDefinition: bd,
	}

	if err = templates.ExecuteTemplate(w, "buildexecution_show.html", data); err != nil {
		w.WriteHeader(404)
	}
}
