package main

import (
	"database/sql"
	"errors"
	"github.com/KaiserWerk/sessionstore"
	"strconv"
)

var golangRuntimes = []string{
	"aix/ppc64",
	"android/386",
	"android/amd64",
	"android/arm",
	"android/arm64",
	"darwin/386",
	"darwin/amd64",
	"darwin/arm",
	"darwin/arm64",
	"dragonfly/amd64",
	"freebsd/386",
	"freebsd/amd64",
	"freebsd/arm",
	"freebsd/arm64",
	"illumos/amd64",
	"js/wasm",
	"linux/386",
	"linux/amd64",
	"linux/arm",
	"linux/arm64",
	"linux/mips",
	"linux/mips64",
	"linux/mips64le",
	"linux/mipsle",
	"linux/ppc64",
	"linux/ppc64le",
	"linux/riscv64",
	"linux/s390x",
	"netbsd/386",
	"netbsd/amd64",
	"netbsd/arm",
	"netbsd/arm64",
	"openbsd/386",
	"openbsd/amd64",
	"openbsd/arm",
	"openbsd/arm64",
	"plan9/386",
	"plan9/amd64",
	"plan9/arm",
	"solaris/amd64",
	"windows/386",
	"windows/amd64",
	"windows/arm",
}
var dotnetRuntimes = []string{
	"win-x64",
	"win-x86",
	"win-arm",
	"win-arm64",
	"win7-x64",
	"win7-x86",
	"win81-x64",
	"win81-x86",
	"win81-arm",
	"win10-x64",
	"win10-x86",
	"win10-arm",
	"win10-arm64",
	"linux-x64",
	"linux-musl-x64",
	"linux-arm",
	"linux-arm64",
	"rhel-x64",
	"rhel.6-x64",
	"tizen",
	"tizen.4.0.0",
	"tizen.5.0.0",
	"osx-x64",
	"osx.10.10-x64",
	"osx.10.11-x64",
	"osx.10.12-x64",
	"osx.10.13-x64",
	"osx.10.14-x64",
}

func getUserByEmail(n string) (user, error) {
	db, err := getDbConnection()
	if err != nil {
		return user{}, errors.New("could not get database connection")
	}
	defer db.Close()
	row := db.QueryRow("SELECT id, displayname, email, password, locked, admin FROM user WHERE email = ?", n)
	var u user
	//var Locked int
	//var Admin int
	err = row.Scan(&u.Id, &u.Displayname, &u.Email, &u.Password, &u.Locked, &u.Admin)
	if err != nil {
		return user{}, errors.New("could not scan")
	}

	return u, nil
}

func getBuildDefCaption(id int) string {
	db, err := getDbConnection()
	if err != nil {
		return "could not fetch"
	}
	defer db.Close()
	var name string
	row := db.QueryRow("SELECT caption FROM build_definition WHERE id = ?", id)
	err = row.Scan(&name)
	if err != nil {
		return "could not scan"
	}
	return name
}

func getUserById(id int) (user, error) {
	var u user
	db, err := getDbConnection()
	if err != nil {
		return u, err
	}
	defer db.Close()
	row := db.QueryRow("SELECT Id, Displayname, Email, Admin FROM user WHERE Id = ?", id)
	err = row.Scan(&u.Id, &u.Displayname, &u.Email, &u.Admin)
	if err != nil {
		return u, err
	}

	return u, nil
}

func getUserFromSession(s sessionstore.Session) (user, error) {
	userIdStr, ok := s.GetVar("user_id")
	if !ok {
		return user{}, nil
	}

	userId, _ := strconv.Atoi(userIdStr)
	user, err := getUserById(userId)
	return user, err
}

func getUsernameById(id int) string {
	var u user
	db, err := getDbConnection()
	if err != nil {
		return "not found"
	}
	defer db.Close()
	row := db.QueryRow("SELECT Id, Displayname, Email, Admin FROM user WHERE Id = ?", id)
	err = row.Scan(&u.Id, &u.Displayname, &u.Email, &u.Admin)
	if err != nil {
		return "not found"
	}

	return u.Displayname
}

func getNewestBuildExecutions(limit int) ([]buildExecution, error) {
	var be buildExecution
	var beList []buildExecution

	db, err := getDbConnection()
	if err != nil {
		return beList, err
	}
	defer db.Close()
	query := "SELECT id, build_definition_id, initiated_by, manual_run, action_log, result, artifact_path, " +
		"execution_time, executed_at FROM build_execution ORDER BY executed_at DESC"
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}
	rows, err := db.Query(query)
	if err != nil {
		return beList, err
	}

	for rows.Next() {
		err = rows.Scan(&be.Id, &be.BuildDefinitionId, &be.InitiatedBy, &be.ManualRun,
			&be.ActionLog, &be.Result, &be.ArtifactPath, &be.ExecutionTime, &be.ExecutedAt)
		if err != nil {
			return beList, err
		}

		beList = append(beList, be)
		be = buildExecution{}
	}

	return beList, nil
}

