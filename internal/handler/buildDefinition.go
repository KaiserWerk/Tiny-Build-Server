package handler

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"

	"github.com/gorilla/mux"
)

// BuildDefinitionListHandler lists all existing build definitions
func (h *HttpHandler) BuildDefinitionListHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("BuildDefinitionListHandler")
	)
	buildDefinitions, err := h.Ds.GetAllBuildDefinitions()
	if err != nil {
		http.Error(w, "could not get all build definitions", http.StatusInternalServerError)
		logger.WithField("error", err.Error()).Error("could not get all build definitions")
		return
	}

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
func (h *HttpHandler) BuildDefinitionAddHandler(w http.ResponseWriter, r *http.Request) {
	var (
		sessMgr     = h.SessMgr
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("BuildDefinitionAddHandler")
	)

	if r.Method == http.MethodPost {
		caption := r.FormValue("caption")
		content := r.FormValue("content")

		if caption == "" || content == "" {
			logger.Info("missing required fields")
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

		_, err := h.Ds.AddBuildDefinition(bd)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not insert build definition")
			w.WriteHeader(500)
			return
		}

		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	skeleton, err := assets.GetMiscFile("build_definition_skeleton.yml")
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get definition skeleton")
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
func (h *HttpHandler) BuildDefinitionEditHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		sessMgr     = h.SessMgr
		vars        = mux.Vars(r)
		logger      = h.ContextLogger("BuildDefinitionEditHandler")
	)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not parse build definition id")
		id = -1
	}

	if r.Method == http.MethodPost {
		caption := r.FormValue("caption")
		content := r.FormValue("content")

		if caption == "" || content == "" {
			logger.WithField("error", err.Error()).Error("required fields missing")
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

		err = h.Ds.UpdateBuildDefinition(bd)
		if err != nil {
			logger.WithField("error", err.Error()).Error("BuildDefinitionEditHandler: could not save updated build definition: " + err.Error())
			sessMgr.AddMessage("error", "An unknown error occurred! Please try again.")
			http.Redirect(w, r, fmt.Sprintf("/builddefinition/%s/edit", vars["id"]), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	bdt, err := h.Ds.GetBuildDefinitionById(id)
	if err != nil {
		logger.WithField("error", err.Error()).Error("BuildDefinitionEditHandler: could not get buildDefinition: " + err.Error())
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
func (h *HttpHandler) BuildDefinitionShowHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("BuildDefinitionShowHandler")
		vars        = mux.Vars(r)
		baseUrl     string
		limit       int = 25
	)

	settings, err := h.Ds.GetAllSettings()
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not fetch settings")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	baseUrl, ok := settings["base_url"]
	if !ok {
		logger.Info("could not get setting base_url; using default")
		baseUrl = "http://127.0.0.1:8271"
	}

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not parse build definition id, setting to -1")
		id = -1
	}
	bd, err := h.Ds.GetBuildDefinitionById(id)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get buildDefinition")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	beList, err := h.Ds.GetNewestBuildExecutions(limit, "build_definition_id = ?", bd.Id)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get newest build executions")
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
		Limit             int
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
		Limit:             limit,
	}

	if err := templateservice.ExecuteTemplate(w, "builddefinition_show.html", data); err != nil {
		w.WriteHeader(404)
	}
}

// BuildDefinitionRemoveHandler removes an existing build definition
func (h *HttpHandler) BuildDefinitionRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("BuildDefinitionRemoveHandler")
		vars        = mux.Vars(r)
	)

	confirm := r.URL.Query().Get("confirm")
	if confirm == "yes" {
		// TODO: implement "yes" action for build definition removal
	}

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    id,
		}).Error("could not parse entry ID")
		w.WriteHeader(500)
		return
	}
	buildDefinition, err := h.Ds.GetBuildDefinitionById(id)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    id,
		}).Error("could not get build definition by ID")
		http.Error(w, "could not get build definition by ID", http.StatusInternalServerError)
		return
	}

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

func (h *HttpHandler) BuildDefinitionListExecutionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement or scrap
}

// BuildDefinitionRestartHandler restarts the build process for a given build definition
func (h *HttpHandler) BuildDefinitionRestartHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("BuildDefinitionRestartHandler")
		vars        = mux.Vars(r)
	)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    id,
		}).Error("could not parse build definition ID")
		http.Redirect(w, r, "/builddefinition/list", http.StatusBadRequest)
		return
	}

	bd, err := h.Ds.GetBuildDefinitionById(id)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    id,
		}).Error("could not get buildDefinition")
		http.Redirect(w, r, "/builddefinition/list", http.StatusBadRequest)
		return
	}

	variables, err := h.Ds.GetAvailableVariablesForUser(currentUser.Id)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"userID": currentUser.Id,
		}).Error("could not get variables")
		http.Redirect(w, r, fmt.Sprintf("/builddefinition/%d/show", bd.Id), http.StatusBadRequest)
		return
	}

	cont, err := helper.UnmarshalBuildDefinitionContent(bd.Content, variables)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not unmarshal build definition content")
		http.Redirect(w, r, fmt.Sprintf("/builddefinition/%d/show", bd.Id), http.StatusBadRequest)
		return
	}

	go buildservice.StartBuildProcess(bd, cont)

	http.Redirect(w, r, fmt.Sprintf("/builddefinition/%d/show", bd.Id), http.StatusSeeOther)
}
