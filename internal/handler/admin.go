package handler

import (
	"database/sql"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/gorilla/mux"
	"net/http"
)

func AdminUserListHandler(w http.ResponseWriter, r *http.Request) {
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	db := helper.GetDbConnection()

	rows, err := db.Query("SELECT id, displayname, email, locked, admin FROM user")
	if err != nil {
		helper.WriteToConsole("could not query users in adminUserListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	var (
		userList []entity.User
		u        entity.User
	)
	for rows.Next() {
		err = rows.Scan(&u.Id, &u.Displayname, &u.Email, &u.Locked, &u.Admin)
		if err != nil {
			helper.WriteToConsole("could not scan user in adminUserListHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}

		userList = append(userList, u)
		u = entity.User{}
	}

	contextData := struct {
		CurrentUser entity.User
		AllUsers    []entity.User
	}{
		CurrentUser: currentUser,
		AllUsers:    userList,
	}

	if err = helper.ExecuteTemplate(w, "admin_user_list.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

func AdminUserAddHandler(w http.ResponseWriter, r *http.Request) {
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessMgr := helper.GetSessionManager()

	if r.Method == "POST" {
		// displayname, email, password, locked, admin
		displayname := r.FormValue("displayname")
		email := r.FormValue("email")
		password := r.FormValue("password")
		var locked, admin bool
		if r.FormValue("locked") == "1" {
			locked = true
		}
		if r.FormValue("admin") == "1" {
			admin = true
		}

		if displayname != "" && email != "" {
			db := helper.GetDbConnection()
			var uid int
			row := db.QueryRow("SELECT id FROM user WHERE displayname = ?", displayname)
			err = row.Scan(&uid)
			if err != nil && err != sql.ErrNoRows {
				helper.WriteToConsole("display name already in use: " + err.Error())
				sessMgr.AddMessage("error", "This display name is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			row = db.QueryRow("SELECT id FROM user WHERE email = ?", email)
			err = row.Scan(&uid)
			if err != nil && err != sql.ErrNoRows {
				helper.WriteToConsole("email already in use: " + err.Error())
				sessMgr.AddMessage("error", "This email address is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			var passwordHash string
			if password != "" {
				passwordHash, err = helper.HashString(password)
				if err != nil {
					helper.WriteToConsole("could not hash password: " + err.Error())
					w.WriteHeader(500)
					return
				}
			}

			_, err = db.Exec("INSERT INTO user (displayname, email, password, locked, admin) VALUES (?, ?, ?, ?, ?)",
				displayname, email, passwordHash, locked, admin)
			if err != nil {
				helper.WriteToConsole("could not insert row: " + err.Error())
				w.WriteHeader(500)
				return
			}

			sessMgr.AddMessage("success", "User account was created successfully!")
			http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
			return
		}
		sessMgr.AddMessage("warning", "You need to supply a display name and an email address.")
	}

	contextData := struct {
		CurrentUser entity.User
	}{
		CurrentUser: currentUser,
	}

	if err = helper.ExecuteTemplate(w, "admin_user_add.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

func AdminUserEditHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	sessMgr := helper.GetSessionManager()
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)

	db := helper.GetDbConnection()
	if r.Method == http.MethodPost {

		displayname := r.FormValue("displayname")
		email := r.FormValue("email")
		password := r.FormValue("password")
		var locked, admin bool
		if r.FormValue("locked") == "1" {
			locked = true
		}
		if r.FormValue("admin") == "1" {
			admin = true
		}

		var u entity.User

		row := db.QueryRow("SELECT id FROM user WHERE displayname = ? AND id != ?", displayname, vars["id"])
		err = row.Scan(&u.Id)
		if err != nil && err != sql.ErrNoRows {
			helper.WriteToConsole("display name already in use: " + err.Error())
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		row = db.QueryRow("SELECT id FROM user WHERE email = ? AND id != ?", email, vars["id"])
		err = row.Scan(&u.Id)
		if err != nil && err != sql.ErrNoRows {
			helper.WriteToConsole("display name already in use: " + err.Error())
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		var pwQueryPart string
		if password != "" {
			pwQueryPart = "password = ? "
		}
		query := fmt.Sprintf("UPDATE user SET displayname = ?, email = ?, %slocked = ?, admin = ? WHERE id = ?", pwQueryPart)
		helper.WriteToConsole("edit user: query to execute: " + query)
		// update user
		if password != "" {
			hashedPassword, err := helper.HashString(password)
			if err != nil {
				helper.WriteToConsole("could not hash password: " + err.Error())
				w.WriteHeader(500)
				return
			}
			_, err = db.Exec(query, displayname, email, hashedPassword, locked, admin)
		} else {
			_, err = db.Exec(query, displayname, email, locked, admin)
		}
		if err != nil {
			helper.WriteToConsole("could not update user: " + err.Error())
			w.WriteHeader(500)
			return
		}

		sessMgr.AddMessage("success", "Changes to user account saved!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	var editedUser entity.User
	row := db.QueryRow("SELECT id, displayname, email, locked, admin FROM user WHERE id = ?", vars["id"])
	err = row.Scan(&editedUser.Id, &editedUser.Displayname, &editedUser.Email, &editedUser.Locked, &editedUser.Admin)
	if err != nil {
		helper.WriteToConsole("could not scan user in adminUserEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	contextData := struct {
		CurrentUser entity.User
		UserToEdit  entity.User
	}{
		CurrentUser: currentUser,
		UserToEdit:  editedUser,
	}

	if err = helper.ExecuteTemplate(w, "admin_user_edit.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

func AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessMgr := helper.GetSessionManager()

	if r.Method == http.MethodPost {
		var errors uint8 = 0
		form := r.FormValue("form")

		if form == "general_settings" {
			baseDatapath := r.FormValue("basedatapath")
			err = helper.SetSetting("basedatapath", baseDatapath)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set baseDatapath")
			}
		} else if form == "security" {
			securityDisableRegistration := r.FormValue("security_disable_registration")
			if securityDisableRegistration != "1" {
				securityDisableRegistration = "0"
			}
			err = helper.SetSetting("security_disable_registration", securityDisableRegistration)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set securityDisableRegistration")
			}

			securityDisablePasswordReset := r.FormValue("security_disable_password_reset")
			if securityDisablePasswordReset != "1" {
				securityDisablePasswordReset = "0"
			}
			err = helper.SetSetting("security_disable_password_reset", securityDisablePasswordReset)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set securityDisableRegistration")
			}

			securityEmailConfirmationRequired := r.FormValue("security_email_confirmation_required")
			if securityEmailConfirmationRequired != "1" {
				securityEmailConfirmationRequired = "0"
			}
			err = helper.SetSetting("security_email_confirmation_required", securityEmailConfirmationRequired)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set securityEmailConfirmationRequired")
			}

			security2fa := r.FormValue("security_2fa")
			if security2fa != "none" && security2fa != "email" && security2fa != "sms" {
				security2fa = "none"
			}
			err = helper.SetSetting("security_2fa", security2fa)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set security2fa")
			}

		} else if form == "smtp" {
			smtpUsername := r.FormValue("smtp_username")
			err = helper.SetSetting("smtp_username", smtpUsername)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpUsername")
			}

			smtpPassword := r.FormValue("smtp_password")
			err = helper.SetSetting("smtp_password", smtpPassword)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpPassword")
			}

			smtpHost := r.FormValue("smtp_host")
			err = helper.SetSetting("smtp_host", smtpHost)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpHost")
			}

			smtpPort := r.FormValue("smtp_port")
			err = helper.SetSetting("smtp_port", smtpPort)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpPort")
			}

			smtpEncryption := r.FormValue("smtp_encryption")
			err = helper.SetSetting("smtp_encryption", smtpEncryption)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpEncryption")
			}
		} else if form == "executables" {
			goExec := r.FormValue("golang_executable")
			err = helper.SetSetting("golang_executable", goExec)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set goExec")
			}

			dotnetExec := r.FormValue("dotnet_executable")
			err = helper.SetSetting("dotnet_executable", dotnetExec)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set dotnetExec")
			}

			rustExec := r.FormValue("rust_executable")
			err = helper.SetSetting("rust_executable", rustExec)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set rustExec")
			}
		}

		if errors > 0 {
			output := fmt.Sprintf("When trying to save admin settings, %d error(s) occured", errors)
			helper.WriteToConsole(output)
			sessMgr.AddMessage("error", output)
		} else {
			sessMgr.AddMessage("success", "Settings saved successfully!")
		}

		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
		return
	}
	allSettings, err := helper.GetAllSettings()
	if err != nil {
		helper.WriteToConsole("could not get allSettings: " + err.Error())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	contextData := struct {
		CurrentUser   entity.User
		AdminSettings map[string]string
	}{
		CurrentUser:   currentUser,
		AdminSettings: allSettings,
	}

	if err = helper.ExecuteTemplate(w, "admin_settings.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}
