package helper

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/sessionstore"
	_ "github.com/go-sql-driver/mysql"
)

var GolangRuntimes = []string{
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
var DotnetRuntimes = []string{
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

func GetUserByEmail(n string) (entity.User, error) {
	db := GetDbConnection()
	row := db.QueryRow("SELECT id, displayname, email, password, locked, admin FROM user WHERE email = ?", n)
	var u entity.User
	err := row.Scan(&u.Id, &u.Displayname, &u.Email, &u.Password, &u.Locked, &u.Admin)
	if err != nil {
		return entity.User{}, errors.New("could not scan (" + err.Error() + ")")
	}

	return u, nil
}

func GetBuildDefCaption(id int) (string, error) {
	db := GetDbConnection()
	var name string
	row := db.QueryRow("SELECT caption FROM build_definition WHERE id = ?", id)
	err := row.Scan(&name)
	if err != nil {
		return "", errors.New("could not scan: " + err.Error())
	}
	return name, nil
}

func GetUserById(id int) (entity.User, error) {
	var u entity.User
	db := GetDbConnection()
	row := db.QueryRow("SELECT Id, Displayname, Email, Admin FROM user WHERE Id = ?", id)
	err := row.Scan(&u.Id, &u.Displayname, &u.Email, &u.Admin)
	if err != nil {
		return u, err
	}

	return u, nil
}

func GetUserFromSession(s sessionstore.Session) (entity.User, error) {
	userIdStr, ok := s.GetVar("user_id")
	if !ok {
		return entity.User{}, nil
	}

	userId, _ := strconv.Atoi(userIdStr)
	user, err := GetUserById(userId)
	return user, err
}

func GetUsernameById(id int) string {
	var u entity.User
	db := GetDbConnection()

	row := db.QueryRow("SELECT Id, Displayname, Email, Admin FROM user WHERE Id = ?", id)
	err := row.Scan(&u.Id, &u.Displayname, &u.Email, &u.Admin)
	if err != nil {
		return "not found"
	}

	return u.Displayname
}

func GetNewestBuildExecutions(limit int) ([]entity.BuildExecution, error) {
	var be entity.BuildExecution
	var beList []entity.BuildExecution

	db := GetDbConnection()
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
		be = entity.BuildExecution{}
	}

	return beList, nil
}

func GetNewestBuildDefinitions(limit int) ([]entity.BuildDefinition, error) {
	var bd entity.BuildDefinition
	var bdList []entity.BuildDefinition

	db := GetDbConnection()
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
		bd = entity.BuildDefinition{}
	}

	return bdList, nil
}

func GetAllSettings() (map[string]string, error) {
	settings := make(map[string]string)
	db := GetDbConnection()

	rows, err := db.Query("SELECT setting_name, setting_value FROM setting")
	if err != nil {
		return settings, err
	}

	var setting entity.AdminSetting
	for rows.Next() {
		err = rows.Scan(&setting.Name, &setting.Value)
		if err != nil {
			return settings, err
		}
		settings[setting.Name] = setting.Value
		setting = entity.AdminSetting{}
	}

	return settings, nil
}

func SetSetting(name, value string) error {
	db := GetDbConnection()
	row := db.QueryRow("SELECT setting_name, setting_value FROM setting WHERE setting_name = ?", name)
	var s entity.AdminSetting
	err := row.Scan(&s.Name, &s.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			//helper.WriteToConsole("no row, inserting")
			_, err = db.Exec("INSERT INTO setting (setting_name, setting_value) VALUES (?, ?)", name, value)
			if err != nil {
				return err
			}
		} else {
			//helper.WriteToConsole("row found, updating")
			//_, err = db.Exec("UPDATE setting SET setting_value = ? WHERE setting_name = ?", value, name)
			//if err != nil {
			//	return err
			//}
			return err
		}
	} else { // brauch ich den Zweig?
		//helper.WriteToConsole("row found, updating (2)")
		_, err = db.Exec("UPDATE setting SET setting_value = ? WHERE setting_name = ?", value, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetBuildTargets() ([]entity.BuildTarget, error) {
	db := GetDbConnection()
	var btList []entity.BuildTarget
	rows, err := db.Query("SELECT id, caption FROM build_target")
	if err != nil {
		return nil, errors.New("could not get buildTargets in getBuildTargets: " + err.Error())
	} else {
		var bt entity.BuildTarget
		for rows.Next() {
			err = rows.Scan(&bt.Id, &bt.Description)
			if err != nil {
				WriteToConsole("could not scan in getBuildTargets: " + err.Error())
				continue
			}
			btList = append(btList, bt)
			bt = entity.BuildTarget{}
		}
	}

	return btList, nil
}

func GetBuildStepsForTarget(id int) ([]entity.BuildStep, error) {
	db := GetDbConnection()
	var bsList []entity.BuildStep
	rows, err := db.Query("SELECT id, build_target_id, caption, command, enabled FROM build_step WHERE "+
		"enabled = 1 AND build_target_id = ?",
		id)
	if err != nil {
		return nil, errors.New("could not get buildSteps in getBuildStepsForTarget: " + err.Error())
	} else {
		var bs entity.BuildStep
		for rows.Next() {
			err = rows.Scan(&bs.Id, &bs.BuildTargetId, &bs.Caption, &bs.Command, &bs.Enabled)
			if err != nil {
				WriteToConsole("could not scan in getBuildStepsForTarget: " + err.Error())
				continue
			}
			bsList = append(bsList, bs)
			bs = entity.BuildStep{}
		}
	}

	return bsList, nil
}
