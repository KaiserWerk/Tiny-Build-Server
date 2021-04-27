package helper

import (
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	"gopkg.in/gomail.v2"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"
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

func SendEmail(settings map[string]string, body string, subject string, to, attachments []string) error {
	if to == nil || len(to) == 0 {
		return fmt.Errorf("could not send email; no recipients supplied")
	}
	//fmt.Println("settings:", settings)
	m := gomail.NewMessage()
	m.SetHeader("From", settings["smtp_username"])
	m.SetHeader("To", to...)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	if attachments != nil && len(attachments) > 0 {
		for _, v := range attachments {
			m.Attach(v)
		}
	}
	//m.Attach("/home/Alex/lolcat.jpg")

	port, err := strconv.Atoi(settings["smtp_port"])
	if err != nil {
		return err
	}
	d := gomail.NewDialer(settings["smtp_host"], port, settings["smtp_username"], settings["smtp_password"])
	//d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func SendEmail2(settings map[string]string, body string, subject string, to, attachments []string) error {
	e := email.NewEmail()
	e.From = settings["smtp_username"]
	e.To = to
	//e.Bcc = []string{"test_bcc@example.com"}
	//e.Cc = []string{"test_cc@example.com"}
	e.Subject = subject
	e.Text = []byte("Text Body is, of course, supported!")
	e.HTML = []byte(body)
	if attachments != nil && len(attachments) > 0 {
		for _, v := range attachments {
			_, err := e.AttachFile(v)
			if err != nil {
				WriteToConsole("could not attach file " + v)
			}
		}
	}
	fmt.Println("settings:", settings)
	return e.Send(fmt.Sprintf("%s:%s", settings["smtp_host"], settings["smtp_port"]), smtp.PlainAuth("", settings["smtp_username"], settings["smtp_password"], settings["smtp_host"]))
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
