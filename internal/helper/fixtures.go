package helper

var EmailSubjects = map[EmailMessageType]string{
	AccountCreated: "Build Server: Your account was created",
	AccountLocked: "Build Server: Your account has been locked",
	ConfirmRegistration: "Build Server: Please confirm your registration",
	PasswordReset: "Build Server: Instructions on how to reset your password",
	RegistrationConfirmed: "Build Server: Your registration was successfully confirmed",
	ResetPassword: "Build Server: Your password has been reset",
}
