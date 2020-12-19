package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"net/http"
)

func UserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	session, err := helper.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := helper.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		sessMgr := helper.GetSessionManager()
		db := helper.GetDbConnection()

		// check if password is correct
		password := r.FormValue("password")
		row := db.QueryRow("SELECT password FROM user WHERE id = ?", currentUser.Id)
		var queriedHash string
		err = row.Scan(&queriedHash)
		if err != nil {
			helper.WriteToConsole("change user settings: error querying user")
			sessMgr.AddMessage("error", "An unexpected error occurred!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

		hash, err := helper.HashString(password)
		if err != nil {
			helper.WriteToConsole("change user settings: could not hash password")
			sessMgr.AddMessage("error", "An unexpected error occurred!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

		if helper.DoesHashMatch(queriedHash, hash) {
			helper.WriteToConsole("change user settings: entered password incorrect")
			sessMgr.AddMessage("error", "You entered an incorrect password!")
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}
		// current password ok, continue

		form := r.FormValue("form")
		if form == "change_data" {
			displayname := r.FormValue("displayname")
			if displayname != "" && displayname != currentUser.Displayname {
				if exists := helper.RowExists("SELECT id FROM user WHERE displayname = ? AND id != ?", displayname, currentUser.Id); exists {
					helper.WriteToConsole("change user settings: displayname " + displayname + " is already in use")
					sessMgr.AddMessage("error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}

				_, err = db.Exec("UPDATE user SET displayname = ? where id = ?", displayname, currentUser.Id)
				if err != nil {
					helper.WriteToConsole("change user settings: displayname " + displayname + " is already in use")
					sessMgr.AddMessage("error", "This display name is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			email := r.FormValue("email")
			if email != "" && email != currentUser.Email {
				if exists := helper.RowExists("SELECT id FROM user WHERE email = ? AND id != ?", email, currentUser.Id); exists {
					helper.WriteToConsole("change user settings: email " + email + " is already in use")
					sessMgr.AddMessage("error", "This email is already in use!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}

				_, err = db.Exec("UPDATE user SET email = ? where id = ?", email, currentUser.Id)
				if err != nil {
					helper.WriteToConsole("change user settings: email " + displayname + " is already in use")
					sessMgr.AddMessage("error", "Could not update data!")
					http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
					return
				}
			}

			if (displayname != "" && displayname != currentUser.Displayname) ||
				(email != "" && email != currentUser.Email) {
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

			hash, err := helper.HashString(newPassword1)
			if err != nil {
				helper.WriteToConsole("change user settings: could not hash new password")
				sessMgr.AddMessage("error", "An unknown error occurred.")
				http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
				return
			}

			_, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", hash, currentUser.Id)
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

	if err = helper.ExecuteTemplate(w, "user_settings.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
