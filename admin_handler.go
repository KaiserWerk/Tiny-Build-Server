package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func adminUserListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection in adminUserListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, displayname, email, locked, admin FROM user")
	if err != nil {
		writeToConsole("could not query users in adminUserListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	var (
		userList []user
		u        user
	)
	for rows.Next() {
		err = rows.Scan(&u.Id, &u.Displayname, &u.Email, &u.Locked, &u.Admin)
		if err != nil {
			writeToConsole("could not scan user in adminUserListHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}

		userList = append(userList, u)
		u = user{}
	}

	contextData := struct {
		CurrentUser user
		AllUsers    []user
	}{
		CurrentUser: currentUser,
		AllUsers:    userList,
	}

	if err = executeTemplate(w, "admin_user_list.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

func adminUserAddHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

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
			db, err := getDbConnection()
			if err != nil {
				writeToConsole("could not get DB connection in adminUserAddHandler: " + err.Error())
				w.WriteHeader(500)
				return
			}
			defer db.Close()

			var u user
			row := db.QueryRow("SELECT id FROM user WHERE displayname = ?", displayname)
			err = row.Scan(&u.Id)
			if err != nil && err != sql.ErrNoRows {
				writeToConsole("display name already in use: " + err.Error())
				sessMgr.AddMessage("error", "This display name is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			row = db.QueryRow("SELECT id FROM user WHERE email = ?", email)
			err = row.Scan(&u.Id)
			if err != nil && err != sql.ErrNoRows {
				writeToConsole("email already in use: " + err.Error())
				sessMgr.AddMessage("error", "This email address is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			var passwordHash string
			if password != "" {
				passwordHash, err = hashString(password)
				if err != nil {
					writeToConsole("could not hash password: " + err.Error())
					w.WriteHeader(500)
					return
				}
			}

			_, err = db.Exec("INSERT INTO user (displayname, email, password, locked, admin) VALUES (?, ?, ?, ?, ?)",
				displayname, email, passwordHash, locked, admin)
			if err != nil {
				writeToConsole("could not insert row: " + err.Error())
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
		CurrentUser user
	}{
		CurrentUser: currentUser,
	}

	if err = executeTemplate(w, "admin_user_add.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

func adminUserEditHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection in adminUserEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

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

		var u user

		row := db.QueryRow("SELECT id FROM user WHERE displayname = ? AND id != ?", displayname, vars["id"])
		err = row.Scan(&u.Id)
		if err != nil && err != sql.ErrNoRows {
			writeToConsole("display name already in use: " + err.Error())
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		row = db.QueryRow("SELECT id FROM user WHERE email = ? AND id != ?", email, vars["id"])
		err = row.Scan(&u.Id)
		if err != nil && err != sql.ErrNoRows {
			writeToConsole("display name already in use: " + err.Error())
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		var pwQueryPart string
		if password != "" {
			pwQueryPart = "password = ? "
		}
		query := fmt.Sprintf("UPDATE user SET displayname = ?, email = ?, %slocked = ?, admin = ? WHERE id = ?", pwQueryPart)
		writeToConsole("edit user: query to execute: " + query)
		// update user
		if password != "" {
			hashedPassword, err := hashString(password)
			if err != nil {
				writeToConsole("could not hash password: " + err.Error())
				w.WriteHeader(500)
				return
			}
			_, err = db.Exec(query, displayname, email, hashedPassword, locked, admin)
		} else {
			_, err = db.Exec(query, displayname, email, locked, admin)
		}
		if err != nil {
			writeToConsole("could not update user: " + err.Error())
			w.WriteHeader(500)
			return
		}

		sessMgr.AddMessage("success", "Changes to user account saved!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	var editedUser user
	row := db.QueryRow("SELECT id, displayname, email, locked, admin FROM user WHERE id = ?", vars["id"])
	err = row.Scan(&editedUser.Id, &editedUser.Displayname, &editedUser.Email, &editedUser.Locked, &editedUser.Admin)
	if err != nil {
		writeToConsole("could not scan user in adminUserEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	contextData := struct {
		CurrentUser user
		UserToEdit  user
	}{
		CurrentUser: currentUser,
		UserToEdit:  editedUser,
	}

	if err = executeTemplate(w, "admin_user_edit.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

func adminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user by ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		var errors uint8 = 0
		form := r.FormValue("form")

		if form == "general_settings" {
			baseDatapath := r.FormValue("basedatapath")
			err = setSetting("basedatapath", baseDatapath)
			if err != nil {
				errors++
				writeToConsole("could not set baseDatapath")
			}
		} else if form == "security" {
			securityDisableRegistration := r.FormValue("security_disable_registration")
			if securityDisableRegistration != "1" {
				securityDisableRegistration = "0"
			}
			err = setSetting("security_disable_registration", securityDisableRegistration)
			if err != nil {
				errors++
				writeToConsole("could not set securityDisableRegistration")
			}

			securityDisablePasswordReset := r.FormValue("security_disable_password_reset")
			if securityDisablePasswordReset != "1" {
				securityDisablePasswordReset = "0"
			}
			err = setSetting("security_disable_password_reset", securityDisablePasswordReset)
			if err != nil {
				errors++
				writeToConsole("could not set securityDisableRegistration")
			}

			securityEmailConfirmationRequired := r.FormValue("security_email_confirmation_required")
			if securityEmailConfirmationRequired != "1" {
				securityEmailConfirmationRequired = "0"
			}
			err = setSetting("security_email_confirmation_required", securityEmailConfirmationRequired)
			if err != nil {
				errors++
				writeToConsole("could not set securityEmailConfirmationRequired")
			}

			security2fa := r.FormValue("security_2fa")
			if security2fa != "none" && security2fa != "email" && security2fa != "sms" {
				security2fa = "none"
			}
			err = setSetting("security_2fa", security2fa)
			if err != nil {
				errors++
				writeToConsole("could not set security2fa")
			}

		} else if form == "smtp" {
			smtpUsername := r.FormValue("smtp_username")
			err = setSetting("smtp_username", smtpUsername)
			if err != nil {
				errors++
				writeToConsole("could not set smtpUsername")
			}

			smtpPassword := r.FormValue("smtp_password")
			err = setSetting("smtp_password", smtpPassword)
			if err != nil {
				errors++
				writeToConsole("could not set smtpPassword")
			}

			smtpHost := r.FormValue("smtp_host")
			err = setSetting("smtp_host", smtpHost)
			if err != nil {
				errors++
				writeToConsole("could not set smtpHost")
			}

			smtpPort := r.FormValue("smtp_port")
			err = setSetting("smtp_port", smtpPort)
			if err != nil {
				errors++
				writeToConsole("could not set smtpPort")
			}

			smtpEncryption := r.FormValue("smtp_encryption")
			err = setSetting("smtp_encryption", smtpEncryption)
			if err != nil {
				errors++
				writeToConsole("could not set smtpEncryption")
			}
		} else if form == "executables" {
			goExec := r.FormValue("golang_executable")
			err = setSetting("golang_executable", goExec)
			if err != nil {
				errors++
				writeToConsole("could not set goExec")
			}

			dotnetExec := r.FormValue("dotnet_executable")
			err = setSetting("dotnet_executable", dotnetExec)
			if err != nil {
				errors++
				writeToConsole("could not set dotnetExec")
			}

			rustExec := r.FormValue("rust_executable")
			err = setSetting("rust_executable", rustExec)
			if err != nil {
				errors++
				writeToConsole("could not set rustExec")
			}
		}

		if errors > 0 {
			output := fmt.Sprintf("When trying to save admin settings, %d error(s) occured", errors)
			writeToConsole(output)
			sessMgr.AddMessage("error", output)
		} else {
			sessMgr.AddMessage("success", "Settings saved successfully!")
		}

		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
		return
	}
	allSettings, err := getAllSettings()
	if err != nil {
		writeToConsole("could not get allSettings: " + err.Error())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	contextData := struct {
		CurrentUser   user
		AdminSettings map[string]string
	}{
		currentUser,
		allSettings,
	}

	if err = executeTemplate(w, "admin_settings.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

