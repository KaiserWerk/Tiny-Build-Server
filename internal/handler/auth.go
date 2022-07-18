package handler

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/mailer"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

// LoginHandler handles logins
func (h *HttpHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// TODO: consider enabled 2fa
	var (
		logger  = h.ContextLogger("LoginHandler")
		sessMgr = h.SessMgr
	)

	if r.Method == http.MethodPost {
		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := h.DBService.GetUserByEmail(email)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"email": email,
			}).Error("could not get user by email")
			sessMgr.AddMessage(w, "error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if security.DoesHashMatch(password, u.Password) {
			if u.Locked {
				sessMgr.AddMessage(w, "warning", "You account has been disabled.")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			//continue settings cookie/starting session
			sess, err := sessMgr.CreateSession(time.Now().Add(30 * 24 * time.Hour))
			if err != nil {
				logger.WithField("error", err.Error()).Error("could not create session")
				sessMgr.AddMessage(w, "error", "Could not create session!")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			sess.SetVar("user_id", fmt.Sprintf("%d", u.ID))
			sessMgr.SetCookie(w, sess.Id, time.Now().Add(30*24*time.Hour))

			r = r.WithContext(context.WithValue(r.Context(), "user", u))
		} else {
			logger.Info("login: not successful, password hash doesn't match")
			sessMgr.AddMessage(w, "error", "Invalid credentials!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		logger.Trace("login: redirecting to dashboard")
		sessMgr.AddMessage(w, "success", "You are now logged in.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := templateservice.ExecuteTemplate(h.Injector(), w, "login.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

// LogoutHandler handles logouts
func (h *HttpHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		logger  = h.ContextLogger("LogoutHandler")
		sessMgr = h.SessMgr
	)

	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get cookie value")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	session, err := sessMgr.GetSession(sessId)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = sessMgr.RemoveSession(session.Id)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not remove session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessMgr.AddMessage(w, "success", "You are now logged out.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RequestNewPasswordHandler handles password reset requests
func (h *HttpHandler) RequestNewPasswordHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		logger  = h.ContextLogger("RequestNewPasswordHandler")
		sessMgr = h.SessMgr
	)
	// TODO: consider disabled pw reset

	if r.Method == http.MethodPost {
		email := r.FormValue("login_email")
		if email != "" {
			u, err := h.DBService.GetUserByEmail(email)
			if err != nil {
				logger.WithField("error", err.Error()).Error("could not get user by email")
				// fake success message
				sessMgr.AddMessage(w, "success", "If this user/email exists, an email has been sent out with "+
					"instructions to set a new password")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			logger.WithField("displayName", u.DisplayName).Debug("user requested new password")

			registrationToken := security.GenerateToken(30)
			t := time.Now().Add(1 * time.Hour)
			err = h.DBService.InsertUserAction(u.ID, "password_reset", registrationToken, sql.NullTime{Valid: true, Time: t})
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error":  err.Error(),
					"userId": u.ID,
				}).Error("could not insert user pw reset action")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			settings, err := h.DBService.GetAllSettings()
			if err != nil {
				logger.WithField("error", err.Error()).Error("could not get all settings")
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

			emailBody, err := templateservice.ParseEmailTemplate(string(mailer.SubjRequestNewPassword), data)
			if err != nil {
				logger.WithField("error", err.Error()).Error("unable to parse email template")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = h.Mailer.SendEmail(
				emailBody,
				string(mailer.SubjRequestNewPassword),
				[]string{u.Email},
				nil,
			)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"email": u.Email,
				}).Warn("could not send email")
			}

			sessMgr.AddMessage(w, "success", "If this user/email exists, an email has been sent out with "+
				"instructions to set a new password.")
			http.Redirect(w, r, "/password/reset", http.StatusSeeOther)
			return
		}
	}

	if err := templateservice.ExecuteTemplate(h.Injector(), w, "password_request.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

// ResetPasswordHandler handles password resets
func (h *HttpHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		logger = h.ContextLogger("ResetPasswordHandler")
		email  = r.URL.Query().Get("email")
		token  = r.URL.Query().Get("token")
	)
	// TODO: consider disabled pw reset via setting

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		token := r.FormValue("token")
		user, err := h.DBService.GetUserByEmail(email)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"email": email,
			}).Error("could not fetch user by email")
			h.SessMgr.AddMessage(w, "error", "A user with the supplied email address does not exist.")
			http.Redirect(w, r, "/password/reset?token="+token, http.StatusSeeOther)
			return
		}

		action, err := h.DBService.GetUserActionByToken(token)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not scan user action")
			h.SessMgr.AddMessage(w, "error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}

		if !action.Validity.Valid || action.Validity.Time.Before(time.Now()) {
			logger.Debug("validity of token ran out")
			h.SessMgr.AddMessage(w, "error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}

		if action.Purpose != "password_reset" {
			logger.WithField("purpose", action.Purpose).Warn("token was for other purpose")
			h.SessMgr.AddMessage(w, "error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}

		pw1 := r.FormValue("password1")
		pw2 := r.FormValue("password2")
		if pw1 != pw2 {
			logger.Debug("passwords don't match")
			h.SessMgr.AddMessage(w, "error", "Your entered passwords don't match.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		// TODO: check length/strength of password?

		hash, err := security.HashString(pw1)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not hash password")
			h.SessMgr.AddMessage(w, "error", "An error occurred. Please try again.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		//_, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", hash, user.Id)
		user.Password = hash
		err = h.DBService.UpdateUser(user)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not update to new password")
			h.SessMgr.AddMessage(w, "error", "The supplied reset token is invalid.")
			http.Redirect(w, r, fmt.Sprintf("/password/reset?email=%s&token=%s", email, token), http.StatusSeeOther)
			return
		}
		//_, err = db.Exec("UPDATE user_action SET validity = ? WHERE purpose = 'password_reset' AND user_id = ?", sql.NullTime{}, user.Id)
		err = h.DBService.InvalidatePasswordResets(user.ID)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not update user actions")
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

	if err := templateservice.ExecuteTemplate(h.Injector(), w, "password_reset.html", data); err != nil {
		w.WriteHeader(404)
	}
}

// RegistrationHandler handles user account registrations
func (h *HttpHandler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		logger  = h.ContextLogger("RegistrationHandler")
		sessMgr = h.SessMgr
	)
	// TODO: consider disabled registration

	if r.Method == http.MethodPost {
		displayName := r.FormValue("display_name")
		email := r.FormValue("email")
		pw1 := r.FormValue("password1")
		pw2 := r.FormValue("password2")

		if pw1 != pw2 {
			logger.Trace("passwords don't match")
			sessMgr.AddMessage(w, "error", "The entered passwords do not match!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		// TODO check password strength
		_, err := h.DBService.GetUserByEmail(email)
		if err == nil {
			logger.Info("user already exists")
			sessMgr.AddMessage(w, "error", "This email address is already in use!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		hash, err := security.HashString(pw1)
		if err != nil {
			logger.WithField("error", err.Error()).Error("password could not be hashed")
			sessMgr.AddMessage(w, "error", "The new account could not be created; please try again!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		_, err = h.DBService.FindUser("display_name = ?", displayName)
		if err == nil {
			logger.WithFields(logrus.Fields{
				"error":       err.Error(),
				"displayname": displayName,
			}).Error("this displayname is already in use")
			sessMgr.AddMessage(w, "error", "This display name is already in use, please select a different one.")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		lastInsertId, err := h.DBService.AddUser(entity.User{DisplayName: displayName, Email: email, Password: hash, Locked: true})
		if err != nil {
			logger.WithField("error", err.Error()).Error("user could not be inserted")
			sessMgr.AddMessage(w, "error", "The new account could not be created; please try again!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		token := security.GenerateToken(30)
		t := time.Now().Add(24 * time.Hour)
		err = h.DBService.AddUserAction(entity.UserAction{UserId: lastInsertId, Purpose: "confirm_registration",
			Token: token, Validity: sql.NullTime{Valid: true, Time: t}})
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not insert user action")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		settings, err := h.DBService.GetAllSettings()
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not fetch settings")
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

		emailBody, err := templateservice.ParseEmailTemplate(string(mailer.SubjConfirmRegistration), data)
		if err != nil {
			logger.WithField("error", err.Error()).Error("unable to parse email template")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = h.Mailer.SendEmail(
			emailBody,
			string(mailer.SubjConfirmRegistration),
			[]string{email},
			nil,
		)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not send email")
			sessMgr.AddMessage(w, "warning", "Your new account was created but the confirmation email could not be sent out!")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		sessMgr.AddMessage(w, "success", "Your new account was created and a confirmation email is on its way to you!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := templateservice.ExecuteTemplate(h.Injector(), w, "register.html", nil); err != nil {
		w.WriteHeader(404)
	}
}

// RegistrationConfirmHandler handles confirmations for newly registered user accounts
func (h *HttpHandler) RegistrationConfirmHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		token  string
		logger = h.ContextLogger("RegistrationConfirmHandler")
	)

	if r.Method == http.MethodPost {
		token = r.FormValue("token")
	} else {
		token = r.URL.Query().Get("token")
	}

	if token != "" {
		ua, err := h.DBService.GetUserActionByToken(token)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not scan")
			h.SessMgr.AddMessage(w, "error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		if ua.Purpose != "confirm_registration" {
			logger.WithField("token", token).Warn("wrong token purpose")
			h.SessMgr.AddMessage(w, "error", "This token is for a different purpose!")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		if !ua.Validity.Valid || ua.Validity.Time.Before(time.Now()) {
			logger.WithFields(logrus.Fields{
				"token":    token,
				"validity": ua.Validity,
			}).Info("token validity ran out")
			h.SessMgr.AddMessage(w, "error", "This token is not valid anymore.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		user, err := h.DBService.GetUserById(ua.UserId)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not get user from DB")
			h.SessMgr.AddMessage(w, "error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		user.Locked = false
		err = h.DBService.UpdateUser(user)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not set locked flag in DB")
			h.SessMgr.AddMessage(w, "error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}
		ua.Validity = sql.NullTime{}
		err = h.DBService.UpdateUserAction(ua)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not null token validity in DB")
			h.SessMgr.AddMessage(w, "error", "An unknown error occurred.")
			http.Redirect(w, r, "/register/confirm", http.StatusSeeOther)
			return
		}

		logger.Trace("confirm registration: successful")
		h.SessMgr.AddMessage(w, "success", "Your account was successfully confirmed! You can now log in.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := templateservice.ExecuteTemplate(h.Injector(), w, "confirm_registration.html", nil); err != nil {
		w.WriteHeader(404)
	}
}
