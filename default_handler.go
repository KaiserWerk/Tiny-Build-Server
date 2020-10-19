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

	//t := templates["index.html"]
	//if t != nil {
	//	err := t.Execute(w, indexData)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
}

func staticAssetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]
	data, err := Asset("public/" + file)
	if err != nil {
		fmt.Println("could not locate asset", file)
		w.Write([]byte("error"))
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
	fmt.Println("content-type:", contentType)
	w.Header().Set("Content-Type", contentType)

	w.Write(data)
}
