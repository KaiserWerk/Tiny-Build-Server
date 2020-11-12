package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"net/http"
	"strconv"
	"time"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	sessMgr := helper.GetSessionManager()
	if r.Method == http.MethodPost {
		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := helper.GetUserByEmail(email)
		if err != nil {
			helper.WriteToConsole("could not get user by Email in LoginHandler: " + err.Error())
			sessMgr.AddMessage("error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if helper.DoesHashMatch(password, u.Password) {
			//helper.WriteToConsole("user " + u.Displayname + " authenticated successfully")
			//continue settings cookie/starting session
			sess, err := sessMgr.CreateSession(time.Now().Add(30 * 24 * time.Hour))
			if err != nil {
				helper.WriteToConsole("could not create session: " + err.Error())
				sessMgr.AddMessage("error", "Could not create session!")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			sess.SetVar("user_id", strconv.Itoa(u.Id))
			err = sessMgr.SetCookie(w, sess.Id)
			if err != nil {
				helper.WriteToConsole("could not set cookie: " + err.Error())
				sessMgr.AddMessage("error", "Session cookie could not be set!")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		} else {
			helper.WriteToConsole("login not successful")
			sessMgr.AddMessage("error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		helper.WriteToConsole("redirecting to dashboard")
		sessMgr.AddMessage("success", "You are now logged in.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := helper.ExecuteTemplate(w, "login.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	sessMgr := helper.GetSessionManager()
	helper.WriteToConsole("getting cookie value")
	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		helper.WriteToConsole("could not get cookie value: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	//helper.WriteToConsole("getting session with Id "+sessId)
	session, err := sessMgr.GetSession(sessId)
	if err != nil {
		helper.WriteToConsole("could not get session: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = sessMgr.RemoveSession(session.Id)
	if err != nil {
		helper.WriteToConsole("could not remove session: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessMgr.AddMessage("success", "You are now logged out.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func RequestNewPasswordHandler(w http.ResponseWriter, r *http.Request) {
	sessMgr := helper.GetSessionManager()
	if r.Method == http.MethodPost {
		email := r.FormValue("login_email")
		if email != "" {
			u, err := helper.GetUserByEmail(email)
			if err != nil {
				helper.WriteToConsole("could not get user by Email in RequestNewPasswordHandler: " + err.Error())
				sessMgr.AddMessage("success", "If this user/email exists, an email has been sent out with "+
					"instructions to set a new password")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			helper.WriteToConsole("user: " + u.Displayname + " requested new password")
			// email an user versenden
			// zur reset seite weiterleiten
			sessMgr.AddMessage("success", "If this user/email exists, an email has been sent out with "+
				"instructions to set a new password")
			http.Redirect(w, r, "/password/reset", http.StatusSeeOther)
			return
		}

		return
	}

	if err := helper.ExecuteTemplate(w, "´password_request.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	//r.Method == "POST" {

	//}

	if err := helper.ExecuteTemplate(w, "password_request.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {

	if err := helper.ExecuteTemplate(w, "register.html", nil); err != nil {
		w.WriteHeader(404)
	}
}
