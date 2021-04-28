package handler

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"net/http"
)

func PayloadReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	// get token
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w,"could not determine token", http.StatusBadRequest)
		return
	}

	ds := databaseService.New()
	// find build definition by token
	bd, err := ds.FindBuildDefinition("token = ?", token)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not find build definition for token %s: %s", token, err.Error()), http.StatusNotFound)
		return
	}

	variables, err := ds.GetAvailableVariablesForUser(bd.CreatedBy)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not determine variables for user: %s", err.Error()), http.StatusNotFound)
		return
	}

	// unmarshal the build definition content
	bdContent, err := helper.UnmarshalBuildDefinitionContent(bd.Content, variables)
	if err != nil {
		http.Error(w, "could not unmarshal build definition content: " + err.Error(), http.StatusNotFound)
		return
	}

	// check if the correct headers, depending on the hoster, are set and
	// have the correct values
	err = buildservice.CheckPayloadHeader(bdContent, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// start the actual build process
	go buildservice.StartBuildProcess(bd, bdContent)
}