func getNewestBuildDefinitions(limit int) ([]buildDefinition, error) {
	var bd buildDefinition
	var bdList []buildDefinition

	db, err := getDbConnection()
	if err != nil {
		return bdList, err
	}
	defer db.Close()
	query := "SELECT id, build_target_id, altered_by, caption, enabled, deployment_enabled, repo_hoster, repo_hoster_url, " +
		"repo_fullname, repo_username, repo_secret, repo_branch, altered_at FROM build_definition ORDER BY altered_at DESC"
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}
	rows, err := db.Query(query)
	if err != nil {
		return bdList, err
	}

	for rows.Next() {
		err = rows.Scan(&bd.Id, &bd.BuildTargetId, &bd.AlteredBy, &bd.Caption, &bd.Enabled, &bd.DeploymentEnabled, &bd.RepoHoster, &bd.RepoHosterUrl,
			&bd.RepoFullname, &bd.RepoUsername, &bd.RepoSecret, &bd.RepoBranch, &bd.AlteredAt)
		if err != nil {
			return bdList, err
		}

		bdList = append(bdList, bd)
		bd = buildDefinition{}
	}

	return bdList, nil
}

func getAllSettings() (map[string]string, error) {
	settings := make(map[string]string)
	db, err := getDbConnection()
	if err != nil {
		return settings, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT setting_name, setting_value FROM setting")
	if err != nil {
		return settings, err
	}

	var setting adminSetting
	for rows.Next() {
		err = rows.Scan(&setting.Name, &setting.Value)
		if err != nil {
			return settings, err
		}
		settings[setting.Name] = setting.Value
		setting = adminSetting{}
	}
	//fmt.Println("settings:", settings)

	return settings, nil
}

func setSetting(name, value string) error {
	db, err := getDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow("SELECT setting_name, setting_value FROM setting WHERE setting_name = ?", name)
	var s adminSetting
	err = row.Scan(&s.Name, &s.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			writeToConsole("no row, inserting")
			_, err = db.Exec("INSERT INTO setting (setting_name, setting_value) VALUES (?, ?)", name, value)
			if err != nil {
				return err
			}
		} else {
			//writeToConsole("row found, updating")
			//_, err = db.Exec("UPDATE setting SET setting_value = ? WHERE setting_name = ?", value, name)
			//if err != nil {
			//	return err
			//}
			return err
		}
	} else { // brauch ich den Zweig?
		writeToConsole("row found, updating (2)")
		_, err = db.Exec("UPDATE setting SET setting_value = ? WHERE setting_name = ?", value, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func getBuildTargets() ([]buildTarget, error) {
	db, err := getDbConnection()
	if err != nil {
		return nil, errors.New("could not get DB connection in getBuildTargets: " + err.Error())
	}
	defer db.Close()

	var btList []buildTarget
	rows, err := db.Query("SELECT id, caption FROM build_target")
	if err != nil {
		return nil, errors.New("could not get buildTargets in getBuildTargets: " + err.Error())
	} else {
		var bt buildTarget
		for rows.Next() {
			err = rows.Scan(&bt.Id, &bt.Description)
			if err != nil {
				writeToConsole("could not scan in getBuildTargets: " + err.Error())
				continue
			}
			btList = append(btList, bt)
			bt = buildTarget{}
		}
	}

	return btList, nil
}

func getBuildStepsForTarget(id int) ([]buildStep, error) {
	db, err := getDbConnection()
	if err != nil {
		return nil, errors.New("could not get DB connection in getBuildStepsForTarget: " + err.Error())
	}
	defer db.Close()

	var bsList []buildStep
	rows, err := db.Query("SELECT id, build_target_id, caption, command, enabled FROM build_step WHERE "+
		"enabled = 1 AND build_target_id = ?",
		id)
	if err != nil {
		return nil, errors.New("could not get buildSteps in getBuildStepsForTarget: " + err.Error())
	} else {
		var bs buildStep
		for rows.Next() {
			err = rows.Scan(&bs.Id, &bs.BuildTargetId, &bs.Caption, &bs.Command, &bs.Enabled)
			if err != nil {
				writeToConsole("could not scan in getBuildStepsForTarget: " + err.Error())
				continue
			}
			bsList = append(bsList, bs)
			bs = buildStep{}
		}
	}

	return bsList, nil
}
