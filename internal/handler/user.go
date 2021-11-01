package handler

import (
	"net/http"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

// UserSettingsHandler handles changing a user's own settings
func (h *HttpHandler) UserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		currentUser = r.Context().Value("user").(entity.User)
		logger = h.ContextLogger("UserSettingsHandler")
	)

	if r.Method == http.MethodPost {
		sessMgr := global.GetSessionManager()
		password := r.FormValue("password")
		if password == "" {
			logger.Info("change user settings: password is empty")
			sessMgr.AddMessage("error", "Please enter your current password!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

		if !security.DoesHashMatch(password, currentUser.Password) {
			logger.Info("change user settings: entered password incorrect")
			sessMgr.AddMessage("error", "You entered an incorrect password!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}
		// current password ok, continue

		form := r.FormValue("form")
		if form == "change_data" {
			changes := 0
			displayname := r.FormValue("displayname")
			if displayname != "" && displayname != currentUser.Displayname {
				if h.Ds.RowExists("SELECT id FROM user WHERE displayname = ? AND id != ?", displayname, currentUser.Id) {
					logger.WithField("displayname", displayname).Info("displayname is already in use")
					sessMgr.AddMessage("error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
				changes++
				currentUser.Displayname = displayname
				err = h.Ds.UpdateUser(currentUser)
				if err != nil {
					logger.WithField("displayname", displayname).Info("displayname is already in use")
					sessMgr.AddMessage("error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			email := r.FormValue("email")
			if email != "" && email != currentUser.Email {
				if h.Ds.RowExists("SELECT id FROM user WHERE email = ? AND id != ?", email, currentUser.Id) {
					logger.WithField("email", email).Info("email is already in use")
					sessMgr.AddMessage("error", "This email is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
				changes++
				currentUser.Email = email
				err = h.Ds.UpdateUser(currentUser)
				if err != nil {
					logger.WithField("email", email).Info("email is already in use")
					sessMgr.AddMessage("error", "Could not update data!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			if changes > 0 {
				logger.Trace("change user settings: update successful")
				sessMgr.AddMessage("success", "Your changes have been saved.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			} else {
				logger.Trace("change user settings: no changes")
				sessMgr.AddMessage("info", "No changes were made.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}
		} else if form == "change_password" {
			newPassword1 := r.FormValue("newpassword1")
			newPassword2 := r.FormValue("newpassword2")

			if newPassword1 == "" || newPassword2 == "" {
				logger.Trace("no new password supplied")
				sessMgr.AddMessage("warning", "No new password supplied.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			if newPassword1 != newPassword2 {
				logger.Trace("new passwords do not match")
				sessMgr.AddMessage("error", "New passwords do not match!")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			hash, err := security.HashString(newPassword1)
			if err != nil {
				logger.Trace("could not hash new password")
				sessMgr.AddMessage("error", "An unknown error occurred.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			currentUser.Password = hash
			err = h.Ds.UpdateUser(currentUser)
			if err != nil {
				logger.Trace("could not set new password")
				sessMgr.AddMessage("error", "An unknown error occurred.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			sessMgr.AddMessage("success", "Your new password hast been set. Please use it for future logins.")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}
	}

	data := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err = templateservice.ExecuteTemplate(w, "user_settings.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
