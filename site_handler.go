package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)


func indexHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	latestBuilds, err := getNewestBuildExecutions(5)
	if err != nil {
		writeToConsole("could not fet latest build executions: " + err.Error())
	}
	latestBuildDefs, err := getNewestBuildDefinitions(5)
	if err != nil {
		writeToConsole("could not fet latest build executions: " + err.Error())
	}

	indexData := struct {
		CurrentUser     user
		LatestBuilds    []buildExecution
		LatestBuildDefs []buildDefinition
	}{
		CurrentUser:     currentUser,
		LatestBuilds:    latestBuilds,
		LatestBuildDefs: latestBuildDefs,
	}

	//otherwise ok (logged in)
	//writeToConsole("login check ok")
	t := templates["index.html"]
	if t != nil {
		err := t.Execute(w, indexData)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := getUserByEmail(email)
		if err != nil {
			writeToConsole("could not get user by Email (maybe doesnt exist): " + err.Error())
			return
		}
		fmt.Printf("User: %v\n", u)
		if doesHashMatch(password, u.Password) {
			writeToConsole("authenticated successfully")
			//continue settings cookie/starting session
			sess, err := sessMgr.CreateSession(time.Now().Add(30 * 24 * time.Hour))
			if err != nil {
				writeToConsole("could not create session: " + err.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			sess.SetVar("user_id", strconv.Itoa(u.Id))
			writeToConsole("session Id: " + sess.Id)
			err = sessMgr.SetCookie(w, sess.Id)
			if err != nil {
				writeToConsole("could not set cookie: " + err.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			writeToConsole("cookie set")
		} else {
			writeToConsole("login not successful")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		writeToConsole("redirect to index page")
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	writeToConsole("getting cookie value")
	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		writeToConsole("could not get cookie value: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	writeToConsole("getting session with Id "+sessId)
	session, err := sessMgr.GetSession(sessId)
	if err != nil {
		writeToConsole("could not get session: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = sessMgr.RemoveSession(session.Id)
	if err != nil {
		writeToConsole("could not remove session: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func requestNewPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		email := r.FormValue("login_email")
		if email != "" {
			u, err := getUserByEmail(email)
			if err != nil {
				writeToConsole("could not get user by Email (maybe doesnt exist): " + err.Error())
				return
			}

			writeToConsole("user: " + u.Displayname)
			// email an user versenden
			// zur reset seite weiterleiten
		}
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

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {

}
func registrationHandler(w http.ResponseWriter, r *http.Request) {

}

func adminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

		//email := r.FormValue("login_email")

	}
	allSettings, err := getAllSettings()
	if err != nil {
		writeToConsole("could not get allSettings: " + err.Error())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	contextData := struct {
		CurrentUser user
		AdminSettings map[string]string
	}{
		currentUser,
		allSettings,
	}

	t := templates["admin_settings.html"]
	if t != nil {
		err := t.Execute(w, contextData)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}