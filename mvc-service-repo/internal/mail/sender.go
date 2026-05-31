// Package mail holds the SMTP sender plus the email template helpers used by
// the service layer. If SMTP_HOST is empty Send becomes a no-op that just
// logs the would-be send via slog, so the example runs without real SMTP.
package mail

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

type Sender struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSender(host, port, user, pass, from string) *Sender {
	return &Sender{host: host, port: port, user: user, pass: pass, from: from}
}

func (s *Sender) Send(to, subject, body string) error {
	if s.host == "" {
		slog.Debug("mail: smtp not configured, skipping",
			"to", to, "subject", subject, "body_len", len(body))
		return nil
	}
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}
	msg := buildMessage(s.from, to, subject, body)
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=utf-8\r\n")
	fmt.Fprintf(&b, "\r\n")
	b.WriteString(body)
	return b.String()
}
