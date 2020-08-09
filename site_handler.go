package main

import (
	"fmt"
	"io"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := templates["index.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		// hier gehts mit data weiter
		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		hash, err := hashString(password)
		if err != nil {
			writeToConsole("could not hash password")
		}
		u, err := getUserByEmail(email)
		if err != nil {
			writeToConsole("could not get user by email (maybe doesnt exist)")
			return
		}
		writeToConsole("user displayname: "+u.displayname+",  password: " + u.password)
		if doesHashMatch(password, u.password) {
			writeToConsole("login successful")
		} else {
			writeToConsole("login not successful")
		}

		writeToConsole("form submitted with email " + email + " and password " + password + " (hashed: "+hash+")")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t := templates["login.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}
