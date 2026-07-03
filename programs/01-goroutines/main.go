// goroutines: three real welcome emails, sent over real SMTP, concurrently.
//
// Points at a local MailHog container (a real SMTP server for dev/test):
//   docker run -d --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog
// In production you'd point Host/Port at smtp.sendgrid.net:587 and auth
// with "apikey" / your SendGrid API key — net/smtp doesn't change at all.
package main

import (
	"fmt"
	"net/smtp"
)

func sendEmail(to string) error {
	msg := []byte("To: " + to + "\r\n" +
		"Subject: Welcome!\r\n\r\n" +
		"Thanks for signing up.\r\n")
	// a REAL SMTP conversation: connect, EHLO, MAIL FROM,
	// RCPT TO, DATA — over an actual TCP socket to port 1025
	return smtp.SendMail("localhost:1025", nil,
		"noreply@example.com", []string{to}, msg)
}

func main() {
	users := []string{
		"asha@example.com",
		"ravi@example.com",
		"meera@example.com",
	}

	done := make(chan error, len(users))
	for _, u := range users {
		go func() { // fires the real SMTP call concurrently
			done <- sendEmail(u)
		}()
	}

	for range users {
		if err := <-done; err != nil {
			fmt.Println("send failed:", err)
			continue
		}
		fmt.Println("✉️  delivered ✓")
	}
}
