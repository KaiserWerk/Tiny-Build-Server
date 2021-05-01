package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"net/http"
)

// UserSettingsHandler handles changing a user's own settings
func UserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

		sessMgr := global.GetSessionManager()
		ds := databaseService.New()
		//defer ds.Quit()

		password := r.FormValue("password")
		if password == "" {
			helper.WriteToConsole("change user settings: password is empty")
			sessMgr.AddMessage("error", "Please enter your current password!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

		if !security.DoesHashMatch(password, currentUser.Password) {
			helper.WriteToConsole("change user settings: entered password incorrect")
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
				if ds.RowExists("SELECT id FROM user WHERE displayname = ? AND id != ?", displayname, currentUser.Id) {
					helper.WriteToConsole("change user settings: displayname " + displayname + " is already in use")
					sessMgr.AddMessage("error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
				changes++
				currentUser.Displayname = displayname
				err = ds.UpdateUser(currentUser)
				//_, err = db.Exec("UPDATE user SET displayname = ? where id = ?", displayname, currentUser.Id)
				if err != nil {
					helper.WriteToConsole("change user settings: displayname " + displayname + " is already in use")
					sessMgr.AddMessage("error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			email := r.FormValue("email")
			if email != "" && email != currentUser.Email {
				if ds.RowExists("SELECT id FROM user WHERE email = ? AND id != ?", email, currentUser.Id) {
					helper.WriteToConsole("change user settings: email " + email + " is already in use")
					sessMgr.AddMessage("error", "This email is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
				changes++
				currentUser.Email = email
				err = ds.UpdateUser(currentUser)
				//_, err = db.Exec("UPDATE user SET email = ? where id = ?", email, currentUser.Id)
				if err != nil {
					helper.WriteToConsole("change user settings: displayname " + displayname + " is already in use")
					sessMgr.AddMessage("error", "Could not update data!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			if changes > 0 {
				helper.WriteToConsole("change user settings: update successful")
				sessMgr.AddMessage("success", "Your changes have been saved.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			} else {
				helper.WriteToConsole("change user settings: no changes")
				sessMgr.AddMessage("info", "No changes were made.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}
		} else if form == "change_password" {
			newPassword1 := r.FormValue("newpassword1")
			newPassword2 := r.FormValue("newpassword2")

			if newPassword1 == "" || newPassword2 == "" {
				helper.WriteToConsole("change user settings: no new password supplied")
				sessMgr.AddMessage("warning", "No new password supplied.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			if newPassword1 != newPassword2 {
				helper.WriteToConsole("change password: new passwords do not match")
				sessMgr.AddMessage("error", "New passwords do not match!")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			hash, err := security.HashString(newPassword1)
			if err != nil {
				helper.WriteToConsole("change user settings: could not hash new password")
				sessMgr.AddMessage("error", "An unknown error occurred.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			currentUser.Password = hash
			err = ds.UpdateUser(currentUser)
			//_, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", hash, currentUser.Id)
			if err != nil {
				helper.WriteToConsole("change user settings: could not set new password")
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
