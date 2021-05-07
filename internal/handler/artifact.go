package handler

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionservice"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

// DownloadNewestArtifactHandler downloads the most recently created version
// of a produced artifact
func DownloadNewestArtifactHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_, err = sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("DownloadNewestArtifactHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)

	ds := databaseservice.New()

	beList, err := ds.GetNewestBuildExecutions(1, "build_definition_id = ?", vars["id"])
	if err != nil {
		helper.WriteToConsole("DownloadNewestArtifactHandler: could not fetch build executions by definition id: " + err.Error())
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	if len(beList) < 1 {
		helper.WriteToConsole("DownloadNewestArtifactHandler: could not find any build executions for definition: " + err.Error())
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	artifact, err := filepath.Abs(beList[0].ArtifactPath)
	if err != nil {
		helper.WriteToConsole("DownloadNewestArtifactHandler: could not determine absolute path of file: " + err.Error())
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	artifact += ".zip"

	//fmt.Printf("file to serve: %s\n", artifact)

	cont, err := ioutil.ReadFile(artifact)
	if err != nil {
		helper.WriteToConsole("DownloadNewestArtifactHandler: could not read artifact file: " + err.Error())
		http.Redirect(w, r, "/builddefinition/list", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filepath.Base(artifact)))
	w.Write(cont)
}

// DownloadSpecificArtifactHandler downloads a artifact produced by a specific
// build execution
func DownloadSpecificArtifactHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_, err = sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("BuildDefinitionShowHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fmt.Fprint(w, "hey")
}
