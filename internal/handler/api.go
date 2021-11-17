package handler

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
)

// PayloadReceiveHandler takes care of accepting the payload from the webhook HTTP call
// sent by a Git hoster
func (h *HttpHandler) PayloadReceiveHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		err    error
		logger = h.ContextLogger("PayloadReceiveHandler")
	)

	// get token
	token := r.URL.Query().Get("token")
	if token == "" {
		logger.Error("missing token")
		http.Error(w, "could not determine token", http.StatusBadRequest)
		return
	}

	// find build definition by token
	bd, err := h.Ds.FindBuildDefinition("token = ?", token)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"token": token,
		}).Error("could not find build definition for token")
		http.Error(w, fmt.Sprintf("could not find build definition for token %s: %s", token, err.Error()), http.StatusNotFound)
		return
	}

	variables, err := h.Ds.GetAvailableVariablesForUser(bd.CreatedBy)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not determine variables for user")
		http.Error(w, fmt.Sprintf("could not determine variables for user: %s", err.Error()), http.StatusNotFound)
		return
	}

	helper.ReplaceVariables(&bd.Content, variables)

	// unmarshal the build definition content
	var bdContent entity.BuildDefinitionContent
	if err = helper.UnmarshalBuildDefinitionContent(bd.Content, &bdContent); err != nil {
		logger.WithField("error", err.Error()).Error("could not unmarshal build definition")
		http.Error(w, "could not unmarshal build definition content: "+err.Error(), http.StatusNotFound)
		return
	}

	// check if the correct headers, depending on the hoster, are set and
	// have the correct values
	err = buildservice.CheckPayloadHeader(bdContent, r)
	if err != nil {
		logger.WithField("error", err.Error()).Error("request headers are incorrect")
		http.Error(w, "request headers are incorrect", http.StatusBadRequest)
		return
	}

	// start the actual build process
	go buildservice.StartBuildProcess(bd, 0)
}
