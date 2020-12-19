package helper

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/gomail.v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func WriteToConsole(s string) {
	fmt.Println("> " + s)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetDbConnection() *sql.DB {
	if db == nil {
		config := GetConfiguration()
		handle, err := sql.Open(config.Database.Driver, config.Database.DSN)
		if err != nil {
			panic(err.Error())
		}
		db = handle
	}
	return db
}

//func SendMail(m *gomail.Message) {
//	// fetch data from system configuration
//	d := gomail.NewDialer("smtp.example.com", 587, "user", "123456")
//	if err := d.DialAndSend(m); err != nil {
//		WriteToConsole("could not send email: " + err.Error())
//	}
//}

func Cleanup() {
	// close DB connection
	db := GetDbConnection()
	err := db.Close()
	if err != nil {
		WriteToConsole("could not close DB connection: " + err.Error())
	}
	// flush log writer
}

func GetHeaderIfSet(r *http.Request, key string) (string, error) {
	header := r.Header.Get(key)
	if header == "" {
		return "", errors.New("header is not set or empty")
	}
	return header, nil
}

func CheckPayloadRequest(r *http.Request) (entity.BuildDefinition, error) {
	// TODO: check client-supplied secret key?
	// get id
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		return entity.BuildDefinition{}, errors.New("could not determine ID of build definition")
	}
	// convert to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return entity.BuildDefinition{}, errors.New("invalid ID value supplied")
	}
	// get DB connection
	db := GetDbConnection()
	// fetch the build definition
	var bd entity.BuildDefinition
	row := db.QueryRow("SELECT id, build_target, build_target_os_arch, build_target_arm, altered_by, caption, "+
		"enabled, deployment_enabled, repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, "+
		"repo_branch, altered_at, apply_migrations, database_dns, meta_migration_id, run_tests, run_benchmark_tests "+
		"FROM build_definition WHERE id = ?", id)
	err = row.Scan(&bd.Id, &bd.BuildTarget, &bd.BuildTargetOsArch, &bd.BuildTargetArm, &bd.AlteredBy, &bd.Caption,
		&bd.Enabled, &bd.DeploymentEnabled,
		&bd.RepoHoster, &bd.RepoHosterUrl, &bd.RepoFullname, &bd.RepoUsername, &bd.RepoSecret, &bd.RepoBranch,
		&bd.AlteredAt, &bd.ApplyMigrations, &bd.DatabaseDSN, &bd.MetaMigrationId, &bd.RunTests,
		&bd.RunBenchmarkTests)
	if err != nil {
		return entity.BuildDefinition{}, errors.New("could not scan buildDefinition")
	}

	// check relevant headers and payload values
	switch bd.RepoHoster {
	case "bitbucket":
		headers := []string{"X-Event-RegToken", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("bitbucket: could not get bitbucket header " + headers[i])
			}
		}

		var payload entity.BitBucketPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("bitbucket: could not decode json payload")
		}
		if payload.Push.Changes[0].New.Name != bd.RepoBranch {
			return entity.BuildDefinition{}, errors.New("bitbucket: branch names do not match (" + payload.Push.Changes[0].New.Name + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return entity.BuildDefinition{}, errors.New("bitbucket: repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "github":
		headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("github: could not get github header " + headers[i])
			}
		}

		var payload entity.GitHubPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("github: could not decode json payload")
		}
		if payload.Repository.DefaultBranch != bd.RepoBranch {
			return entity.BuildDefinition{}, errors.New("github: branch names do not match (" + payload.Repository.DefaultBranch + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return entity.BuildDefinition{}, errors.New("github: repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "gitlab":
		headers := []string{"X-GitLab-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("gitlab: could not get gitlab header " + headers[i])
			}
		}

		var payload entity.GitLabPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("gitlab: could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != bd.RepoBranch {
			return entity.BuildDefinition{}, errors.New("gitlab: branch names do not match (" + branch + ")")
		}
		if payload.Project.PathWithNamespace != bd.RepoFullname {
			return entity.BuildDefinition{}, errors.New("gitlab: repository names do not match (" + payload.Project.PathWithNamespace + ")")
		}
	case "gitea":
		headers := []string{"X-Gitea-Delivery", "X-Gitea-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("gitea: could not get gitea header " + headers[i])
			}
		}

		var payload entity.GiteaPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("gitea: could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != bd.RepoBranch {
			return entity.BuildDefinition{}, errors.New("gitea: branch names do not match (" + branch + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return entity.BuildDefinition{}, errors.New("gitea: repository names do not match (" + payload.Repository.FullName + ")")
		}
	}

	return bd, nil
}

func ReadConsoleInput(externalShutdownCh chan os.Signal) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		if err != nil {
			fmt.Printf("  could not process input %v\n", input)
			continue
		}

		switch string(input) {
		case "":
			continue
		case "cluck":
			animal := `   \\
   (o>
\\_//)
 \_/_)
  _|_  
You found the chicken. Hooray!`
			fmt.Println(animal)
		case "shutdown":
			WriteToConsole("shutdown via console initiated...")
			time.Sleep(time.Second)
			externalShutdownCh <- os.Interrupt
		case "reload config":
			WriteToConsole("reloading configuration...")
			time.Sleep(time.Second)
			// TODO

			WriteToConsole("done")
		case "invalidate sessions":
			WriteToConsole("invalidating all sessions...")
			GetSessionManager().RemoveAllSessions()
			time.Sleep(time.Second)
			WriteToConsole("done")
		case "list sessions":
			WriteToConsole("all sessions:")
			for _, v := range GetSessionManager().Sessions {
				WriteToConsole("Id: " + v.Id + "\tLifetime:" + v.Lifetime.Format("2006-01-02 15:04:05"))
			}
		default:
			WriteToConsole("unrecognized command: " + string(input))
		}
	}
}

func SendEmail(messageType EmailMessageType, data interface{}, subject string, to []string) error {
	if len(to) == 0 {
		return fmt.Errorf("could not send email; no recipients supplied")
	}
	settings, err := GetAllSettings()
	if err != nil {
		WriteToConsole("could not get all settings: " + err.Error())
		return err
	}

	emailBody, err := ParseEmailTemplate(string(messageType), data)
	if err != nil {
		return fmt.Errorf("unable to parse email template: %s", err.Error())
	}

	m := gomail.NewMessage()
	m.SetHeader("From", settings["smtp_username"])
	m.SetHeader("To", to...)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", emailBody)
	//m.Attach("/home/Alex/lolcat.jpg")

	port, _ := strconv.Atoi(settings["smtp_port"])
	d := gomail.NewDialer(settings["smtp_host"], port, settings["smtp_username"], settings["smtp_password"])

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	//emailAuth := smtp.PlainAuth("", settings["smtp_username"], settings["smtp_password"], settings["smtp_host"])
	//
	//emailBody, err := ParseEmailTemplate(string(messageType), data)
	//if err != nil {
	//	return fmt.Errorf("unable to parse email template: %s", err.Error())
	//}
	//
	//mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	//subject := "Subject: " + "Test Email" + "!\n"
	//msg := []byte(subject + mime + "\n" + emailBody)
	//addr := fmt.Sprintf("%s:%s", settings["smtp_host"], settings["smtp_port"])
	//
	//if err := smtp.SendMail(addr, emailAuth, settings["smtp_username"], to, msg); err != nil {
	//	return err
	//}
	return nil
}
