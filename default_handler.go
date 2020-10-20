package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
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
	latestBuilds, err := getNewestBuildExecutions(5)
	if err != nil {
		writeToConsole("could not fetch latest build executions: " + err.Error())
	}
	latestBuildDefs, err := getNewestBuildDefinitions(5)
	if err != nil {
		writeToConsole("could not fetch latest build executions: " + err.Error())
	}

	indexData := struct {
		CurrentUser     user
		LatestBuilds    []buildExecution
		LatestBuildDefs []buildDefinition
	}{
		CurrentUser:     currentUser,
		LatestBuilds:    latestBuilds,
		LatestBuildDefs: latestBuildDefs,
	}

	if err := executeTemplate(w, "index.html", indexData); err != nil {
		w.WriteHeader(404)
	}
}

func staticAssetHandler(w http.ResponseWriter, r *http.Request) {
	//writeToConsole("asset handler hit")
	vars := mux.Vars(r)
	file := vars["file"]

	var path string
	switch true {
	case strings.Contains(r.URL.Path, "assets"):
		path = "assets"
	case strings.Contains(r.URL.Path, "js"):
		path = "js"
	case strings.Contains(r.URL.Path, "css"):
		path = "css"
	}

	data, err := Asset("public/" + path + "/" + file)
	if err != nil {
		fmt.Println("could not locate asset", file)
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
	default:
		contentType = "text/plain"
	}
	w.Header().Set("Content-Type", contentType)

	_, _ = w.Write(data)
}
