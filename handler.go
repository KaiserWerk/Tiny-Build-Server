package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}

func bitBucketReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var payload BitBucketPushPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("could not decode request body"))
		return
	}

	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println("Could not parse query params")
		return
	}

	if len(queryParams["id"]) == 0 || len(queryParams["token"]) == 0 {
		fmt.Println("Missing r parameters")
		return
	}
	// AUCH HEADER PRÜFEN!

	headers := []string{"X-Event-Key", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
	headerValues := make([]string, len(headers))
	for i := range headers {
		headerValues[i], err = getHeaderIfSet(r, headers[i])
		if err != nil {
			log.Printf("could not get header %v\n", headers[i])
		}
	}

	// all strings
	id := queryParams["id"][0]
	token := queryParams["token"][0]
	branch := payload.Push.Changes[0].New.Name
	repoFullName := payload.Repository.FullName

	fmt.Println("id: " + id)
	fmt.Println("token: " + token)

	fmt.Println("branch: " + branch)
	fmt.Println("repo full name: " + repoFullName)

	buildDefinition, err := loadBuildDefinition(id)
	if err != nil {
		fmt.Println("error while loading build configuration:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("received, but bad request"))
		return
	}

	if buildDefinition.AuthToken != token {
		fmt.Println("no auth token match")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("received, but auth token mismatch"))
		return
	}

	if buildDefinition.Repository.FullName != repoFullName || buildDefinition.Repository.Branch != branch {
		fmt.Println("repo name or branch mismatch")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("recevied, but repository name/branch mismatch"))
	}

	// now we can start the build process
	// something like
	fmt.Println("starting build process for id " + id)
	go startBuildProcess(id, buildDefinition)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("received, everything fine"))
}

func gitHubReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var payload GitHubPushPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("could not decode request body"))
		return
	}
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println("Could not parse query params")
		return
	}

	if len(queryParams["id"]) == 0 || len(queryParams["token"]) == 0 {
		fmt.Println("Missing request parameters")
		return
	}

	headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
	headerValues := make([]string, len(headers))
	for i := range headers {
		headerValues[i], err = getHeaderIfSet(r, headers[i])
		if err != nil {
			log.Printf("could not get header %v\n", headers[i])
		}
	}

	// all strings
	id := queryParams["id"][0]
	token := queryParams["token"][0]
	repoFullName := payload.Repository.FullName
	branch := payload.Repository.DefaultBranch // other: MasterBranch

	buildConfig, err := loadBuildDefinition(id)
	if err != nil {
		fmt.Println("error while loading build configuration:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("received, but bad request"))
		return
	}

	if buildConfig.AuthToken != token {
		fmt.Println("no auth token match")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("received, but auth token mismatch"))
		return
	}

	if buildConfig.Repository.FullName != repoFullName || buildConfig.Repository.Branch != branch {
		fmt.Println("repo name or branch mismatch")
		w.WriteHeader(http.StatusBadRequest)
		str := fmt.Sprintf("repo name expected: %v, got %v instead; branch name expected: %v, got %v instead",
			buildConfig.Repository.FullName, repoFullName, buildConfig.Repository.Branch, branch)
		w.Write([]byte("recevied, but repository name/branch mismatch: " + str))
		return
	}

	// now we can start the build process
	// something like
	// go startBuildProcess(buildConfig, payload)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("received, build process initiated, everything fine"))
}

func gitLabReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var payload GitLabPushPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("could not decode request body"))
		return
	}
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println("Could not parse query params")
		return
	}

	if len(queryParams["id"]) == 0 || len(queryParams["token"]) == 0 {
		fmt.Println("Missing r parameters")
		return
	}
	// AUCH HEADER PRÜFEN!

	headers := []string{"X-Event-Key", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
	headerValues := make([]string, len(headers))
	for i := range headers {
		headerValues[i], err = getHeaderIfSet(r, headers[i])
		if err != nil {
			log.Printf("could not get header %v\n", headers[i])
		}
	}

	// all strings
	id := queryParams["id"][0]
	token := queryParams["token"][0]
	fmt.Printf("gitlab receive handler: id=%v, token=%v\n", id, token)
}

func giteaReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var payload GiteaPushPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("could not decode request body"))
		return
	}
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println("Could not parse query params")
		return
	}

	if len(queryParams["id"]) == 0 || len(queryParams["token"]) == 0 {
		fmt.Println("Missing r parameters")
		return
	}
	// AUCH HEADER PRÜFEN!

	headers := []string{"X-Event-Key", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
	headerValues := make([]string, len(headers))
	for i := range headers {
		headerValues[i], err = getHeaderIfSet(r, headers[i])
		if err != nil {
			log.Printf("could not get header %v\n", headers[i])
		}
	}

	// all strings
	id := queryParams["id"][0]
	token := queryParams["token"][0]
	fmt.Printf("gitea receive handler: id=%v, token=%v\n", id, token)
}

