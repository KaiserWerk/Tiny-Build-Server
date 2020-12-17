package helper

var EmailSubjects = map[EmailMessageType]string{
	AccountCreated: "Tiny Build Server: Your account was created",
	AccountLocked: "Tiny Build Server: Your account has been locked",
	ConfirmRegistration: "Tiny Build Server: Please confirm your registration",
	PasswordReset: "Tiny Build Server: Instructions on how to reset your password",
	RegistrationConfirmed: "Tiny Build Server: Your registration was successfully confirmed",
	ResetPassword: "Tiny Build Server: Your password has been reset",
}
