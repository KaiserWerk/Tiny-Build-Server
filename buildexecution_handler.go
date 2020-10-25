package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func buildExecutionListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection in buildDefinitionListHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	data := struct {
		CurrentUser		user
	}{
		CurrentUser: 	currentUser,
	}

	if err := executeTemplate(w, "buildexecution_list.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func buildExecutionShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection in buildDefinitionListHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	vars := mux.Vars(r)

	var be buildExecution
	row := db.QueryRow("SELECT id, build_definition_id, initiated_by, manual_run, action_log," +
		" result, artifact_path, execution_time, executed_at FROM build_execution WHERE id = ?", vars["id"])
	err = row.Scan(&be.Id, &be.BuildDefinitionId, &be.InitiatedBy, &be.ManualRun, &be.ActionLog,
		&be.Result, &be.ArtifactPath, &be.ExecutionTime, &be.ExecutedAt)
	if err != nil {
		writeToConsole("could not scan buildExecution in buildExecutionShowHandler")
		w.WriteHeader(500)
		return
	}

	var bd buildDefinition
	row = db.QueryRow("SELECT id, caption FROM build_definition WHERE id = ?", be.BuildDefinitionId)
	err = row.Scan(&bd.Id, &bd.Caption)
	if err != nil {
		writeToConsole("could not scan buildDefinition in buildExecutionShowHandler")
		w.WriteHeader(500)
		return
	}

	data := struct {
		CurrentUser		user
		BuildExecution  buildExecution
		BuildDefinition buildDefinition
	}{
		CurrentUser: 	currentUser,
		BuildExecution: be,
		BuildDefinition: bd,
	}

	if err = executeTemplate(w, "buildexecution_show.html", data); err != nil {
		w.WriteHeader(404)
	}
}
