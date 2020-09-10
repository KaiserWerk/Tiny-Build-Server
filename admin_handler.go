package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

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
		if form == "security" {
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
		}

		if errors > 0 {
			writeToConsole("When trying to save admin settings, 1 or more errors occured")
			// add flashbag
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
		CurrentUser user
		AdminSettings map[string]string
	}{
		currentUser,
		allSettings,
	}

	t := templates["admin_settings.html"]
	if t != nil {
		err := t.Execute(w, contextData)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildTargetListHandler(w http.ResponseWriter, r *http.Request) {
	// read, nigga, read!
	t := templates["admin_buildtarget_list.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildTargetAddHandler(w http.ResponseWriter, r *http.Request) {
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
		description := r.FormValue("description")
		if description != "" {
			db, err := getDbConnection()
			if err != nil {
				writeToConsole("could not get DB connection in adminBuildTargetAddHandler: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer db.Close()

			_, err = db.Exec("INSERT INTO build_target (description) VALUES (?)", description)
			if err != nil {
				writeToConsole("could not insert new build step: " + err.Error())
				// add flashbag
			}
		}
	}

	data := struct {
		CurrentUser		user
	} {
		CurrentUser: currentUser,
	}

	t := templates["admin_buildtarget_add.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildTargetEditHandler(w http.ResponseWriter, r *http.Request) {
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
		writeToConsole("could not establish DB connection in adminBuildTargetEditHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	row := db.QueryRow("SELECT id, description FROM build_target WHERE id = ?", vars["id"])
	var bt buildTarget
	err = row.Scan(&bt.Id, &bt.Description)
	if err != nil {
		writeToConsole("could not scan in adminBuildTargetEditHandler: " + err.Error())
		http.Redirect(w, r, "/admin/buildtarget/list", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := struct {
		CurrentUser		user
		BuildTarget		buildTarget
	} {
		CurrentUser: currentUser,
		BuildTarget: bt,
	}

	t := templates["admin_buildtarget_edit.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildTargetRemoveHandler(w http.ResponseWriter, r *http.Request) {

	t := templates["admin_buildtarget_remove.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildStepListHandler(w http.ResponseWriter, r *http.Request) {

	t := templates["admin_buildstep_list.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildStepAddHandler(w http.ResponseWriter, r *http.Request) {

	t := templates["admin_buildstep_add.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildStepEditHandler(w http.ResponseWriter, r *http.Request) {

	t := templates["admin_buildstep_edit.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildStepRemoveHandler(w http.ResponseWriter, r *http.Request) {

	t := templates["admin_buildstep_remove.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

