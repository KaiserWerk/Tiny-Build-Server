package handler

import (
	"net/http"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

// UserSettingsHandler handles changing a user's own settings
func (h *HTTPHandler) UserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		err         error
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("UserSettingsHandler")
	)

	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		if password == "" {
			logger.Info("change user settings: password is empty")
			h.SessionService.AddMessage(w, "error", "Please enter your current password!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

		if !security.DoesHashMatch(password, currentUser.Password) {
			logger.Info("change user settings: entered password incorrect")
			h.SessionService.AddMessage(w, "error", "You entered an incorrect password!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}
		// current password ok, continue

		form := r.FormValue("form")
		if form == "change_data" {
			changes := 0
			displayname := r.FormValue("displayname")
			if displayname != "" && displayname != currentUser.DisplayName {
				if h.DBService.RowExists("SELECT id FROM user WHERE display_name = ? AND id != ?", displayname, currentUser.ID) {
					logger.WithField("displayname", displayname).Info("displayname is already in use")
					h.SessionService.AddMessage(w, "error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
				changes++
				currentUser.DisplayName = displayname
				err = h.DBService.UpdateUser(currentUser)
				if err != nil {
					logger.WithField("displayname", displayname).Info("displayname is already in use")
					h.SessionService.AddMessage(w, "error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			email := r.FormValue("email")
			if email != "" && email != currentUser.Email {
				if h.DBService.RowExists("SELECT id FROM user WHERE email = ? AND id != ?", email, currentUser.ID) {
					logger.WithField("email", email).Info("email is already in use")
					h.SessionService.AddMessage(w, "error", "This email is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
				changes++
				currentUser.Email = email
				err = h.DBService.UpdateUser(currentUser)
				if err != nil {
					logger.WithField("email", email).Info("email is already in use")
					h.SessionService.AddMessage(w, "error", "Could not update data!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			if changes > 0 {
				logger.Trace("change user settings: update successful")
				h.SessionService.AddMessage(w, "success", "Your changes have been saved.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			} else {
				logger.Trace("change user settings: no changes")
				h.SessionService.AddMessage(w, "info", "No changes were made.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}
		} else if form == "change_password" {
			newPassword1 := r.FormValue("newpassword1")
			newPassword2 := r.FormValue("newpassword2")

			if newPassword1 == "" || newPassword2 == "" {
				logger.Trace("no new password supplied")
				h.SessionService.AddMessage(w, "warning", "No new password supplied.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			if newPassword1 != newPassword2 {
				logger.Trace("new passwords do not match")
				h.SessionService.AddMessage(w, "error", "New passwords do not match!")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			hash, err := security.HashString(newPassword1)
			if err != nil {
				logger.Trace("could not hash new password")
				h.SessionService.AddMessage(w, "error", "An unknown error occurred.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			currentUser.Password = hash
			err = h.DBService.UpdateUser(currentUser)
			if err != nil {
				logger.Trace("could not set new password")
				h.SessionService.AddMessage(w, "error", "An unknown error occurred.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			h.SessionService.AddMessage(w, "success", "Your new password hast been set. Please use it for future logins.")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}
	}

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err = templateservice.ExecuteTemplate(h.Injector(), w, "user_settings.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
