package fixtures

// EmailMessageType defines different types of email messages
type EmailMessageType string

const (
	AccountCreatedEmail        EmailMessageType = "account_created"
	AccountLockedEmail         EmailMessageType = "account_locked"
	ConfirmRegistrationEmail   EmailMessageType = "confirm_registration"
	RequestNewPasswordEmail    EmailMessageType = "request_new_password"
	RegistrationConfirmedEmail EmailMessageType = "registration_confirmed"
	ConfirmPasswordResetEmail  EmailMessageType = "confirm_password_reset"
	DeploymentEmail            EmailMessageType = "deployment"
	Test                       EmailMessageType = "test"
)

var EmailSubjects = map[EmailMessageType]string{
	AccountCreatedEmail:        "Tiny Build Server: Your account was created",
	AccountLockedEmail:         "Tiny Build Server: Your account has been locked",
	ConfirmRegistrationEmail:   "Tiny Build Server: Please confirm your registration",
	RequestNewPasswordEmail:    "Tiny Build Server: Instructions on how to reset your password",
	RegistrationConfirmedEmail: "Tiny Build Server: Your registration was successfully confirmed",
	ConfirmPasswordResetEmail:  "Tiny Build Server: Your password has been reset",
	DeploymentEmail:            "Tiny Build Server: New email deployment",
}
