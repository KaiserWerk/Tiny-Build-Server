package main

import (
	"fmt"
	"net/http"
)

func payloadReceiveHandler(w http.ResponseWriter, r *http.Request) {
	bd, err := checkPayloadRequest(r)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, `{"status": "error", "message": "%s"}`, err.Error())
		return
	}

	go startBuildProcess(bd)

	_, _ = fmt.Fprint(w, `{"status": "success", "message": "build execution initiated"}`)
}

func bitBucketReceiveHandler(w http.ResponseWriter, r *http.Request) {
	//	var payload bitBucketPushPayload
	//	err := json.NewDecoder(r.Body).Decode(&payload)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		w.Write([]byte("could not decode request body"))
	//		return
	//	}
	//
	//	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	//	if err != nil {
	//		fmt.Println("Could not parse query params")
	//		return
	//	}
	//
	//	if len(queryParams["Id"]) == 0 || len(queryParams["token"]) == 0 {
	//		fmt.Println("Missing r parameters")
	//		return
	//	}
	//	// AUCH HEADER PRÜFEN!
	//
	//	headers := []string{"X-Event-Key", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
	//	headerValues := make([]string, len(headers))
	//	for i := range headers {
	//		headerValues[i], err = getHeaderIfSet(r, headers[i])
	//		if err != nil {
	//			log.Printf("could not get header %v\n", headers[i])
	//		}
	//	}
	//
	//	// all strings
	//	id := queryParams["Id"][0]
	//	token := queryParams["token"][0]
	//	branch := payload.Push.Changes[0].New.Name
	//	repoFullName := payload.Repository.FullName
	//
	//	fmt.Println("Id: " + id)
	//	fmt.Println("token: " + token)
	//
	//	fmt.Println("branch: " + branch)
	//	fmt.Println("repo full name: " + repoFullName)
	//
	//	buildDefinition, err := loadBuildDefinition(id)
	//	if err != nil {
	//		fmt.Println("error while loading build configuration:", err.Error())
	//		w.WriteHeader(http.StatusBadRequest)
	//		w.Write([]byte("received, but bad request"))
	//		return
	//	}
	//
	//	if buildDefinition.AuthToken != token {
	//		fmt.Println("no auth token match")
	//		w.WriteHeader(http.StatusUnauthorized)
	//		w.Write([]byte("received, but auth token mismatch"))
	//		return
	//	}
	//
	//	if buildDefinition.Repository.FullName != repoFullName || buildDefinition.Repository.Branch != branch {
	//		fmt.Println("repo name or branch mismatch")
	//		w.WriteHeader(http.StatusBadRequest)
	//		w.Write([]byte("received, but repository name/branch mismatch"))
	//	}
	//
	//	// now we can start the build process
	//	// something like
	//	fmt.Println("starting build process for Id " + id)
	//	go startBuildProcess(id, buildDefinition)
	//
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("received, build process initiated, everything fine"))
}

func gitHubReceiveHandler(w http.ResponseWriter, r *http.Request) {
	//var payload gitHubPushPayload
	//err := json.NewDecoder(r.Body).Decode(&payload)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	w.Write([]byte("could not decode request body"))
	//	return
	//}
	//queryParams, err := url.ParseQuery(r.URL.RawQuery)
	//if err != nil {
	//	fmt.Println("Could not parse query params")
	//	return
	//}
	//
	//if len(queryParams["Id"]) == 0 || len(queryParams["token"]) == 0 {
	//	fmt.Println("Missing request parameters")
	//	return
	//}
	//
	//headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
	//headerValues := make([]string, len(headers))
	//for i := range headers {
	//	headerValues[i], err = getHeaderIfSet(r, headers[i])
	//	if err != nil {
	//		log.Printf("could not get header %v\n", headers[i])
	//	}
	//}
	//
	//// all strings
	//id := queryParams["Id"][0]
	//token := queryParams["token"][0]
	//repoFullName := payload.Repository.FullName
	//branch := payload.Repository.DefaultBranch // other: MasterBranch
	//
	//buildConfig, err := loadBuildDefinition(id)
	//if err != nil {
	//	fmt.Println("error while loading build configuration:", err.Error())
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte("received, but bad request"))
	//	return
	//}
	//
	//if buildConfig.AuthToken != token {
	//	fmt.Println("no auth token match")
	//	w.WriteHeader(http.StatusUnauthorized)
	//	w.Write([]byte("received, but auth token mismatch"))
	//	return
	//}
	//
	//if buildConfig.Repository.FullName != repoFullName || buildConfig.Repository.Branch != branch {
	//	fmt.Println("repo name or branch mismatch")
	//	w.WriteHeader(http.StatusBadRequest)
	//	str := fmt.Sprintf("repo name expected: %v, got %v instead; branch name expected: %v, got %v instead",
	//		buildConfig.Repository.FullName, repoFullName, buildConfig.Repository.Branch, branch)
	//	w.Write([]byte("recevied, but repository name/branch mismatch: " + str))
	//	return
	//}
	//
	//// now we can start the build process
	//// something like
	//go startBuildProcess(id, buildConfig)
	//
	//w.WriteHeader(http.StatusOK)
	//w.Write([]byte("received, build process initiated, everything fine"))
}

