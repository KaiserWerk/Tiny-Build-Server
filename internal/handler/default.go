package handler

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

// IndexHandler serves the dashboard
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		helper.WriteToConsole("dashboard: redirect to login: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionservice.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	ds := databaseservice.New()
	//defer ds.Quit()

	latestBuilds, err := ds.GetNewestBuildExecutions(5, "")
	if err != nil {
		helper.WriteToConsole("could not fetch latest build executions: " + err.Error())
	}
	latestBuildDefs, err := ds.GetNewestBuildDefinitions(5)
	if err != nil {
		helper.WriteToConsole("could not fetch latest build definitions: " + err.Error())
	}

	data := struct {
		CurrentUser     entity.User
		LatestBuilds    []entity.BuildExecution
		LatestBuildDefs []entity.BuildDefinition
	}{
		CurrentUser:     currentUser,
		LatestBuilds:    latestBuilds,
		LatestBuildDefs: latestBuildDefs,
	}

	if err := templateservice.ExecuteTemplate(w, "index.html", data); err != nil {
		fmt.Println("Error: " + err.Error())
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// StaticAssetHandler serves static file. http.FileServer does not work as desired
func StaticAssetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]

	var path string
	switch true {
	case strings.Contains(r.URL.Path, "assets/"):
		path = "assets"
	case strings.Contains(r.URL.Path, "js/"):
		path = "js"
	case strings.Contains(r.URL.Path, "css/"):
		path = "css"
	}

	data, err := internal.FSByte(false, "/public/"+path+"/"+file)
	if err != nil {
		helper.WriteToConsole("could not locate asset " + file)
		w.WriteHeader(404)
		return
	}

	var ext string
	if strings.Contains(file, ".") {
		parts := strings.Split(file, ".")
		ext = parts[len(parts)-1]
	}

	var contentType string // = http.DetectContentType(data)
	switch ext {
	case "css":
		contentType = "text/css"
	case "js":
		contentType = "text/javascript"
	case "html":
		contentType = "text/html"
	case "jpg":
		fallthrough
	case "jpeg":
		contentType = "image/jpeg"
	case "gif":
		contentType = "image/gif"
	case "png":
		contentType = "image/png"
	case "svg":
		contentType = "image/svg+xml"
	default:
		contentType = "text/plain"
	}
	w.Header().Set("Content-Type", contentType)

	_, _ = w.Write(data)
}
