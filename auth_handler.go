package main

import (
	"net/http"
	"strconv"
	"time"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := getUserByEmail(email)
		if err != nil {
			writeToConsole("could not get user by Email (maybe doesnt exist): " + err.Error())
			sessMgr.AddMessage("error", "Invalid credentials!")
			// flashbag
			return
		}

		if doesHashMatch(password, u.Password) {
			writeToConsole("user " + u.Displayname + " authenticated successfully")
			//continue settings cookie/starting session
			sess, err := sessMgr.CreateSession(time.Now().Add(30 * 24 * time.Hour))
			if err != nil {
				writeToConsole("could not create session: " + err.Error())
				sessMgr.AddMessage("error", "Could not create session!")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			sess.SetVar("user_id", strconv.Itoa(u.Id))
			err = sessMgr.SetCookie(w, sess.Id)
			if err != nil {
				writeToConsole("could not set cookie: " + err.Error())
				sessMgr.AddMessage("error", "Session cookie could not be set!")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		} else {
			writeToConsole("login not successful")
			sessMgr.AddMessage("error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		sessMgr.AddMessage("success", "You are now logged in.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := executeTemplate(w, "login.html", nil); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["login.html"]
	//if t != nil {
	//	err := t.Execute(w, nil)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	writeToConsole("getting cookie value")
	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		writeToConsole("could not get cookie value: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	//writeToConsole("getting session with Id "+sessId)
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

	sessMgr.AddMessage("success", "You are now logged out.")
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

	if err := executeTemplate(w, "´password_request.html", nil); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["login.html"]
	//if t != nil {
	//	err := t.Execute(w, nil)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	if err := executeTemplate(w, "password_request.html", nil); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["password_request.html"]
	//if t != nil {
	//	err := t.Execute(w, nil)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
}

func registrationHandler(w http.ResponseWriter, r *http.Request) {

	if err := executeTemplate(w, "register.html", nil); err != nil {
		w.WriteHeader(404)
	}

	//t := templates["register.html"]
	//if t != nil {
	//	err := t.Execute(w, nil)
	//	if err != nil {
	//		fmt.Println("error:", err.Error())
	//	}
	//} else {
	//	w.WriteHeader(http.StatusNotFound)
	//}
}
