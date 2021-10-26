package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"

	"github.com/gorilla/mux"
)

// BuildDefinitionListHandler lists all existing build definitions
func BuildDefinitionListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionListHandler: could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseservice.New()
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

// BuildDefinitionAddHandler adds a new build definition
func BuildDefinitionAddHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionservice.GetUserFromSession(session)
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
			Caption:   caption,
			Token:     security.GenerateToken(20),
			Content:   content,
			CreatedBy: currentUser.Id,
		}

		ds := databaseservice.New()
		_, err := ds.AddBuildDefinition(bd)
		if err != nil {
			helper.WriteToConsole("BuildDefinitionAddHandler: could not insert build definition " + err.Error())
			w.WriteHeader(500)
			return
		}

		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	skeleton, err := assets.GetMiscFile("build_definition_skeleton.yml")
	if err != nil {
		helper.WriteToConsole("BuildDefinitionAddHandler: could not get definition skeleton")
		return
	}

	data := struct {
		CurrentUser entity.User
		Skeleton    string
	}{
		CurrentUser: currentUser,
		Skeleton:    string(skeleton),
	}

	if err := templateservice.ExecuteTemplate(w, "builddefinition_add.html", data); err != nil {
		w.WriteHeader(404)
	}
}

// BuildDefinitionEditHandler allows for editing an existing build definition
func BuildDefinitionEditHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID in buildDefinitionEditHandler")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	sessMgr := global.GetSessionManager()
	ds := databaseservice.New()
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
			Id:       id,
			Caption:  caption,
			Content:  content,
			EditedBy: currentUser.Id,
			EditedAt: sql.NullTime{
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

// BuildDefinitionShowHandler shows details of a build definition
func BuildDefinitionShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionShowHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	ds := databaseservice.New()
	//defer ds.Quit()

	settings, err := ds.GetAllSettings()
	if err != nil {
		helper.WriteToConsole("BuildDefinitionShowHandler: could not fetch settings: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	baseUrl, ok := settings["base_url"]
	if !ok {
		helper.WriteToConsole("BuildDefinitionShowHandler: could not get setting base_url")
		baseUrl = "http://127.0.0.1:8271"
	}

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

	beList, err := ds.GetNewestBuildExecutions(10, "build_definition_id = ?", bd.Id)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionShowHandler: could not get newest build executions: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		BaseUrl           string
	}{
		BuildDefinition:   bd,
		RecentExecutions:  recentExecutions,
		CurrentUser:       currentUser,
		TotalBuildCount:   len(beList),
		FailedBuildCount:  failedBuildCount,
		SuccessBuildCount: successBuildCount,
		SuccessRate:       fmt.Sprintf("%.2f", successRate),
		AvgRuntime:        fmt.Sprintf("%.2f", avg),
		BaseUrl:           baseUrl,
	}

	if err := templateservice.ExecuteTemplate(w, "builddefinition_show.html", data); err != nil {
		w.WriteHeader(404)
	}
}

// BuildDefinitionRemoveHandler removes an existing build definition
func BuildDefinitionRemoveHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID in buildDefinitionEditHandler")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseservice.New()
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

// BuildDefinitionRestartHandler restarts the build process for a given build definition
func BuildDefinitionRestartHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID in buildDefinitionEditHandler")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseservice.New()
	//defer ds.Quit()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("BuildDefinitionRestartHandler: could not parse build definition id")
		http.Redirect(w, r, "/builddefinition/list", http.StatusBadRequest)
		return
	}

	bd, err := ds.GetBuildDefinitionById(id)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionRestartHandler: could not get buildDefinition: " + err.Error())
		http.Redirect(w, r, "/builddefinition/list", http.StatusBadRequest)
		return
	}

	variables, err := ds.GetAvailableVariablesForUser(currentUser.Id)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionRestartHandler: could not get variables: " + err.Error())
		http.Redirect(w, r, fmt.Sprintf("/builddefinition/%d/show", bd.Id), http.StatusBadRequest)
		return
	}

	cont, err := helper.UnmarshalBuildDefinitionContent(bd.Content, variables)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionRestartHandler: could not unmarshal build definition content: " + err.Error())
		http.Redirect(w, r, fmt.Sprintf("/builddefinition/%d/show", bd.Id), http.StatusBadRequest)
		return
	}

	go buildservice.StartBuildProcess(bd, cont)

	http.Redirect(w, r, fmt.Sprintf("/builddefinition/%d/show", bd.Id), http.StatusOK)
}
