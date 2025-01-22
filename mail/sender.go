package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(
		subject,
		content string,
		to,
		cc,
		bcc,
		attachFiles []string,
	) error
}

type GmailSender struct {
	name          string
	emailAddress  string
	emailPassword string
}

func NewGmailSender(name, emailAddress, emailPassword string) EmailSender {
	return &GmailSender{
		name,
		emailAddress,
		emailPassword,
	}
}

func (sender *GmailSender) SendEmail(
	subject,
	content string,
	to,
	cc,
	bcc,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.emailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.emailAddress, sender.emailPassword, smtpAuthAddress)
	return e.Send(smtpServerAddress, smtpAuth)
}
