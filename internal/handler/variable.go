package handler

import (
	"net/http"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

// VariableListHandler lists all variables available to the logged-in user
func (h *HttpHandler) VariableListHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger = h.ContextLogger("VariableListHandler")
	)

	variables, err := h.Ds.GetAvailableVariablesForUser(currentUser.Id)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get variables for user")
		http.Error(w, "could not get variables for user", http.StatusInternalServerError)
		return
	}

	data := struct {
		CurrentUser entity.User
		Variables   []entity.UserVariable
	}{
		CurrentUser: currentUser,
		Variables:   variables,
	}

	if err := templateservice.ExecuteTemplate(w, "variable_list.html", data); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}

// VariableAddHandler adds a new variable
func (h *HttpHandler) VariableAddHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger = h.ContextLogger("VariableAddHandler")
	)

	if r.Method == http.MethodPost {
		logger.Trace("TODO")
		// TODO: implement
	}

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "variable_edit.html", data); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}

// VariableEditHandler edits a variable
func (h *HttpHandler) VariableEditHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger = h.ContextLogger("VariableEditHandler")
	)

	if r.Method == http.MethodPost {
		logger.Trace("TODO")
		// TODO: implement
	}

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "variable_edit.html", data); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}

// VariableRemoveHandler removes a variable
func (h *HttpHandler) VariableRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var (
		currentUser = r.Context().Value("user").(entity.User)
		_ = h.ContextLogger("VariableRemoveHandler")
	)

	// TODO: implement

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "variable_remove.html", data); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}