func gitLabReceiveHandler(w http.ResponseWriter, r *http.Request) {
	//var payload gitLabPushPayload
	//err := json.NewDecoder(r.Body).Decode(&payload)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	w.Write([]byte("could not decode request body"))
	//	return
	//}
	//queryParams, err := url.ParseQuery(r.URL.RawQuery)
	//if err != nil {
	//	fmt.Println("Could not parse query params")
	//	return
	//}
	//
	//if len(queryParams["Id"]) == 0 || len(queryParams["token"]) == 0 {
	//	fmt.Println("Missing r parameters")
	//	return
	//}
	//// AUCH HEADER PRÜFEN!
	//
	//headers := []string{"X-GitLab-Event"}
	//headerValues := make([]string, len(headers))
	//for i := range headers {
	//	headerValues[i], err = getHeaderIfSet(r, headers[i])
	//	if err != nil {
	//		log.Printf("could not get header %v\n", headers[i])
	//	}
	//}
	//
	//// all strings
	//id := queryParams["Id"][0]
	//token := queryParams["token"][0]
	//repoFullName := payload.Project.PathWithNamespace
	//branch := payload.Project.DefaultBranch
	//
	//buildConfig, err := loadBuildDefinition(id)
	//if err != nil {
	//	fmt.Println("error while loading build configuration:", err.Error())
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte("received, but bad request"))
	//	return
	//}
	//
	//if buildConfig.AuthToken != token {
	//	fmt.Println("no auth token match")
	//	w.WriteHeader(http.StatusUnauthorized)
	//	w.Write([]byte("received, but auth token mismatch"))
	//	return
	//}
	//
	//if buildConfig.Repository.FullName != repoFullName || buildConfig.Repository.Branch != branch {
	//	fmt.Println("repo name or branch mismatch")
	//	w.WriteHeader(http.StatusBadRequest)
	//	str := fmt.Sprintf("repo name expected: %v, got %v instead; branch name expected: %v, got %v instead",
	//		buildConfig.Repository.FullName, repoFullName, buildConfig.Repository.Branch, branch)
	//	w.Write([]byte("recevied, but repository name/branch mismatch: " + str))
	//	return
	//}
	//
	//// now we can start the build process
	//// something like
	//go startBuildProcess(id, buildConfig)
	//
	//w.WriteHeader(http.StatusOK)
	//w.Write([]byte("received, build process initiated, everything fine"))
}

func giteaReceiveHandler(w http.ResponseWriter, r *http.Request) {
	//var payload giteaPushPayload
	//err := json.NewDecoder(r.Body).Decode(&payload)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	w.Write([]byte("could not decode request body"))
	//	return
	//}
	// defer r.Body.Close()
	//queryParams, err := url.ParseQuery(r.URL.RawQuery)
	//if err != nil {
	//	fmt.Println("Could not parse query params")
	//	return
	//}
	//
	//if len(queryParams["Id"]) == 0 || len(queryParams["token"]) == 0 {
	//	fmt.Println("Missing r parameters")
	//	return
	//}
	//// AUCH HEADER PRÜFEN!
	//
	//headers := []string{"X-Gitea-Delivery", "X-Gitea-Event"}
	//headerValues := make([]string, len(headers))
	//for i := range headers {
	//	headerValues[i], err = getHeaderIfSet(r, headers[i])
	//	if err != nil {
	//		log.Printf("could not get header %v\n", headers[i])
	//	}
	//}
	//
	//// all strings
	//id := queryParams["Id"][0]
	//token := queryParams["token"][0]
	//fmt.Printf("gitea receive handler still incomplete: Id=%v, token=%v\n", id, token)
}
