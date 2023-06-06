package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(token string, email string) string {
	// Sender data.
	from := os.Getenv("email_account")
	password := os.Getenv("email_password")
	if from == "" || password == "" {
		return "Unable to send reset email: sender credentials not provided"
	}

	// Receiver email address.
	to := []string{
		email,
	}

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message.
	var message []byte
	reset_url := os.Getenv("reset_url")
	if reset_url == "" {
		message = []byte("Hello! Use the reset token below to change your OGrEE password:\r\n" +
			token + "\r\n")
	} else {
		message = []byte("Hello! Use the link below to reset your OGrEE password.\r\n" +
			reset_url + token + "\r\n")
	}

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return err.Error()
	}
	fmt.Println("Email Sent Successfully!")
	return ""
}
