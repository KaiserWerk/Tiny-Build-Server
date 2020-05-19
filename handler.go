package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
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

	eventKey, err := getHeaderIfSet(r, "X-Event-Key")
	if err != nil {
		log.Println("could not get header X-Event-Key")
	}
	hookUuid, err := getHeaderIfSet(r, "X-Hook-UUID")
	if err != nil {
		log.Println("could not get header X-Hook-UUID")
	}
	requestUuid, err := getHeaderIfSet(r, "X-Request-UUID")
	if err != nil {
		log.Println("could not get header X-Request-UUID")
	}
	attempt, err := getHeaderIfSet(r, "X-Attempt-Number")
	if err != nil {
		log.Println("could not get header X-Attempt-Number")
	}

	fmt.Println("fetched bitbucket headers:", eventKey, hookUuid, requestUuid, attempt)

	// all strings
	id := queryParams["id"][0]
	token := queryParams["token"][0]
	branch := payload.Push.Changes[0].New.Name
	repoFullName := payload.Repository.FullName

	fmt.Println("id: " + id)
	fmt.Println("token: " + token)

	fmt.Println("branch: " + branch)
	fmt.Println("repo full name: " + repoFullName)

	buildConfig, err := loadBuildDefinition(id)
	if err != nil {
		fmt.Println("error while loading build configuration:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("received, but bad request"))
		return
	}

	// auth token check
	fmt.Printf("build config - auth token: %v\n", buildConfig.AuthToken)


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
	// AUCH HEADER PRÜFEN!
}

func gitLabReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var payload GitLabPushPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("could not decode request body"))
		return
	}
	// AUCH HEADER PRÜFEN!
}

func giteaReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var payload GiteaPushPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("could not decode request body"))
		return
	}
	// AUCH HEADER PRÜFEN!
}

