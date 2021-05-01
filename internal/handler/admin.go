package handler

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// AdminUserListHandler lists all existing user accounts
func AdminUserListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := security.CheckLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := sessionService.GetUserFromSession(session)
	if err != nil {
		helper.WriteToConsole("could not fetch user by ID: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ds := databaseService.New()

	userList, err := ds.GetAllUsers()
	if err != nil {
		helper.WriteToConsole("adminUserListHandler: could not query users: " + err.Error())
		w.WriteHeader(500)
		return
	}

	contextData := struct {
		CurrentUser entity.User
		AllUsers    []entity.User
	}{
		CurrentUser: currentUser,
		AllUsers:    userList,
	}

	if err = templateservice.ExecuteTemplate(w, "admin_user_list.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

// AdminUserAddHandler handles adding a new user account
func AdminUserAddHandler(w http.ResponseWriter, r *http.Request) {
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessMgr := global.GetSessionManager()

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

			ds := databaseService.New()

			_, err := ds.FindUser("displayname = ?", displayname)
			if err == nil {
				helper.WriteToConsole("AdminUserAddHandler: displayname already in use: " + err.Error())
				sessMgr.AddMessage("error", "This display name is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			_, err = ds.FindUser("email = ?", email)
			if err == nil {
				helper.WriteToConsole("AdminUserAddHandler: email already in use: " + err.Error())
				sessMgr.AddMessage("error", "This email address is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			var passwordHash string
			if password != "" {
				passwordHash, err = security.HashString(password)
				if err != nil {
					helper.WriteToConsole("AdminUserAddHandler: could not hash password: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				password = security.GenerateToken(10)
				passwordHash, err = security.HashString(password)
				if err != nil {
					helper.WriteToConsole("AdminUserAddHandler: could not hash generated password: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			userToAdd := entity.User{
				Displayname: displayname,
				Email:       email,
				Password:    passwordHash,
				Locked:      locked,
				Admin:       admin,
			}
			_, err = ds.AddUser(userToAdd)
			if err != nil {
				helper.WriteToConsole("AdminUserAddHandler: could not insert new user: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
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

	if err = templateservice.ExecuteTemplate(w, "admin_user_add.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

// AdminUserEditHandler handles edits to an existing user account
func AdminUserEditHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	sessMgr := global.GetSessionManager()
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)

	ds := databaseService.New()
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

		_, err = ds.FindUser("displayname = ? AND id != ?", displayname, vars["id"])
		if err == nil {
			helper.WriteToConsole("display name already in use")
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		_, err = ds.FindUser("email = ? AND id != ?", displayname, vars["id"])
		if err == nil {
			helper.WriteToConsole("display name already in use")
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		userId, err := strconv.Atoi(vars["id"])
		if err != nil {
			helper.WriteToConsole("invalid user id: " + err.Error())
			sessMgr.AddMessage("error", "You supplied an invalid user id!")
			http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
			return
		}

		updateUser := entity.User{
			Id:          userId,
			Displayname: displayname,
			Email:       email,
			Locked:      locked,
			Admin:       admin,
		}

		if password != "" {
			passwordHash, err := security.HashString(password)
			if err != nil {
				helper.WriteToConsole("AdminUserEditHandler: could not hash password: " + err.Error())
				w.WriteHeader(500)
				return
			}

			updateUser.Password = passwordHash
		}
		err = ds.UpdateUser(updateUser)
		//query := fmt.Sprintf("UPDATE user SET displayname = ?, email = ?, %slocked = ?, admin = ? WHERE id = ?", pwQueryPart)
		//helper.WriteToConsole("edit user: query to execute: " + query)
		// update user

		if err != nil {
			helper.WriteToConsole("could not update user: " + err.Error())
			w.WriteHeader(500)
			return
		}

		sessMgr.AddMessage("success", "Changes to user account saved!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("invalid user id: " + err.Error())
		sessMgr.AddMessage("error", "You supplied an invalid user id!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}
	editedUser, err := ds.GetUserById(userId)
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

	if err = templateservice.ExecuteTemplate(w, "admin_user_edit.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

// AdminUserRemoveHandler handles removals of user accounts
func AdminUserRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	sessMgr := global.GetSessionManager()
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		helper.WriteToConsole("invalid user id: " + err.Error())
		sessMgr.AddMessage("error", "You supplied an invalid user id!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	ds := databaseService.New()
	if r.Method == http.MethodPost {
		err = ds.DeleteUser(userId)
		if err != nil {
			helper.WriteToConsole("error removing user: " + err.Error())
			sessMgr.AddMessage("error", "An unknown error occurred, please try again.")
			http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	user, err := ds.GetUserById(userId)
	if err != nil {
		helper.WriteToConsole("could not scan user in adminUserEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	contextData := struct {
		CurrentUser  entity.User
		UserToRemove entity.User
	}{
		CurrentUser:  currentUser,
		UserToRemove: user,
	}

	if err = templateservice.ExecuteTemplate(w, "admin_user_remove.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}

// AdminSettingsHandler handles editing af administrative settings
func AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
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
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessMgr := global.GetSessionManager()
	ds := databaseService.New()
	//defer ds.Quit()

	if r.Method == http.MethodPost {

		var errors uint8 = 0
		form := r.FormValue("form")

		if form == "general_settings" {
			baseDatapath := r.FormValue("base_datapath")
			err = ds.SetSetting("base_datapath", baseDatapath)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set baseDatapath")
			}
			baseUrl := r.FormValue("base_url")
			err = ds.SetSetting("base_url", baseUrl)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set baseUrl")
			}
		} else if form == "security" {
			securityDisableRegistration := r.FormValue("security_disable_registration")
			if securityDisableRegistration != "1" {
				securityDisableRegistration = "0"
			}
			err = ds.SetSetting("security_disable_registration", securityDisableRegistration)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set securityDisableRegistration")
			}

			securityDisablePasswordReset := r.FormValue("security_disable_password_reset")
			if securityDisablePasswordReset != "1" {
				securityDisablePasswordReset = "0"
			}
			err = ds.SetSetting("security_disable_password_reset", securityDisablePasswordReset)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set securityDisableRegistration")
			}

			securityEmailConfirmationRequired := r.FormValue("security_email_confirmation_required")
			if securityEmailConfirmationRequired != "1" {
				securityEmailConfirmationRequired = "0"
			}
			err = ds.SetSetting("security_email_confirmation_required", securityEmailConfirmationRequired)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set securityEmailConfirmationRequired")
			}

			security2fa := r.FormValue("security_2fa")
			if security2fa != "none" && security2fa != "email" && security2fa != "sms" {
				security2fa = "none"
			}
			err = ds.SetSetting("security_2fa", security2fa)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set security2fa")
			}

		} else if form == "smtp" {
			smtpUsername := r.FormValue("smtp_username")
			err = ds.SetSetting("smtp_username", smtpUsername)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpUsername")
			}

			smtpPassword := r.FormValue("smtp_password")
			err = ds.SetSetting("smtp_password", smtpPassword)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpPassword")
			}

			smtpHost := r.FormValue("smtp_host")
			err = ds.SetSetting("smtp_host", smtpHost)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpHost")
			}

			smtpPort := r.FormValue("smtp_port")
			err = ds.SetSetting("smtp_port", smtpPort)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpPort")
			}

			smtpEncryption := r.FormValue("smtp_encryption")
			err = ds.SetSetting("smtp_encryption", smtpEncryption)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set smtpEncryption")
			}
		} else if form == "executables" {
			goExec := r.FormValue("golang_executable")
			err = ds.SetSetting("golang_executable", goExec)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set goExec")
			}

			dotnetExec := r.FormValue("dotnet_executable")
			err = ds.SetSetting("dotnet_executable", dotnetExec)
			if err != nil {
				errors++
				helper.WriteToConsole("could not set dotnetExec")
			}

			rustExec := r.FormValue("rust_executable")
			err = ds.SetSetting("rust_executable", rustExec)
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
	allSettings, err := ds.GetAllSettings()
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

	if err = templateservice.ExecuteTemplate(w, "admin_settings.html", contextData); err != nil {
		w.WriteHeader(404)
	}
}
