package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//writeToConsole("getting cookie value")
	//sessId, err := sessMgr.GetCookieValue(r)
	//if err != nil {
	//	writeToConsole("couldnt get cookie value")
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}
	//writeToConsole("getting session with id "+sessId)
	//session, err := sessMgr.GetSession(sessId)
	//if err != nil {
	//	writeToConsole("couldnt get session")
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}
	//if session == nil {
	//	writeToConsole("Session is NIL!")
	//}
	//writeToConsole("getting userID")
	//userIdStr, ok := session.GetVar("user_id")
	//if !ok {
	//	writeToConsole("couldnt get userID")
	//	http.Redirect(w, r, "/login", http.StatusSeeOther)
	//	return
	//}
	//
	//userId := userIdStr.(int)
	//
	// otherwise ok (logged in)
	writeToConsole("login check ok")
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
		if sessMgr == nil {
			writeToConsole("SessionManager is NIL!")
			return
		}

		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := getUserByEmail(email)
		if err != nil {
			writeToConsole("could not get user by email (maybe doesnt exist): " + err.Error())
			return
		}
		fmt.Printf("User: %v\n", u)
		if doesHashMatch(password, u.password) {
			writeToConsole("authenticated successfully")
			//continue settings cookie/starting session
			_, err := sessMgr.CreateSession(sessMgr.CookieName, time.Now().Add(30 * 24 * time.Hour))
			if err != nil {
				writeToConsole("could not create session: " + err.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			//if sess == nil {
			//	writeToConsole("Session is NIL!")
			//}
			//sess.SetVar("user_id", strconv.Itoa(u.id))
			//writeToConsole("session id: " + sess.Id)
			//sessMgr.SetCookie(w, sess.Id)
			////http.SetCookie(w, &http.Cookie{
			////	Name:       sessMgr.CookieName,
			////	Value:      sess.Id,
			////	Path:       "/",
			////	Expires:    time.Now().Add(30*24*time.Hour),
			////	HttpOnly:   true,
			////})
			//writeToConsole("cookie set")
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

func pingHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}
