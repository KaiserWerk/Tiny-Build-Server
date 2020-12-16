package helper

type EmailMessageType string

const (
	AccountCreated        EmailMessageType = "account_created"
	AccountLocked         EmailMessageType = "account_locked"
	ConfirmRegistration   EmailMessageType = "confirm_registration"
	PasswordReset         EmailMessageType = "password_reset"
	RegistrationConfirmed EmailMessageType = "registration_confirmed"
	ResetPassword         EmailMessageType = "reset_password"
)

