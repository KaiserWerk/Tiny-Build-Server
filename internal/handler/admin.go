package handler

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/security"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"

	"github.com/gorilla/mux"
)

// AdminUserListHandler lists all existing user accounts
func (h *HttpHandler) AdminUserListHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		logger = h.ContextLogger("AdminUserListHandler")
		currentUser = r.Context().Value("user").(entity.User)
	)

	userList, err := h.Ds.GetAllUsers()
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not query users")
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
func (h *HttpHandler) AdminUserAddHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		err error
		logger = h.ContextLogger("AdminUserAddHandler")
		currentUser = r.Context().Value("user").(entity.User)
		sessMgr = h.SessMgr
	)

	if r.Method == "POST" {
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
			ds := h.Ds

			_, err := ds.FindUser("displayname = ?", displayname)
			if err == nil {
				logger.WithField("error", err.Error()).Error("displayname already in use")
				sessMgr.AddMessage("error", "This display name is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			_, err = ds.FindUser("email = ?", email)
			if err == nil {
				logger.WithField("error", err.Error()).Error("email already in use")
				sessMgr.AddMessage("error", "This email address is already in use!")
				http.Redirect(w, r, "/admin/user/add", http.StatusSeeOther)
				return
			}

			var passwordHash string
			if password != "" {
				passwordHash, err = security.HashString(password)
				if err != nil {
					logger.WithField("error", err.Error()).Error("could not hash password")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				password = security.GenerateToken(10)
				passwordHash, err = security.HashString(password)
				if err != nil {
					logger.WithField("error", err.Error()).Error("could not hash generated password")
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
				logger.WithField("error", err.Error()).Error("could not insert new user")
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
func (h *HttpHandler) AdminUserEditHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		err error
		sessMgr = h.SessMgr
		logger = h.ContextLogger("AdminUserEditHandler")
		currentUser = r.Context().Value("user").(entity.User)
		vars = mux.Vars(r)
	)

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

		_, err = h.Ds.FindUser("displayname = ? AND id != ?", displayname, vars["id"])
		if err == nil {
			logger.WithField("error", err.Error()).Error("display name already in use")
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		_, err = h.Ds.FindUser("email = ? AND id != ?", displayname, vars["id"])
		if err == nil {
			logger.WithField("error", err.Error()).Error("display name already in use")
			sessMgr.AddMessage("error", "This display name is already in use!")
			http.Redirect(w, r, "/admin/user/"+vars["id"]+"/edit", http.StatusSeeOther)
			return
		}

		userId, err := strconv.Atoi(vars["id"])
		if err != nil {
			logger.WithField("error", err.Error()).Error("invalid user id")
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
				logger.WithField("error", err.Error()).Error("could not hash password")
				w.WriteHeader(500)
				return
			}

			updateUser.Password = passwordHash
		}
		err = h.Ds.UpdateUser(updateUser)
		if err != nil {
			logger.WithField("error", err.Error()).Error("could not update user")
			w.WriteHeader(500)
			return
		}

		sessMgr.AddMessage("success", "Changes to user account saved!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithField("error", err.Error()).Error("invalid user id")
		sessMgr.AddMessage("error", "You supplied an invalid user id!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}
	editedUser, err := h.Ds.GetUserById(userId)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not scan user")
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
func (h *HttpHandler) AdminUserRemoveHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		err error
		sessMgr = h.SessMgr
		logger = h.ContextLogger("AdminUserRemoveHandler")
		currentUser = r.Context().Value("user").(entity.User)
		vars = mux.Vars(r)
	)


	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"userId": userId,
		}).Error("invalid user id")
		sessMgr.AddMessage("error", "You supplied an invalid user id!")
		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		err = h.Ds.DeleteUser(userId)
		if err != nil {
			logger.WithField("error", err.Error()).Error("error removing user")
			sessMgr.AddMessage("error", "An unknown error occurred, please try again.")
			http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/admin/user/list", http.StatusSeeOther)
		return
	}

	user, err := h.Ds.GetUserById(userId)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not scan user")
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
func (h *HttpHandler) AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		err error
		logger = h.ContextLogger("AdminSettingsHandler")
		currentUser = r.Context().Value("user").(entity.User)
		sessMgr = h.SessMgr
	)

	if r.Method == http.MethodPost {
		var errors uint8 = 0
		form := r.FormValue("form")
		if form == "general_settings" {
			baseDatapath := r.FormValue("base_datapath")
			err = h.Ds.SetSetting("base_datapath", baseDatapath)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "base_datapath",
				}).Error("could not save setting")
			}
			baseUrl := r.FormValue("base_url")
			err = h.Ds.SetSetting("base_url", baseUrl)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "base_url",
				}).Error("could not save setting")
			}
		} else if form == "security" {
			securityDisableRegistration := r.FormValue("security_disable_registration")
			if securityDisableRegistration != "1" {
				securityDisableRegistration = "0"
			}
			err = h.Ds.SetSetting("security_disable_registration", securityDisableRegistration)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "security_disable_registration",
				}).Error("could not save setting")
			}

			securityDisablePasswordReset := r.FormValue("security_disable_password_reset")
			if securityDisablePasswordReset != "1" {
				securityDisablePasswordReset = "0"
			}
			err = h.Ds.SetSetting("security_disable_password_reset", securityDisablePasswordReset)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "security_disable_password_reset",
				}).Error("could not save setting")
			}

			securityEmailConfirmationRequired := r.FormValue("security_email_confirmation_required")
			if securityEmailConfirmationRequired != "1" {
				securityEmailConfirmationRequired = "0"
			}
			err = h.Ds.SetSetting("security_email_confirmation_required", securityEmailConfirmationRequired)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "security_email_confirmation_required",
				}).Error("could not save setting")
			}

			security2fa := r.FormValue("security_2fa")
			if security2fa != "none" && security2fa != "email" && security2fa != "sms" {
				security2fa = "none"
			}
			err = h.Ds.SetSetting("security_2fa", security2fa)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "security_2fa",
				}).Error("could not save setting")
			}

		} else if form == "smtp" {
			smtpUsername := r.FormValue("smtp_username")
			err = h.Ds.SetSetting("smtp_username", smtpUsername)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "smtp_username",
				}).Error("could not save setting")
			}

			smtpPassword := r.FormValue("smtp_password")
			err = h.Ds.SetSetting("smtp_password", smtpPassword)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "smtp_password",
				}).Error("could not save setting")
			}

			smtpHost := r.FormValue("smtp_host")
			err = h.Ds.SetSetting("smtp_host", smtpHost)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "smtp_host",
				}).Error("could not save setting")
			}

			smtpPort := r.FormValue("smtp_port")
			err = h.Ds.SetSetting("smtp_port", smtpPort)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "smtp_port",
				}).Error("could not save setting")
			}

			smtpEncryption := r.FormValue("smtp_encryption")
			err = h.Ds.SetSetting("smtp_encryption", smtpEncryption)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "smtp_encryption",
				}).Error("could not save setting")
			}
		} else if form == "executables" {
			goExec := r.FormValue("golang_executable")
			err = h.Ds.SetSetting("golang_executable", goExec)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "golang_executable",
				}).Error("could not save setting")
			}

			dotnetExec := r.FormValue("dotnet_executable")
			err = h.Ds.SetSetting("dotnet_executable", dotnetExec)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "dotnet_executable",
				}).Error("could not save setting")
			}

			rustExec := r.FormValue("rust_executable")
			err = h.Ds.SetSetting("rust_executable", rustExec)
			if err != nil {
				errors++
				logger.WithFields(logrus.Fields{
					"error": err.Error(),
					"setting": "rust_executable",
				}).Error("could not save setting")
			}
		}

		if errors > 0 {
			output := fmt.Sprintf("When trying to save admin settings, %d error(s) occurred", errors)
			logger.Debug(output)
			sessMgr.AddMessage("error", output)
		} else {
			sessMgr.AddMessage("success", "Settings saved successfully!")
		}

		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
		return
	}
	allSettings, err := h.Ds.GetAllSettings()
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get allSettings: " + err.Error())
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
