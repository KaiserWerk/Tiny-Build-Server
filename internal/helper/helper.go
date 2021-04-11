package helper

import (
	"errors"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/fixtures"
	"gopkg.in/gomail.v2"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

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

func GetHeaderIfSet(r *http.Request, key string) (string, error) {
	header := r.Header.Get(key)
	if header == "" {
		return "", errors.New("header is not set or empty")
	}
	return header, nil
}

//func ReadConsoleInput(externalShutdownCh chan os.Signal) {
//	reader := bufio.NewReader(os.Stdin)
//	for {
//		input, _, err := reader.ReadLine()
//		if err != nil {
//			fmt.Printf("  could not process input %v\n", input)
//			continue
//		}
//
//		switch string(input) {
//		case "":
//			continue
//		case "cluck":
//			animal := `   \\
//   (o>
//\\_//)
// \_/_)
//  _|_
//You found the chicken. Hooray!`
//			fmt.Println(animal)
//		case "shutdown":
//			WriteToConsole("shutdown via console initiated...")
//			time.Sleep(time.Second)
//			externalShutdownCh <- os.Interrupt
//		case "reload config":
//			WriteToConsole("reloading configuration...")
//			time.Sleep(time.Second)
//			// TODO
//
//			WriteToConsole("done")
//		case "invalidate sessions":
//			WriteToConsole("invalidating all sessions...")
//			global.GetSessionManager().RemoveAllSessions()
//			time.Sleep(time.Second)
//			WriteToConsole("done")
//		case "list sessions":
//			WriteToConsole("all sessions:")
//			for _, v := range global.GetSessionManager().Sessions {
//				WriteToConsole("Id: " + v.Id + "\tLifetime:" + v.Lifetime.Format("2006-01-02 15:04:05"))
//			}
//		default:
//			WriteToConsole("unrecognized command: " + string(input))
//		}
//	}
//}

func SendEmail(messageType fixtures.EmailMessageType, settings map[string]string, body string, subject string, to []string) error {
	if len(to) == 0 {
		return fmt.Errorf("could not send email; no recipients supplied")
	}
	//settings, err := sessionService.GetAllSettings(global.GetDbConnection())
	//if err != nil {
	//	WriteToConsole("could not get all settings: " + err.Error())
	//	return err
	//}

	//emailBody, err := templateservice.ParseEmailTemplate(string(messageType), data)
	//if err != nil {
	//	return fmt.Errorf("unable to parse email template: %s", err.Error())
	//}

	m := gomail.NewMessage()
	m.SetHeader("From", settings["smtp_username"])
	m.SetHeader("To", to...)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	//m.Attach("/home/Alex/lolcat.jpg")

	port, err := strconv.Atoi(settings["smtp_port"])
	if err != nil {
		return err
	}
	d := gomail.NewDialer(settings["smtp_host"], port, settings["smtp_username"], settings["smtp_password"])
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
