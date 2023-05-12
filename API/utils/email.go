package utils

import (
	"fmt"
	"net/smtp"
)

func SendEmail(token string, email string) {
	// Sender data.
	from := ""
	password := ""

	// Receiver email address.
	to := []string{
		email,
	}

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message.
	message := []byte("Hello! Use the link below to reset your OGrEE password.\r\n" +
		"http://localhost:52836/#/reset?token=" +
		token + "\r\n")

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}
