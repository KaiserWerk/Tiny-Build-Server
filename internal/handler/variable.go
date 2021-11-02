package handler

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
		sessMgr = h.SessMgr
		ds = h.Ds
	)

	if r.Method == http.MethodPost {
		varName := r.FormValue("var_name")
		varVal := r.FormValue("var_value")
		var varPublic bool
		if r.FormValue("var_public") == "1" {
			varPublic = true
		}
		if varName == "" || varVal == "" {
			sessMgr.AddMessage("warning", "Please enter both a variable name and a value.")
			http.Redirect(w, r, "/variable/add", http.StatusSeeOther)
			return
		}

		_, err := ds.FindVariable("user_entry_id = ? AND variable = ?", currentUser.Id, varName)
		if err == nil {
			sessMgr.AddMessage("error", "This variable already exists!")
			http.Redirect(w, r, "/variable/add", http.StatusSeeOther)
			return
		}

		uv := entity.UserVariable{
			UserEntryId: currentUser.Id,
			Variable:    varName,
			Value:       varVal,
			Public:      varPublic,
		}


		if _, err = ds.AddVariable(uv); err != nil {
			logger.WithField("error", err.Error()).Error("could not insert new user variable")
			sessMgr.AddMessage("error", "The variable could not be added!")
			http.Redirect(w, r, "/variable/add", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/variable/list", http.StatusSeeOther)
		return
	}

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "variable_add.html", data); err != nil {
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
		logger = h.ContextLogger("VariableRemoveHandler")
		vars = mux.Vars(r)
	)

	v, err := h.Ds.FindVariable("user_entry_id = ? AND id = ?", currentUser.Id, vars["id"])
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"variableId": vars["id"],
			"userId": currentUser.Id,
		}).Error("could not find variable")
		h.SessMgr.AddMessage("error", "The variable could not be found or it is not yours!")
		http.Redirect(w, r, "/variable/list", http.StatusSeeOther)
		return
	}

	if err = h.Ds.DeleteVariable(v.Id); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"variableId": vars["id"],
			"userId": currentUser.Id,
		}).Error("could not delete variable from DB")
		h.SessMgr.AddMessage("error", "The variable could not be removed!")
		// no redirect here
	}

	http.Redirect(w, r, "/variable/list", http.StatusSeeOther)
}
