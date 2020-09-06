package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)


func indexHandler(w http.ResponseWriter, r *http.Request) {
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
	latestBuilds, err := getNewestBuildExecutions(5)
	if err != nil {
		writeToConsole("could not fet latest build executions: " + err.Error())
	}
	latestBuildDefs, err := getNewestBuildDefinitions(5)
	if err != nil {
		writeToConsole("could not fet latest build executions: " + err.Error())
	}

	indexData := struct {
		CurrentUser     user
		LatestBuilds    []buildExecution
		LatestBuildDefs []buildDefinition
	}{
		CurrentUser:     currentUser,
		LatestBuilds:    latestBuilds,
		LatestBuildDefs: latestBuildDefs,
	}

	//otherwise ok (logged in)
	//writeToConsole("login check ok")
	t := templates["index.html"]
	if t != nil {
		err := t.Execute(w, indexData)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		email := r.FormValue("login_email")
		password := r.FormValue("login_password")
		u, err := getUserByEmail(email)
		if err != nil {
			writeToConsole("could not get user by Email (maybe doesnt exist): " + err.Error())
			return
		}
		fmt.Printf("User: %v\n", u)
		if doesHashMatch(password, u.Password) {
			writeToConsole("authenticated successfully")
			//continue settings cookie/starting session
			sess, err := sessMgr.CreateSession(time.Now().Add(30 * 24 * time.Hour))
			if err != nil {
				writeToConsole("could not create session: " + err.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			sess.SetVar("user_id", strconv.Itoa(u.Id))
			writeToConsole("session Id: " + sess.Id)
			err = sessMgr.SetCookie(w, sess.Id)
			if err != nil {
				writeToConsole("could not set cookie: " + err.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			writeToConsole("cookie set")
		} else {
			writeToConsole("login not successful")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		writeToConsole("redirect to index page")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t := templates["login.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	writeToConsole("getting cookie value")
	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		writeToConsole("could not get cookie value: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	//writeToConsole("getting session with Id "+sessId)
	session, err := sessMgr.GetSession(sessId)
	if err != nil {
		writeToConsole("could not get session: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = sessMgr.RemoveSession(session.Id)
	if err != nil {
		writeToConsole("could not remove session: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func requestNewPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		email := r.FormValue("login_email")
		if email != "" {
			u, err := getUserByEmail(email)
			if err != nil {
				writeToConsole("could not get user by Email (maybe doesnt exist): " + err.Error())
				return
			}

			writeToConsole("user: " + u.Displayname)
			// email an user versenden
			// zur reset seite weiterleiten
		}
	}

	t := templates["login.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	t := templates["password_request.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func registrationHandler(w http.ResponseWriter, r *http.Request) {
	t := templates["register.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
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

func buildDefinitionListHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["build_definition_list.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionAddHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["build_definition_add.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionEditHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["build_definition_edit.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionShowHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r) // id
	db, err := getDbConnection()
	if err != nil {
		writeToConsole("could not get DB connection: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var bd buildDefinition
	row := db.QueryRow("SELECT * FROM build_definition WHERE id = ?", vars["id"])
	err = row.Scan(&bd.Id, &bd.AlteredBy, &bd.Caption, &bd.Enabled, &bd.DeploymentEnabled, &bd.RepoHoster, &bd.RepoHosterUrl,
		&bd.RepoFullname, &bd.RepoUsername, &bd.RepoSecret, &bd.RepoBranch, &bd.AlteredAt)
	if err != nil {
		writeToConsole("could not scan buildDefinition: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var be buildExecution
	var beList = make([]buildExecution, 0)
	rows, err := db.Query("SELECT id, result, execution_time, executed_at FROM build_execution WHERE " +
		"build_definition_id = ? ORDER BY executed_at DESC", bd.Id)
	if err != nil {
		writeToConsole("could not fetch most recent build executions: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		err = rows.Scan(&be.Id, &be.Result, &be.ExecutionTime, &be.ExecutedAt)
		if err != nil {
			writeToConsole("could not scan build execution: " + err.Error())
			continue
		}
		beList = append(beList, be)
		be = buildExecution{}
	}

	failedBuildCount := 0
	successBuildCount := 0
	avg := 0.0
	i := 0
	for _, v := range beList {
		if v.Result == "success" {
			successBuildCount++
		}
		if v.Result == "failed" {
			failedBuildCount++
		}
		avg += v.ExecutionTime
		i++
	}
	//writeToConsole("total execution time: " + fmt.Sprintf("%v in %v iterations", avg, i))
	avg = avg / float64(i)



	successRate := float64(successBuildCount) / float64(i) * 100


	var recentExecutions []buildExecution
	if len(beList) >= 5 {
		recentExecutions = beList[:5]
	} else {
		for _, v := range beList {
			recentExecutions = append(recentExecutions, v)
		}
	}

	data := struct {
		BuildDefinition buildDefinition
		RecentExecutions []buildExecution
		CurrentUser user
		TotalBuildCount int
		FailedBuildCount int
		SuccessBuildCount int
		SuccessRate	string
		AvgRuntime string
	} {
		BuildDefinition: bd,
		RecentExecutions: recentExecutions,
		CurrentUser: currentUser,
		TotalBuildCount: len(beList),
		FailedBuildCount: failedBuildCount,
		SuccessBuildCount: successBuildCount,
		SuccessRate: fmt.Sprintf("%.2f", successRate),
		AvgRuntime: fmt.Sprintf("%.2f", avg),
	}

	t := templates["build_definition_show.html"]
	if t != nil {
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		writeToConsole("template build_definition_show.html not found")
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionRemoveHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["build_definition_remove.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildDefinitionListExecutionsHandler(w http.ResponseWriter, r *http.Request) {

}

func buildExecutionListHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["build_execution_list.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildExecutionShowHandler(w http.ResponseWriter, r *http.Request) {


	t := templates["build_execution_show.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

