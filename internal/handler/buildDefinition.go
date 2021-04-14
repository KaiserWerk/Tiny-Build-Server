package handler

import (
	"database/sql"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionService"
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
		helper.WriteToConsole("BuildDefinitionListHandler: could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseService.New()
	buildDefinitions, err := ds.GetAllBuildDefinitions()

	data := struct {
		CurrentUser      entity.User
		BuildDefinitions []entity.BuildDefinition
	}{
		CurrentUser:      currentUser,
		BuildDefinitions: buildDefinitions,
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

	sessMgr := global.GetSessionManager()

	if r.Method == http.MethodPost {

		caption := r.FormValue("caption")
		content := r.FormValue("content")

		if caption == "" || content == "" {
			helper.WriteToConsole("BuildDefinitionAddHandler: missing required fields")
			sessMgr.AddMessage("info", "Fields caption and content cannot be empty")
			http.Redirect(w, r, "/builddefinition/add", http.StatusSeeOther)
			return
		}

		bd := entity.BuildDefinition{
			Caption:         caption,
			Content:         content,
			CreatedBy:       currentUser.Id,
		}

		ds := databaseService.New()
		_, err := ds.AddBuildDefinition(bd)
		if err != nil {
			helper.WriteToConsole("BuildDefinitionAddHandler: could not insert build definition " + err.Error())
			w.WriteHeader(500)
			return
		}

		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	skeleton, err := internal.FSString(true, "/templates/misc/build_definition_skeleton.yml")
	if err != nil {
		helper.WriteToConsole("BuildDefinitionAddHandler: could not get definition skeleton")
		return
	}

	data := struct {
		CurrentUser entity.User
		Skeleton    string
	}{
		CurrentUser: currentUser,
		Skeleton:    skeleton,
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
	sessMgr := global.GetSessionManager()
	ds := databaseService.New()
	//defer ds.Quit()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("BuildDefinitionEditHandler: could not parse build definition id")
		id = -1
	}

	if r.Method == http.MethodPost {
		caption := r.FormValue("caption")
		content := r.FormValue("content")

		if caption == "" || content == "" {
			helper.WriteToConsole("BuildDefinitionEditHandler: required fields missing")
			sessMgr.AddMessage("warning", "Please fill in required fields.")
			http.Redirect(w, r, fmt.Sprintf("/builddefinition/%s/edit", vars["id"]), http.StatusSeeOther)
			return
		}

		bd := entity.BuildDefinition{
			Id:              id,
			Caption:         caption,
			Content:         content,
			EditedBy:        currentUser.Id,
			EditedAt:        sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
		}

		err = ds.UpdateBuildDefinition(bd)
		if err != nil {
			helper.WriteToConsole("BuildDefinitionEditHandler: could not save updated build definition: " + err.Error())
			sessMgr.AddMessage("error", "An unknown error occurred! Please try again.")
			http.Redirect(w, r, fmt.Sprintf("/builddefinition/%s/edit", vars["id"]), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	bdt, err := ds.GetBuildDefinitionById(id)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionEditHandler: could not get buildDefinition: " + err.Error())
		w.WriteHeader(500)
		return
	}

	data := struct {
		CurrentUser     entity.User
		BuildDefinition entity.BuildDefinition
	}{
		CurrentUser:     currentUser,
		BuildDefinition: bdt,
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
	//defer ds.Quit()

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
	//defer ds.Quit()

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
