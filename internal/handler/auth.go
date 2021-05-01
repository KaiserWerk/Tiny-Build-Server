package handler

import (
	"database/sql"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/fixtures"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"net/http"
	"strconv"
	"time"
)

// LoginHandler handles logins
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: consider enabled 2fa

	if r.Method == http.MethodPost {
		sessMgr := global.GetSessionManager()
		ds := databaseservice.New()
		//defer ds.Quit()

		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := ds.GetUserByEmail(email)
		if err != nil {
			helper.WriteToConsole("login: could not get user by Email in LoginHandler: " + err.Error())
			sessMgr.AddMessage("error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if security.DoesHashMatch(password, u.Password) {
			//helper.WriteToConsole("user " + u.Displayname + " authenticated successfully")

			if u.Locked {
				sessMgr.AddMessage("warning", "You account has been disabled.")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

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

	if err := templateservice.ExecuteTemplate(w, "login.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

// LogoutHandler handles logouts
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	sessMgr := global.GetSessionManager()
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

// RequestNewPasswordHandler handles password reset requests
func RequestNewPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: consider disabled pw reset
	sessMgr := global.GetSessionManager()

	if r.Method == http.MethodPost {
		ds := databaseservice.New()
		//defer ds.Quit()

		email := r.FormValue("login_email")
		if email != "" {
			u, err := ds.GetUserByEmail(email)
			if err != nil {
				helper.WriteToConsole("could not get user by Email in RequestNewPasswordHandler: " + err.Error())
				// fake success message
				sessMgr.AddMessage("success", "If this user/email exists, an email has been sent out with "+
					"instructions to set a new password")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			helper.WriteToConsole("user " + u.Displayname + " requested new password")

			registrationToken := security.GenerateToken(30)

			//_, err = db.Exec("INSERT INTO user_action (user_id, purpose, token, validity) VALUES (?, ?, ?, ?)",
			//	u.Id, "password_reset", registrationToken, time.Now().Add(1*time.Hour))
			t := time.Now().Add(1 * time.Hour)
			err = ds.InsertUserAction(u.Id, "password_reset", registrationToken, sql.NullTime{Valid: true, Time: t})
			if err != nil {
				helper.WriteToConsole("could not insert user pw reset action: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			settings, err := ds.GetAllSettings()
			if err != nil {
				helper.WriteToConsole("could not get all settings: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			data := struct {
				BaseUrl string
				Email   string
				Token   string
			}{
				BaseUrl: settings["base_url"],
				Email:   u.Email,
				Token:   registrationToken,
			}

			emailBody, err := templateservice.ParseEmailTemplate(string(fixtures.RequestNewPasswordEmail), data)
			if err != nil {
				helper.WriteToConsole("unable to parse email template: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = helper.SendEmail(
				settings,
				emailBody,
				fixtures.EmailSubjects[fixtures.RequestNewPasswordEmail],
				[]string{u.Email},
				nil,
			)
			if err != nil {
				helper.WriteToConsole("RequestNewPasswordHandler: could not send email: " + err.Error())
				//w.WriteHeader(http.StatusInternalServerError)
				//return
			}

			sessMgr.AddMessage("success", "If this user/email exists, an email has been sent out with "+
				"instructions to set a new password.")
			http.Redirect(w, r, "/password/reset", http.StatusSeeOther)
			return
		}
	}

	if err := templateservice.ExecuteTemplate(w, "password_request.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

// ResetPasswordHandler handles password resets
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: consider disabled pw reset
	sessMgr := global.GetSessionManager()
	email := r.URL.Query().Get("email")
	token := r.URL.Query().Get("token")

	if r.Method == http.MethodPost {
		ds := databaseservice.New()
		//defer ds.Quit()

		email := r.FormValue("email")
		token := r.FormValue("token")
		user, err := ds.GetUserByEmail(email)
		if err != nil {
			helper.WriteToConsole("could not fetch user by email: " + err.Error())
			sessMgr.AddMessage("error", "A user with the supplied email address does not exist.")
			http.Redirect(w, r, "/password/reset?token="+token, http.StatusSeeOther)
			return
		}

		// TODO: move to method in databaseservice
		//row := db.QueryRow("SELECT id, user_id, purpose, token, validity FROM user_action WHERE token = ?", token)
		//err = row.Scan(&action.Id, &action.UserId, &action.Purpose, &action.Token, &action.Validity)
		action, err := ds.GetUserActionByToken(token)
		if err != nil {
			helper.WriteToConsole("could not scan user action in ResetPasswordHandler: " + err.Error())
			sessMgr.AddMessage("error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}

		if !action.Validity.Valid || action.Validity.Time.Before(time.Now()) {
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

		hash, err := security.HashString(pw1)
		if err != nil {
			helper.WriteToConsole("could not hash password: " + err.Error())
			sessMgr.AddMessage("error", "An error occurred. Please try again.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		//_, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", hash, user.Id)
		user.Password = hash
		err = ds.UpdateUser(user)
		if err != nil {
			helper.WriteToConsole("could not update to new password: " + err.Error())
			sessMgr.AddMessage("error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		//_, err = db.Exec("UPDATE user_action SET validity = ? WHERE purpose = 'password_reset' AND user_id = ?", sql.NullTime{}, user.Id)
		err = ds.InvalidatePasswordResets(user.Id)
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

	if err := templateservice.ExecuteTemplate(w, "password_reset.html", data); err != nil {
		w.WriteHeader(404)
	}
}

// RegistrationHandler handles user account registrations
func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: consider disabled registration
	sessMgr := global.GetSessionManager()

	if r.Method == http.MethodPost {
		ds := databaseservice.New()
		//defer ds.Quit()

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
		_, err := ds.GetUserByEmail(email)
		if err == nil {
			helper.WriteToConsole("registration: user already exists")
			sessMgr.AddMessage("error", "This email address is already in use!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		hash, err := security.HashString(pw1)
		if err != nil {
			helper.WriteToConsole("registration: password could not be hashed: " + err.Error())
			sessMgr.AddMessage("error", "The new account could not be created; please try again!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		_, err = ds.FindUser("displayname = ?", displayName)
		if err == nil {
			helper.WriteToConsole("this displayname is already in use")
			sessMgr.AddMessage("error", "This display name is already in use, please select a different one.")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		//res, err := db.Exec("INSERT INTO user (displayname, email, password, locked) VALUES (?, ?, ?, 1)",
		//	displayName, email, hash)
		lastInsertId, err := ds.AddUser(entity.User{Displayname: displayName, Email: email, Password: hash, Locked: true})
		if err != nil {
			helper.WriteToConsole("registration: user could not be inserted: " + err.Error())
			sessMgr.AddMessage("error", "The new account could not be created; please try again!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		//if err != nil {
		//	helper.WriteToConsole("registration: could not obtain last insert id: " + err.Error())
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}

		token := security.GenerateToken(30)
		t := time.Now().Add(24 * time.Hour)
		err = ds.AddUserAction(entity.UserAction{UserId: lastInsertId, Purpose: "confirm_registration",
			Token: token, Validity: sql.NullTime{Valid: true, Time: t}})
		//_, err = db.Exec("INSERT INTO user_action (user_id, purpose, token, validity) VALUES (?, ?, ?, ?)",
		//	lastInsertId, "confirm_registration", token, time.Now().Add(24*time.Hour))
		if err != nil {
			helper.WriteToConsole("registration: could not insert user action: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		settings, err := ds.GetAllSettings()
		if err != nil {
			helper.WriteToConsole("registration: could not fetch settings: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := struct {
			BaseUrl string
			Token   string
		}{
			BaseUrl: settings["base_url"],
			Token:   token,
		}

		emailBody, err := templateservice.ParseEmailTemplate(string(fixtures.ConfirmRegistrationEmail), data)
		if err != nil {
			helper.WriteToConsole("unable to parse email template: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = helper.SendEmail(
			settings,
			emailBody,
			fixtures.EmailSubjects[fixtures.ConfirmRegistrationEmail],
			[]string{email},
			nil,
		)
		if err != nil {
			helper.WriteToConsole("registration: could not send email: " + err.Error())
			sessMgr.AddMessage("warning", "Your new account was created but the confirmation email could not be sent out!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		sessMgr.AddMessage("success", "Your new account was created and a confirmation email is on its way to you!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := templateservice.ExecuteTemplate(w, "register.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

// RegistrationConfirmHandler handles confirmations for newly registered user accounts
func RegistrationConfirmHandler(w http.ResponseWriter, r *http.Request) {

	var token string

	if r.Method == http.MethodPost {
		token = r.FormValue("token")
	} else {
		token = r.URL.Query().Get("token")
	}

	if token != "" {
		sessMgr := global.GetSessionManager()
		ds := databaseservice.New()
		//defer ds.Quit()

		ua, err := ds.GetUserActionByToken(token)
		//row := db.QueryRow("SELECT user_id, purpose, validity FROM user_action WHERE token = ?", token)
		//ua := entity.UserAction{Token: token}
		//err := row.Scan(&ua.UserId, &ua.Purpose, &ua.Validity)
		if err != nil {
			helper.WriteToConsole("confirm registration: could not scan: " + err.Error())
			sessMgr.AddMessage("error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		if ua.Purpose != "confirm_registration" {
			helper.WriteToConsole("confirm registration: wrong purpose")
			sessMgr.AddMessage("error", "This token is for a different purpose!")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		if !ua.Validity.Valid || ua.Validity.Time.Before(time.Now()) {
			helper.WriteToConsole("confirm registration: token validity run out")
			sessMgr.AddMessage("error", "This token is not valid anymore.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		user, err := ds.GetUserById(ua.UserId)
		if err != nil {
			helper.WriteToConsole("confirm registration: could not get user from db: " + err.Error())
			sessMgr.AddMessage("error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}
		user.Locked = false
		//_, err = db.Exec("UPDATE user SET locked = 0 WHERE id = ?", ua.UserId)
		err = ds.UpdateUser(user)
		if err != nil {
			helper.WriteToConsole("confirm registration: could not set locked flag in db: " + err.Error())
			sessMgr.AddMessage("error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}
		ua.Validity = sql.NullTime{}
		err = ds.UpdateUserAction(ua)
		//_, err = db.Exec("UPDATE user_action SET validity = ? WHERE token = ?", sql.NullTime{}, ua.Token)
		if err != nil {
			helper.WriteToConsole("confirm registration: could not null token validity in db: " + err.Error())
			sessMgr.AddMessage("error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		helper.WriteToConsole("confirm registration: successful")
		sessMgr.AddMessage("success", "Your account was successfully confirmed! You can now log in.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := templateservice.ExecuteTemplate(w, "confirm_registration.html", nil); err != nil {
		w.WriteHeader(404)
	}
}
