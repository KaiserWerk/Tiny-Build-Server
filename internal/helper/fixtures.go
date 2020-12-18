package helper

var EmailSubjects = map[EmailMessageType]string{
	AccountCreatedEmail:        "Tiny Build Server: Your account was created",
	AccountLockedEmail:         "Tiny Build Server: Your account has been locked",
	ConfirmRegistrationEmail:   "Tiny Build Server: Please confirm your registration",
	RequestNewPasswordEmail:    "Tiny Build Server: Instructions on how to reset your password",
	RegistrationConfirmedEmail: "Tiny Build Server: Your registration was successfully confirmed",
	ConfirmPasswordResetEmail:  "Tiny Build Server: Your password has been reset",
}
