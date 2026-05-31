// Package smtp holds the real SMTP adapter for the mail.Sender port.
package smtp

import (
	"context"
	"fmt"
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

func New(host, port, user, pass, from string) *Sender {
	return &Sender{host: host, port: port, user: user, pass: pass, from: from}
}

func (s *Sender) Send(_ context.Context, to, subject, body string) error {
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
