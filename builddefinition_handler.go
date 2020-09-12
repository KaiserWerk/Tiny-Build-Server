package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func buildDefinitionListHandler(w http.ResponseWriter, r *http.Request) {
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

	type preparedBuildDefinition struct {
		Id					int
		Caption				string
		Target				string
		Executions			int
		RepoHost			string
		RepoName			string
		Enabled				bool
		DeploymentEnabled	bool
	}

	var bdList []preparedBuildDefinition
	var bd preparedBuildDefinition
	rows, err := db.Query("SELECT bd.id, bd.caption, bt.description AS target, COUNT(be.id), bd.repo_hoster," +
		" bd.repo_fullname, bd.enabled, bd.deployment_enabled FROM build_definition bd, build_target bt, build_execution " +
		"be WHERE bd.build_target_id = bt.id GROUP BY bd.id")
	if err != nil {
		writeToConsole("could not query build definitions in buildDefinitionListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	for rows.Next() {
		err = rows.Scan(&bd.Id, &bd.Caption, &bd.Target, &bd.Executions, &bd.RepoHost, &bd.RepoName, &bd.Enabled,
			&bd.DeploymentEnabled)
		if err != nil {
			writeToConsole("could not scan buildDefinition in buildDefinitionListHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		bdList = append(bdList, bd)
		bd = preparedBuildDefinition{}
	}

	data := struct {
		CurrentUser			user
		BuildDefinitions	[]preparedBuildDefinition
	}{
		CurrentUser: 		currentUser,
		BuildDefinitions: 	bdList,
	}

	t := templates["builddefinition_list.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionAddHandler(w http.ResponseWriter, r *http.Request) {

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

	data := struct {
		CurrentUser user
	} {
		CurrentUser: currentUser,
	}

	t := templates["builddefinition_add.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionEditHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["builddefinition_edit.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionShowHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r) // id
	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var bd buildDefinition
	row := db.QueryRow("SELECT * FROM build_definition WHERE id = ?", vars["id"])
	err = row.Scan(&bd.Id, &bd.BuildTargetId, &bd.AlteredBy, &bd.Caption, &bd.Enabled, &bd.DeploymentEnabled, &bd.RepoHoster, &bd.RepoHosterUrl,
		&bd.RepoFullname, &bd.RepoUsername, &bd.RepoSecret, &bd.RepoBranch, &bd.AlteredAt, &bd.MetaMigrationId)
	if err != nil {
		writeToConsole("could not scan buildDefinition: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var be buildExecution
	var beList = make([]buildExecution, 0)
	rows, err := db.Query("SELECT id, result, execution_time, executed_at FROM build_execution WHERE " +
		"build_definition_id = ? ORDER BY executed_at DESC", bd.Id)
	if err != nil {
		writeToConsole("could not fetch most recent build executions: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		err = rows.Scan(&be.Id, &be.Result, &be.ExecutionTime, &be.ExecutedAt)
		if err != nil {
			writeToConsole("could not scan build execution: " + err.Error())
			continue
		}
		beList = append(beList, be)
		be = buildExecution{}
	}

	failedBuildCount := 0
	successBuildCount := 0
	avg := 0.0
	i := 0
	for _, v := range beList {
		if v.Result == "success" {
			successBuildCount++
		}
		if v.Result == "failed" {
			failedBuildCount++
		}
		avg += v.ExecutionTime
		i++
	}

	avg = avg / float64(i)
	successRate := float64(successBuildCount) / float64(i) * 100
	var recentExecutions []buildExecution
	if len(beList) >= 5 {
		recentExecutions = beList[:5]
	} else {
		for _, v := range beList {
			recentExecutions = append(recentExecutions, v)
		}
	}

	data := struct {
		BuildDefinition buildDefinition
		RecentExecutions []buildExecution
		CurrentUser user
		TotalBuildCount int
		FailedBuildCount int
		SuccessBuildCount int
		SuccessRate	string
		AvgRuntime string
	} {
		BuildDefinition: bd,
		RecentExecutions: recentExecutions,
		CurrentUser: currentUser,
		TotalBuildCount: len(beList),
		FailedBuildCount: failedBuildCount,
		SuccessBuildCount: successBuildCount,
		SuccessRate: fmt.Sprintf("%.2f", successRate),
		AvgRuntime: fmt.Sprintf("%.2f", avg),
	}

	t := templates["builddefinition_show.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		writeToConsole("template build_definition_show.html not found")
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionRemoveHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["builddefinition_remove.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionListExecutionsHandler(w http.ResponseWriter, r *http.Request) {

}
