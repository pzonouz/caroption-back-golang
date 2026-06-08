package utils

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendMail(
	from string,
	addrs []string,
	subject string,
	ccsAddr string,
	ccsName string,
	htmlMessage string,
) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", addrs...)
	m.SetAddressHeader("Cc", ccsAddr, ccsName)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlMessage)

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	d := gomail.NewDialer(
		os.Getenv("SMTP_ADDR"),
		port,
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
