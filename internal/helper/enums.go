package helper

type EmailMessageType string

const (
	AccountCreatedEmail        EmailMessageType = "account_created"
	AccountLockedEmail         EmailMessageType = "account_locked"
	ConfirmRegistrationEmail   EmailMessageType = "confirm_registration"
	RequestNewPasswordEmail    EmailMessageType = "request_new_password"
	RegistrationConfirmedEmail EmailMessageType = "registration_confirmed"
	ConfirmPasswordResetEmail  EmailMessageType = "confirm_password_reset"
)


