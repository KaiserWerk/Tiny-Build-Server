package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
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

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection in buildDefinitionAddHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	btList, err := getBuildTargets()
	if err != nil {
		writeToConsole("could not fetch buildTargets in buildDefinitionAddHandler")
		w.WriteHeader(500)
		return
	}

	if r.Method == http.MethodPost {
		targetId := r.FormValue("target_id")
		caption := r.FormValue("caption")
		var enabled bool
		if r.FormValue("enabled") == "1" {
			enabled = true
		}

		repoHoster := r.FormValue("repo_hoster")
		repoHosterUrl := r.FormValue("repo_hoster_url")
		repoFullname := r.FormValue("repo_fullname")
		repoUsername := r.FormValue("repo_username")
		repoSecret := r.FormValue("repo_secret")
		repoBranch := r.FormValue("repo_branch")

		action := r.FormValue("action")

		result, err := db.Exec("INSERT INTO build_definition (build_target_id, altered_by, caption, enabled, " +
			"deployment_enabled, repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, " +
			"repo_branch, altered_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			targetId, currentUser.Id, caption, enabled, 0, repoHoster, repoHosterUrl, repoFullname, repoUsername,
			repoSecret, repoBranch, time.Now())
		if err != nil {
			writeToConsole("could not insert build definition in buildDefinitionAddHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		liid, err := result.LastInsertId()
		if err != nil {
			writeToConsole("could not get lastInsertId in buildDefinitionAddHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}

		err = r.ParseForm() // Required if you don't call r.FormValue()
		if err != nil {
			writeToConsole("could not parse form in buildDefinitionAddHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}

		// add build step references to newly inserted build definition
		buildSteps := r.Form["build_steps"]
		for _, v := range buildSteps {
			_, err = db.Exec("INSERT INTO definition_step_taxonomy (build_definition_id, build_step_id) VALUES (?, ?)",
				liid, v)
			if err != nil {
				writeToConsole("could not insert taxonomy entry in buildDefinitionAddHandler: " + err.Error())
				continue
			}
		}

		if action == "save_depl" {
			writeToConsole("redirect to edit deployments")
			http.Redirect(w, r, "/builddefinition/" + strconv.Itoa(int(liid)) + "/edit?tab=deployments", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	var selectedTarget int
	temp := r.URL.Query().Get("target")
	if temp != "" {
		selectedTarget, _ = strconv.Atoi(temp)
	}

	bsList, err := getBuildStepsForTarget(selectedTarget)
	if err != nil {
		writeToConsole("could not fetch buildSteps in buildDefinitionAddHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	data := struct {
		CurrentUser 			user
		BuildTargets			[]buildTarget
		SelectedTarget			int
		AvailableBuildSteps		[]buildStep
	} {
		CurrentUser: 			currentUser,
		BuildTargets:   		btList,
		SelectedTarget: 		selectedTarget,
		AvailableBuildSteps: 	bsList,
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
	rows, err := db.Query("SELECT id, build_definition_id, initiated_by, manual_run, result, execution_time, executed_at FROM build_execution WHERE " +
		"build_definition_id = ? ORDER BY executed_at DESC", bd.Id)
	if err != nil {
		writeToConsole("could not fetch most recent build executions: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		err = rows.Scan(&be.Id, &be.BuildDefinitionId, &be.InitiatedBy, &be.ManualRun, &be.Result, &be.ExecutionTime, &be.ExecutedAt)
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
