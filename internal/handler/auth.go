package handler

import (
	"database/sql"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"net/http"
	"strconv"
	"time"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: consider enabled 2fa
	sessMgr := helper.GetSessionManager()
	if r.Method == http.MethodPost {
		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := helper.GetUserByEmail(email)
		if err != nil {
			helper.WriteToConsole("login: could not get user by Email in LoginHandler: " + err.Error())
			sessMgr.AddMessage("error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if helper.DoesHashMatch(password, u.Password) {
			//helper.WriteToConsole("user " + u.Displayname + " authenticated successfully")
			//continue settings cookie/starting session
			sess, err := sessMgr.CreateSession(time.Now().Add(30 * 24 * time.Hour))
			if err != nil {
				helper.WriteToConsole("login: could not create session: " + err.Error())
				sessMgr.AddMessage("error", "Could not create session!")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			sess.SetVar("user_id", strconv.Itoa(u.Id))
			err = sessMgr.SetCookie(w, sess.Id)
			if err != nil {
				helper.WriteToConsole("login: could not set cookie: " + err.Error())
				sessMgr.AddMessage("error", "Session cookie could not be set!")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		} else {
			helper.WriteToConsole("login: not successful, password hash doesn't match")
			sessMgr.AddMessage("error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		helper.WriteToConsole("login: redirecting to dashboard")
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
	// TODO: consider disabled pw reset
	sessMgr := helper.GetSessionManager()
	if r.Method == http.MethodPost {
		email := r.FormValue("login_email")
		if email != "" {
			u, err := helper.GetUserByEmail(email)
			if err != nil {
				helper.WriteToConsole("could not get user by Email in RequestNewPasswordHandler: " + err.Error())
				// fake success message
				sessMgr.AddMessage("success", "If this user/email exists, an email has been sent out with "+
					"instructions to set a new password")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			helper.WriteToConsole("user " + u.Displayname + " requested new password")

			registrationToken := helper.GenerateToken(60)

			db := helper.GetDbConnection()
			_, err = db.Exec("INSERT INTO user_action (user_id, purpose, token, validity) VALUES (?, ?, ?, ?)",
				u.Id, "password_reset", registrationToken, time.Now().Add(1 * time.Hour))
			if err != nil {
				helper.WriteToConsole("could not insert user pw reset action: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			settings, err := helper.GetAllSettings()
			if err != nil {
				helper.WriteToConsole("could not get all settings: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			data := struct {
				BaseUrl string
				Email string
				Token string
			}{
				BaseUrl: settings["base_url"],
				Email: u.Email,
				Token: registrationToken,
			}
			err = helper.SendEmail(helper.PasswordReset, data, helper.EmailSubjects[helper.PasswordReset], []string{u.Email})
			if err != nil {
				helper.WriteToConsole("could not send email: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			sessMgr.AddMessage("success", "If this user/email exists, an email has been sent out with "+
				"instructions to set a new password.")
			http.Redirect(w, r, "/password/reset", http.StatusSeeOther)
			return
		}
	}

	if err := helper.ExecuteTemplate(w, "password_request.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: consider disabled pw reset
	sessMgr := helper.GetSessionManager()
	email := r.URL.Query().Get("email")
	token := r.URL.Query().Get("token")

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		token := r.FormValue("token")
		user, err := helper.GetUserByEmail(email)
		if err != nil {
			helper.WriteToConsole("could not fetch user by email: " + err.Error())
			sessMgr.AddMessage("error", "A user with the supplied email address does not exist.")
			http.Redirect(w, r, "/password/reset?token=" + token, http.StatusSeeOther)
			return
		}

		db := helper.GetDbConnection()

		var action entity.UserAction
		row := db.QueryRow("SELECT id, user_id, purpose, token, validity FROM user_action WHERE token = ?", token)
		err = row.Scan(&action.Id, &action.UserId, &action.Purpose, &action.Token, &action.Validity)
		if err != nil {
			helper.WriteToConsole("could not scan user action in ResetPasswordHandler: " + err.Error())
			sessMgr.AddMessage("error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}

		if action.Validity.Before(time.Now()) {
			helper.WriteToConsole("validity of token ran out")
			sessMgr.AddMessage("error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}

		if action.Purpose != "password_reset" {
			helper.WriteToConsole("token was for other purpose: " + action.Purpose)
			sessMgr.AddMessage("error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}

		pw1 := r.FormValue("password1")
		pw2 := r.FormValue("password2")
		if pw1 != pw2 {
			helper.WriteToConsole("passwords don't match")
			sessMgr.AddMessage("error", "Your entered passwords don't match.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		// TODO: check length/strength of password?

		hash, err := helper.HashString(pw1)
		if err != nil {
			helper.WriteToConsole("could not hash password: " + err.Error())
			sessMgr.AddMessage("error", "An error occurred. Please try again.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		_, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", hash, user.Id)
		if err != nil {
			helper.WriteToConsole("could not update to new password: " + err.Error())
			sessMgr.AddMessage("error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		_, err = db.Exec("UPDATE user_action SET validity = ? WHERE purpose = 'password_reset' AND user_id = ?", sql.NullTime{}, user.Id)
		if err != nil {
			helper.WriteToConsole("could not update user actions: " + err.Error())
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := struct {
		Email string
		Token string
	}{
		Email: email,
		Token: token,
	}

	if err := helper.ExecuteTemplate(w, "password_reset.html", data); err != nil {
		w.WriteHeader(404)
	}
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: consider disabled registration
	sessMgr := helper.GetSessionManager()

	if r.Method == http.MethodPost {
		displayName := r.FormValue("display_name")
		email := r.FormValue("email")
		pw1 := r.FormValue("password1")
		pw2 := r.FormValue("password2")

		if pw1 != pw2 {
			helper.WriteToConsole("registration: passwords don't match")
			sessMgr.AddMessage("error", "The entered passwords do not match!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		// TODO check password strength
		_, err := helper.GetUserByEmail(email)
		if err != nil {
			helper.WriteToConsole("registration: user already exists")
			sessMgr.AddMessage("error", "This email address is already in use!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		db := helper.GetDbConnection()
		_, err = db.Exec("INSERT INTO user (displayname, email, password, locked) VALUES (?, ?, ?, 1)",
			displayName, email, pw1)

	}

	if err := helper.ExecuteTemplate(w, "register.html", nil); err != nil {
		w.WriteHeader(404)
	}
}
