package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func bitBucketReceiveHandler(writer http.ResponseWriter, request *http.Request) {
	var payload BitBucketPushPayload
	err := json.NewDecoder(request.Body).Decode(&payload)
	if err != nil {
		fmt.Println(err.Error())
		writer.Write([]byte("could not decode request body"))
		return
	}

	queryParams, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		fmt.Println("Could not parse query params")
		return
	}

	if len(queryParams["id"]) == 0 || len(queryParams["token"]) == 0 {
		fmt.Println("Missing request parameters")
		return
	}
	// AUCH HEADER PRÜFEN!
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
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("received, but bad request"))
		return
	}

	fmt.Printf("build config - auth token: %v\n", buildConfig.AuthToken)

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("received, everything fine"))
}

func gitHubReceiveHandler(w http.ResponseWriter, r *http.Request) {
	// AUCH HEADER PRÜFEN!
}

func gitLabReceiveHandler(w http.ResponseWriter, r *http.Request) {
	// AUCH HEADER PRÜFEN!
}

