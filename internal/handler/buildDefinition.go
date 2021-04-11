package handler

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func BuildDefinitionListHandler(w http.ResponseWriter, r *http.Request) {
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

	db := global.GetDbConnection()
	type preparedBuildDefinition struct {
		Id                int
		Caption           string
		Target            string
		TargetOsArch      string
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
		helper.WriteToConsole("could not query build definitions in buildDefinitionListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	for rows.Next() {
		err = rows.Scan(&bd.Id, &bd.Caption, &bd.Target, &bd.TargetOsArch, &bd.Executions, &bd.RepoHost,
			&bd.RepoName, &bd.Enabled, &bd.DeploymentEnabled)
		if err != nil {
			helper.WriteToConsole("could not scan buildDefinition in buildDefinitionListHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		bdList = append(bdList, bd)
		bd = preparedBuildDefinition{}
	}

	data := struct {
		CurrentUser      entity.User
		BuildDefinitions []preparedBuildDefinition
	}{
		CurrentUser:      currentUser,
		BuildDefinitions: bdList,
	}

	if err := templateservice.ExecuteTemplate(w, "builddefinition_list.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func BuildDefinitionAddHandler(w http.ResponseWriter, r *http.Request) {
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

	db := global.GetDbConnection()

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
			"repo_branch, altered_at, apply_migrations, database_dsn, run_tests, run_benchmark_tests) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			targetId, currentUser.Id, caption, enabled, 0, repoHoster, repoHosterUrl, repoFullname, repoUsername,
			repoSecret, repoBranch, time.Now(), applyMigrations, databaseDsn, runTests, runBenchmarkTests)
		if err != nil {
			helper.WriteToConsole("could not insert build definition in buildDefinitionAddHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		liid, err := result.LastInsertId()
		if err != nil {
			helper.WriteToConsole("could not get lastInsertId in buildDefinitionAddHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}

		//err = r.ParseForm() // Required if you don't call r.FormValue()
		//if err != nil {
		//	helper.WriteToConsole("could not parse form in buildDefinitionAddHandler: " + err.Error())
		//	w.WriteHeader(500)
		//	return
		//}

		if action == "save_depl" {
			helper.WriteToConsole("redirect to edit deployments")
			http.Redirect(w, r, "/builddefinition/"+strconv.Itoa(int(liid))+"/edit?tab=deployments", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	data := struct {
		CurrentUser       entity.User
	}{
		CurrentUser:       currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "builddefinition_add.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func BuildDefinitionEditHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID in buildDefinitionEditHandler")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseService.New()
	defer ds.Quit()

	vars := mux.Vars(r)
	if r.Method == http.MethodPost {

	}

	// TODO: Rework into method
	//var bdt entity.BuildDefinition
	//row := db.QueryRow("SELECT id, build_target, altered_by, caption, enabled, deployment_enabled, "+
	//	"repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, repo_branch, altered_at, "+
	//	"apply_migrations, database_dsn, meta_migration_id, run_tests, run_benchmark_tests "+
	//	"FROM build_definition WHERE id = ?", vars["id"])
	//err = row.Scan(&bdt.Id, &bdt.BuildTarget, &bdt.AlteredBy, &bdt.Caption, &bdt.Enabled, &bdt.DeploymentEnabled,
	//	&bdt.RepoHoster, &bdt.RepoHosterUrl, &bdt.RepoFullname, &bdt.RepoUsername, &bdt.RepoSecret, &bdt.RepoBranch,
	//	&bdt.AlteredAt, &bdt.ApplyMigrations, &bdt.DatabaseDSN, &bdt.MetaMigrationId, &bdt.RunTests,
	//	&bdt.RunBenchmarkTests,
	//)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("BuildDefinitionEditHandler: could not parse build definition id, setting to -1")
		id = -1
	}
	bdt, err := ds.GetBuildDefinitionById(id)
	if err != nil {
		helper.WriteToConsole("could not scan buildDefinition in buildDefinitionEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	selectedTab := r.URL.Query().Get("tab")

	data := struct {
		CurrentUser             entity.User
		SelectedBuildDefinition entity.BuildDefinition
		SelectedTab             string
	}{
		CurrentUser:             currentUser,
		SelectedBuildDefinition: bdt,
		SelectedTab:             selectedTab,
	}

	if err := templateservice.ExecuteTemplate(w, "builddefinition_edit.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func BuildDefinitionShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("show build definition handler: could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r) // id
	ds := databaseService.New()
	defer ds.Quit()

	// TODO: rework into method
	//var bd entity.BuildDefinition
	//row := db.QueryRow("SELECT id, build_target, build_target_os_arch, build_target_arm, altered_by, caption, enabled, deployment_enabled, repo_hoster, "+
	//	"repo_hoster_url, repo_fullname, repo_username, repo_secret, repo_branch, altered_at FROM build_definition WHERE id = ?", vars["id"])
	//err = row.Scan(&bd.Id, &bd.BuildTarget, &bd.BuildTargetOsArch, &bd.BuildTargetArm, &bd.AlteredBy, &bd.Caption,
	//	&bd.Enabled, &bd.DeploymentEnabled, &bd.RepoHoster, &bd.RepoHosterUrl, &bd.RepoFullname, &bd.RepoUsername,
	//	&bd.RepoSecret, &bd.RepoBranch, &bd.AlteredAt)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("BuildDefinitionShowHandler: could not parse build definition id, setting to -1")
		id = -1
	}
	bd, err := ds.GetBuildDefinitionById(id)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionShowHandler: could not scan buildDefinition: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//var be entity.BuildExecution
	//var beList = make([]entity.BuildExecution, 0)
	//rows, err := db.Query("SELECT id, build_definition_id, initiated_by, manual_run, result, execution_time, executed_at FROM build_execution WHERE "+
	//	"build_definition_id = ? ORDER BY executed_at DESC", bd.Id)
	//if err != nil {
	//	helper.WriteToConsole("show build definition handler: could not fetch most recent build executions: " + err.Error())
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//
	//for rows.Next() {
	//	err = rows.Scan(&be.Id, &be.BuildDefinitionId, &be.InitiatedBy, &be.ManualRun, &be.Result, &be.ExecutionTime, &be.ExecutedAt)
	//	if err != nil {
	//		helper.WriteToConsole("show build definition handler: could not scan build execution: " + err.Error())
	//		continue
	//	}
	//	beList = append(beList, be)
	//	be = entity.BuildExecution{}
	//}

	beList, err := ds.FindBuildExecutions("build_definition_id = ?", bd.Id)

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
	var recentExecutions []entity.BuildExecution
	if len(beList) >= 5 {
		recentExecutions = beList[:5]
	} else {
		for _, v := range beList {
			recentExecutions = append(recentExecutions, v)
		}
	}

	data := struct {
		BuildDefinition   entity.BuildDefinition
		RecentExecutions  []entity.BuildExecution
		CurrentUser       entity.User
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

	if err := templateservice.ExecuteTemplate(w, "builddefinition_show.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func BuildDefinitionRemoveHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID in buildDefinitionEditHandler")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseService.New()
	defer ds.Quit()

	vars := mux.Vars(r)

	confirm := r.URL.Query().Get("confirm")
	if confirm == "yes" {
		// TODO: implement "yes" action for build definition removal
	}

	//var buildDefinition entity.BuildDefinition
	//row := db.QueryRow("SELECT id, build_target_id, altered_by, caption, enabled, deployment_enabled, "+
	//	"repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, repo_branch, altered_at, "+
	//	"meta_migration_id FROM build_definition WHERE id = ?", vars["id"])
	//err = row.Scan(&buildDefinition.Id, &buildDefinition.BuildTarget, &buildDefinition.AlteredBy,
	//	&buildDefinition.Caption, &buildDefinition.Enabled, &buildDefinition.DeploymentEnabled,
	//	&buildDefinition.RepoHoster, &buildDefinition.RepoHosterUrl, &buildDefinition.RepoFullname,
	//	&buildDefinition.RepoUsername, &buildDefinition.RepoSecret, &buildDefinition.RepoBranch,
	//	&buildDefinition.AlteredAt, &buildDefinition.MetaMigrationId)
	//if err != nil {
	//	helper.WriteToConsole("could not scan buildDefinition in buildDefinitionEditHandler: " + err.Error())
	//	w.WriteHeader(500)
	//	return
	//}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("BuildDefinitionRemoveHandler: could not parse entry ID: " + err.Error())
		w.WriteHeader(500)
		return
	}
	buildDefinition, err := ds.GetBuildDefinitionById(id)

	data := struct {
		CurrentUser     entity.User
		BuildDefinition entity.BuildDefinition
	}{
		CurrentUser:     currentUser,
		BuildDefinition: buildDefinition,
	}

	if err := templateservice.ExecuteTemplate(w, "builddefinition_remove.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func BuildDefinitionListExecutionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement or scrap
}

func BuildDefinitionRestartHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement or scrap
}
