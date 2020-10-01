package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
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
		CurrentUser   user
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
		writeToConsole("could not get DB connection in adminBuildTargetListHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var btList []buildTarget
	rows, err := db.Query("SELECT id, caption FROM build_target")
	if err != nil {
		writeToConsole("could not get buildTargets: " + err.Error())
	} else {
		var bt buildTarget
		for rows.Next() {
			err = rows.Scan(&bt.Id, &bt.Description)
			if err != nil {
				writeToConsole("could not scan in adminBuildTargetListHandler: " + err.Error())
				continue
			}
			btList = append(btList, bt)
			bt = buildTarget{}
		}
	}
	data := struct {
		CurrentUser  user
		BuildTargets []buildTarget
	}{
		CurrentUser:  currentUser,
		BuildTargets: btList,
	}

	t := templates["admin_buildtarget_list.html"]
	if t != nil {
		err := t.Execute(w, data)
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

			_, err = db.Exec("INSERT INTO build_target (caption) VALUES (?)", description)
			if err != nil {
				writeToConsole("could not insert new build step: " + err.Error())
				sessMgr.AddMessage("error", "An error occured.")
			}

			http.Redirect(w, r, "/admin/buildtarget/list", http.StatusSeeOther)
			return
		} else {
			sessMgr.AddMessage("error", "You need so supply a description.")
		}
	}

	data := struct {
		CurrentUser user
	}{
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
		writeToConsole("could not fetch user from session")
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
	defer db.Close()

	row := db.QueryRow("SELECT id, caption FROM build_target WHERE id = ?", vars["id"])
	var bt buildTarget
	err = row.Scan(&bt.Id, &bt.Description)
	if err != nil {
		writeToConsole("could not scan in adminBuildTargetEditHandler: " + err.Error())
		http.Redirect(w, r, "/admin/buildtarget/list", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

		id := r.FormValue("id")
		description := r.FormValue("description")
		_, err = db.Exec("UPDATE build_target SET caption = ? WHERE id = ?", description, id)
		if err != nil {
			writeToConsole("could not update buildtarget: " + err.Error())
		}

		http.Redirect(w, r, "/admin/buildtarget/list", http.StatusSeeOther)
		return
	}

	data := struct {
		CurrentUser user
		BuildTarget buildTarget
	}{
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
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user from session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	sessMgr.AddMessage("error", "Could not be removed (not implemented).")
	// check if build steps exist
	// check if build definitions exist
	// check if build executions exist
	writeToConsole("adminBuildTargetRemoveHandler. not implemented yet")
	//vars := mux.Vars(r)
	//
	//db, err := getDbConnection()
	//if err != nil {
	//	writeToConsole("could not establish DB connection in adminBuildTargetRemoveHandler: " + err.Error())
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//defer db.Close()

	data := struct {
		CurrentUser user
	}{
		CurrentUser: currentUser,
	}

	t := templates["admin_buildtarget_remove.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildStepListHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user from session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not establish DB connection in adminBuildStepListHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var btList []buildTarget
	var bt buildTarget
	rowsBt, err := db.Query("SELECT * FROM build_target")
	if err != nil {
		writeToConsole("could not query bt rows in adminBuildStepListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	for rowsBt.Next() {
		err = rowsBt.Scan(&bt.Id, &bt.Description)
		if err != nil {
			writeToConsole("could not scan bt in adminBuildStepListHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		btList = append(btList, bt)
		bt = buildTarget{}
	}

	type preparedBuildStep struct {
		Id          int
		Description string
		Caption     string
		Command     string
		Enabled     bool
	}

	var bsList []preparedBuildStep
	var bs preparedBuildStep
	var rowsBs *sql.Rows
	target := r.URL.Query().Get("target")
	if target != "" {
		rowsBs, err = db.Query("SELECT bs.id, bt.description, bs.caption, bs.command, bs.enabled FROM build_step bs, "+
			"build_target bt WHERE bs.build_target_id = bt.id AND build_target_id = ?", target)
	} else {
		rowsBs, err = db.Query("SELECT bs.id, bt.description, bs.caption, bs.command, bs.enabled FROM build_step bs, " +
			"build_target bt WHERE bs.build_target_id = bt.id")
	}
	if err != nil {
		writeToConsole("could not query bs rows in adminBuildStepListHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}
	for rowsBs.Next() {
		err = rowsBs.Scan(&bs.Id, &bs.Description, &bs.Caption, &bs.Command, &bs.Enabled)
		if err != nil {
			writeToConsole("could not scan bs in adminBuildStepListHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		bsList = append(bsList, bs)
		bs = preparedBuildStep{}
	}

	targetId, _ := strconv.Atoi(target)

	data := struct {
		CurrentUser  user
		BuildTargets []buildTarget
		BuildSteps   []preparedBuildStep
		TargetId     int
	}{
		CurrentUser:  currentUser,
		BuildTargets: btList,
		BuildSteps:   bsList,
		TargetId:     targetId,
	}

	t := templates["admin_buildstep_list.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildStepAddHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user from session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not establish DB connection in adminBuildStepAddHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if r.Method == http.MethodPost {

		targetId := r.FormValue("target_id")
		caption := r.FormValue("caption")
		command := r.FormValue("command")
		var enabled bool
		if r.FormValue("enabled") == "1" {
			enabled = true
		} else {
			enabled = false
		}

		if caption == "" || command == "" {
			sessMgr.AddMessage("error", "You must supply a caption and a command.")
			http.Redirect(w, r, "/admin/buildstep/add", http.StatusSeeOther)
			return
		}

		_, err = db.Exec("INSERT INTO build_step (build_target_id, caption, command, enabled) VALUES (?, ?, ?, ?)",
			targetId, caption, command, enabled)

		http.Redirect(w, r, "/admin/buildstep/list", http.StatusSeeOther)
		return
	}

	var btList []buildTarget
	var bt buildTarget
	rowsBt, err := db.Query("SELECT * FROM build_target")
	if err != nil {
		writeToConsole("could not query bt rows in adminBuildStepAddHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	for rowsBt.Next() {
		err = rowsBt.Scan(&bt.Id, &bt.Description)
		if err != nil {
			writeToConsole("could not scan bt in adminBuildStepAddHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		btList = append(btList, bt)
		bt = buildTarget{}
	}

	data := struct {
		CurrentUser  user
		BuildTargets []buildTarget
	}{
		CurrentUser:  currentUser,
		BuildTargets: btList,
	}

	t := templates["admin_buildstep_add.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func adminBuildStepEditHandler(w http.ResponseWriter, r *http.Request) {
	session, err := checkLogin(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	currentUser, err := getUserFromSession(session)
	if err != nil {
		writeToConsole("could not fetch user from session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if !currentUser.Admin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not establish DB connection in adminBuildStepEditHandler: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	vars := mux.Vars(r)

	if r.Method == http.MethodPost {
		id := vars["id"]
		targetId := r.FormValue("target_id")
		caption := r.FormValue("caption")
		command := r.FormValue("command")
		var enabled bool
		if r.FormValue("enabled") == "1" {
			enabled = true
		} else {
			enabled = false
		}

		if caption == "" || command == "" {
			sessMgr.AddMessage("error", "You must supply a caption and a command.")
			http.Redirect(w, r, "/admin/buildstep/"+id+"/edit", http.StatusSeeOther)
			return
		}

		_, err = db.Exec("UPDATE build_step SET build_target_id = ?, caption = ?, command = ?, enabled = ? WHERE id = ?",
			targetId, caption, command, enabled, id)

		http.Redirect(w, r, "/admin/buildstep/list", http.StatusSeeOther)
		return
	}

	var btList []buildTarget
	var bt buildTarget
	rowsBt, err := db.Query("SELECT * FROM build_target")
	if err != nil {
		writeToConsole("could not query bt rows in adminBuildStepEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	for rowsBt.Next() {
		err = rowsBt.Scan(&bt.Id, &bt.Description)
		if err != nil {
			writeToConsole("could not scan bt in adminBuildStepEditHandler: " + err.Error())
			w.WriteHeader(500)
			return
		}
		btList = append(btList, bt)
		bt = buildTarget{}
	}

	var bs buildStep
	row := db.QueryRow("SELECT id, build_target_id, caption, command, enabled FROM build_step WHERE id = ?",
		vars["id"])
	err = row.Scan(&bs.Id, &bs.BuildTargetId, &bs.Caption, &bs.Command, &bs.Enabled)
	if err != nil {
		writeToConsole("could not scan bs in adminBuildStepEditHandler: " + err.Error())
		w.WriteHeader(500)
		return
	}

	tid, _ := strconv.Atoi(vars["id"])

	data := struct {
		CurrentUser  user
		BuildTargets []buildTarget
		BuildStep    buildStep
		TargetId     int
	}{
		CurrentUser:  currentUser,
		BuildTargets: btList,
		BuildStep:    bs,
		TargetId:     tid,
	}

	t := templates["admin_buildstep_edit.html"]
	if t != nil {
		err := t.Execute(w, data)
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
