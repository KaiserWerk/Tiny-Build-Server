package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"net/http"
)

func VariableListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("VariableListHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ds := databaseService.New()
	variables, err := ds.GetAvailableVariablesForUser(currentUser.Id)

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

func VariableAddHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("VariableAddHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

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

func VariableShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("VariableShowHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "variable_list.html", data); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}

func VariableEditHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("VariableEditHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

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

func VariableRemoveHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("VariableRemoveHandler: could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err := templateservice.ExecuteTemplate(w, "variable_remove.html", data); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}
