// Package mailer contains mail operations
package mailer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	mail "github.com/xhit/go-simple-mail/v2"
)

// Mailer Object is responsible for handling mailing configuration and logic.
type Mailer struct {
	From     *string
	To       []string
	Password *string
	SMTPHost *string
	SMTPPort *int
}

// Send is a wrapper for mailing logic, this could potentially use an API instead of SMTP,
// depending on flexibility and performance requirements.
func (m *Mailer) Send(ctx *context.Context, subject, body string) {
	log := logr.FromContextOrDiscard(*ctx).WithValues("from", *m.From, "to", strings.Join(m.To, ","))

	if *m.SMTPHost == "" {
		log.V(1).Info("SMTP Host is not set, will not send email")
		return
	}

	go func() {
		// SMTP Server
		server := mail.NewSMTPClient()
		server.Host = *m.SMTPHost
		server.Port = *m.SMTPPort
		server.Username = *m.From
		server.Password = *m.Password
		server.Encryption = mail.EncryptionSTARTTLS
		server.Authentication = mail.AuthPlain
		server.KeepAlive = false
		server.ConnectTimeout = 10 * time.Second
		server.SendTimeout = 10 * time.Second

		smtpClient, err := server.Connect()
		if err != nil {
			log.Error(err, "failed to connect to SMTP")
			return
		}
		defer smtpClient.Close()

		// New email
		email := mail.NewMSG()
		email.SetFrom(fmt.Sprintf("From Blacklister <%s>", *m.From)).
			AddTo(m.To...).
			SetSubject(subject)
		email.SetBody(mail.TextPlain, body)

		if email.Error != nil {
			log.Error(email.Error, "email error")
			return
		}

		// Send
		log.V(1).Info("Sending email", "smtp_host", *m.SMTPHost, "smtp_port", *m.SMTPPort)
		err = email.Send(smtpClient)
		if err != nil {
			log.Error(err, "Failed to send email")
			return
		}
		log.V(1).Info("Email sent")
	}()
}
