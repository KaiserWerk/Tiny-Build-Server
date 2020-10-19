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
		Id                int
		Caption           string
		Target            string
		TargetOsArch	string
		Executions        int
		RepoHost          string
		RepoName          string
		Enabled           bool
		DeploymentEnabled bool
	}

	var bdList []preparedBuildDefinition
	var bd preparedBuildDefinition
	rows, err := db.Query("SELECT bd.id, bd.caption, bd.build_target, bd.build_target_os_arch, COUNT(be.id), " +
		"bd.repo_hoster, bd.repo_fullname, bd.enabled, bd.deployment_enabled FROM build_definition bd LEFT JOIN " +
		"build_execution be ON be.build_definition_id = bd.id GROUP BY bd.id ORDER BY bd.caption")
	if err != nil {
		writeToConsole("could not query build definitions in buildDefinitionListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	for rows.Next() {
		err = rows.Scan(&bd.Id, &bd.Caption, &bd.Target, &bd.TargetOsArch, &bd.Executions, &bd.RepoHost,
			&bd.RepoName, &bd.Enabled, &bd.DeploymentEnabled)
		if err != nil {
			writeToConsole("could not scan buildDefinition in buildDefinitionListHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		bdList = append(bdList, bd)
		bd = preparedBuildDefinition{}
	}

	data := struct {
		CurrentUser      user
		BuildDefinitions []preparedBuildDefinition
	}{
		CurrentUser:      currentUser,
		BuildDefinitions: bdList,
	}

	if err := executeTemplate(w, "builddefinition_list.html", data); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["builddefinition_list.html"]
	//if t != nil {
	//	err := t.Execute(w, data)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
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

	//btList, err := getBuildTargets()
	//if err != nil {
	//	writeToConsole("could not fetch buildTargets in buildDefinitionAddHandler")
	//	w.WriteHeader(500)
	//	return
	//}

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

		var applyMigrations bool
		if r.FormValue("apply_migrations") == "1" {
			applyMigrations = true
		}
		databaseDsn := r.FormValue("database_dsn")
		var runTests bool
		if r.FormValue("run_tests") == "1" {
			runTests = true
		}
		var runBenchmarkTests bool
		if r.FormValue("run_benchmark_tests") == "1" {
			runBenchmarkTests = true
		}

		action := r.FormValue("action")

		result, err := db.Exec("INSERT INTO build_definition (build_target_id, altered_by, caption, enabled, "+
			"deployment_enabled, repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, "+
			"repo_branch, altered_at, apply_migrations, database_dsn, run_tests, run_benchmark_tests) " +
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			targetId, currentUser.Id, caption, enabled, 0, repoHoster, repoHosterUrl, repoFullname, repoUsername,
			repoSecret, repoBranch, time.Now(), applyMigrations, databaseDsn, runTests, runBenchmarkTests)
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

		//err = r.ParseForm() // Required if you don't call r.FormValue()
		//if err != nil {
		//	writeToConsole("could not parse form in buildDefinitionAddHandler: " + err.Error())
		//	w.WriteHeader(500)
		//	return
		//}


		if action == "save_depl" {
			writeToConsole("redirect to edit deployments")
			http.Redirect(w, r, "/builddefinition/"+strconv.Itoa(int(liid))+"/edit?tab=deployments", http.StatusSeeOther)
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

	var runtimes []string
	switch selectedTarget {
	case 1:
		runtimes = golangRuntimes
	case 2:
		runtimes = dotnetRuntimes
	case 3:
		// php does not have runtimes
	case 4:
		// rust + cross-compile? would be nice
	}

	data := struct {
		CurrentUser         user
		SelectedTarget      int
		AvailableRuntimes   []string
	}{
		CurrentUser:       currentUser,
		SelectedTarget:    selectedTarget,
		AvailableRuntimes: runtimes,
	}

	if err := executeTemplate(w, "builddefinition_add.html", data); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["builddefinition_add.html"]
	//if t != nil {
	//	err := t.Execute(w, data)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
}

func buildDefinitionEditHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID in buildDefinitionEditHandler")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection in buildDefinitionEditHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	vars := mux.Vars(r)

	if r.Method == http.MethodPost {

	}

	var bdt buildDefinition
	row := db.QueryRow("SELECT id, build_target, altered_by, caption, enabled, deployment_enabled, "+
		"repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, repo_branch, altered_at, "+
		"apply_migrations, database_dsn, meta_migration_id, run_tests, run_benchmark_tests "+
		"FROM build_definition WHERE id = ?", vars["id"])
	err = row.Scan(&bdt.Id, &bdt.BuildTargetId, &bdt.AlteredBy, &bdt.Caption, &bdt.Enabled, &bdt.DeploymentEnabled,
		&bdt.RepoHoster, &bdt.RepoHosterUrl, &bdt.RepoFullname, &bdt.RepoUsername, &bdt.RepoSecret, &bdt.RepoBranch,
		&bdt.AlteredAt, &bdt.ApplyMigrations, &bdt.DatabaseDSN, &bdt.MetaMigrationId, &bdt.RunTests,
		&bdt.RunBenchmarkTests,
	)
	if err != nil {
		writeToConsole("could not scan buildDefinition in buildDefinitionEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	selectedTab := r.URL.Query().Get("tab")

	var runtimes []string
	switch bdt.BuildTargetId {
	case 1:
		runtimes = golangRuntimes
	case 2:
		runtimes = dotnetRuntimes
	case 3:
		// php does not have runtimes
	case 4:
		// rust + cross-compile? would be nice
	}

	data := struct {
		CurrentUser             user
		SelectedBuildDefinition buildDefinition
		SelectedTab             string
		AvailableRuntimes   	[]string
	}{
		CurrentUser:             currentUser,
		SelectedBuildDefinition: bdt,
		SelectedTab:             selectedTab,
		AvailableRuntimes:		 runtimes,
	}

	if err := executeTemplate(w, "builddefinition_edit.html", data); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["builddefinition_edit.html"]
	//if t != nil {
	//	err := t.Execute(w, data)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
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
	row := db.QueryRow("SELECT id, build_target, altered_by, caption, enabled, deployment_enabled, repo_hoster, "+
		"repo_hoster_url, repo_fullname, repo_username, repo_secret, repo_branch, altered_at FROM build_definition WHERE id = ?", vars["id"])
	err = row.Scan(&bd.Id, &bd.BuildTargetId, &bd.AlteredBy, &bd.Caption, &bd.Enabled, &bd.DeploymentEnabled, &bd.RepoHoster, &bd.RepoHosterUrl,
		&bd.RepoFullname, &bd.RepoUsername, &bd.RepoSecret, &bd.RepoBranch, &bd.AlteredAt, &bd.MetaMigrationId)
	if err != nil {
		writeToConsole("could not scan buildDefinition: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var be buildExecution
	var beList = make([]buildExecution, 0)
	rows, err := db.Query("SELECT id, build_definition_id, initiated_by, manual_run, result, execution_time, executed_at FROM build_execution WHERE "+
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
		BuildDefinition   buildDefinition
		RecentExecutions  []buildExecution
		CurrentUser       user
		TotalBuildCount   int
		FailedBuildCount  int
		SuccessBuildCount int
		SuccessRate       string
		AvgRuntime        string
	}{
		BuildDefinition:   bd,
		RecentExecutions:  recentExecutions,
		CurrentUser:       currentUser,
		TotalBuildCount:   len(beList),
		FailedBuildCount:  failedBuildCount,
		SuccessBuildCount: successBuildCount,
		SuccessRate:       fmt.Sprintf("%.2f", successRate),
		AvgRuntime:        fmt.Sprintf("%.2f", avg),
	}

	if err := executeTemplate(w, "builddefinition_show.html", data); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["builddefinition_show.html"]
	//if t != nil {
	//	err := t.Execute(w, data)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	writeToConsole("template build_definition_show.html not found")
	//	w.WriteHeader(http.StatusNotFound)
	//}
}

func buildDefinitionRemoveHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID in buildDefinitionEditHandler")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection in buildDefinitionEditHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	vars := mux.Vars(r)

	confirm := r.URL.Query().Get("confirm")
	if confirm == "yes" {

	}

	var buildDefinitionTemp buildDefinition
	row := db.QueryRow("SELECT id, build_target_id, altered_by, caption, enabled, deployment_enabled, "+
		"repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, repo_branch, altered_at, "+
		"meta_migration_id FROM build_definition WHERE id = ?", vars["id"])
	err = row.Scan(&buildDefinitionTemp.Id, &buildDefinitionTemp.BuildTargetId, &buildDefinitionTemp.AlteredBy,
		&buildDefinitionTemp.Caption, &buildDefinitionTemp.Enabled, &buildDefinitionTemp.DeploymentEnabled,
		&buildDefinitionTemp.RepoHoster, &buildDefinitionTemp.RepoHosterUrl, &buildDefinitionTemp.RepoFullname,
		&buildDefinitionTemp.RepoUsername, &buildDefinitionTemp.RepoSecret, &buildDefinitionTemp.RepoBranch,
		&buildDefinitionTemp.AlteredAt, &buildDefinitionTemp.MetaMigrationId)
	if err != nil {
		writeToConsole("could not scan buildDefinition in buildDefinitionEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	data := struct {
		CurrentUser     user
		BuildDefinition buildDefinition
	}{
		CurrentUser:     currentUser,
		BuildDefinition: buildDefinitionTemp,
	}

	if err := executeTemplate(w, "builddefinition_remove.html", data); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["builddefinition_remove.html"]
	//if t != nil {
	//	err := t.Execute(w, data)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
}

func buildDefinitionListExecutionsHandler(w http.ResponseWriter, r *http.Request) {

}
