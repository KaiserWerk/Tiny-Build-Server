package mailer

import (
	"errors"
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
	"gopkg.in/gomail.v2"
)

// EmailSubject defines different types of email messages
type EmailSubject string

const (
	SubjAccountCreated        EmailSubject = "Tiny Build Server: Your account was created"
	SubjAccountLocked         EmailSubject = "Tiny Build Server: Your account has been locked"
	SubjConfirmRegistration   EmailSubject = "Tiny Build Server: Please confirm your registration"
	SubjRequestNewPassword    EmailSubject = "Tiny Build Server: Instructions on how to reset your password"
	SubjRegistrationConfirmed EmailSubject = "Tiny Build Server: Your registration was successfully confirmed"
	SubjConfirmPasswordReset  EmailSubject = "Tiny Build Server: Your password has been reset"
	SubjNewDeployment         EmailSubject = "Tiny Build Server: New email deployment"
)

var (
	ErrMissingSettings = errors.New("the mailer settings are nil or incomplete")
	ErrNoRecipient     = errors.New("no recipient was supplied")
)

type Mailer struct {
	Settings map[string]string
}

// SendEmail sends and email
func (mailer *Mailer) SendEmail(body string, subject string, to, attachments []string) error {
	if mailer.Settings == nil || mailer.Settings["smtp_host"] == "" {
		return ErrMissingSettings
	}
	if to == nil || len(to) == 0 {
		return ErrNoRecipient
	}

	m := gomail.NewMessage()
	m.SetHeader("From", mailer.Settings["smtp_username"])
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

	port, err := strconv.Atoi(mailer.Settings["smtp_port"])
	if err != nil {
		return err
	}
	d := gomail.NewDialer(mailer.Settings["smtp_host"], port, mailer.Settings["smtp_username"], mailer.Settings["smtp_password"])
	//d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

// SendEmail2 sends and email, as well
func SendEmail2(settings map[string]string, body string, subject string, to, attachments []string) error {
	e := email.NewEmail()
	e.From = settings["smtp_username"]
	e.To = to
	//e.Bcc = []string{"test_bcc@example.com"}
	//e.Cc = []string{"test_cc@example.com"}
	e.Subject = subject
	//e.Text = []byte("Text Body is, of course, supported!")
	e.HTML = []byte(body)
	if attachments != nil && len(attachments) > 0 {
		for _, v := range attachments {
			_, err := e.AttachFile(v)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println("settings:", settings)
	return e.Send(fmt.Sprintf("%s:%s", settings["smtp_host"], settings["smtp_port"]), smtp.PlainAuth("", settings["smtp_username"], settings["smtp_password"], settings["smtp_host"]))
}
