package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

// Config holds SMTP connection settings.
type Config struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

// Client sends transactional email via SMTP.
type Client struct {
	cfg Config
}

func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

// SendMagicLink emails a sign-in link to the given address.
func (c *Client) SendMagicLink(to, url string) error {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "magic_link.html", map[string]string{"URL": url}); err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	subject := "Sign in to Evermeet"
	return c.send(to, subject, buf.String())
}

// SendRSVPConfirmation emails a confirmation of successful RSVP.
func (c *Client) SendRSVPConfirmation(to, eventTitle, date, location string) error {
	var buf bytes.Buffer
	data := map[string]string{
		"EventTitle": eventTitle,
		"Date":       date,
		"Location":   location,
	}
	if err := templates.ExecuteTemplate(&buf, "rsvp_confirmation.html", data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	subject := "RSVP Confirmed: " + eventTitle
	return c.send(to, subject, buf.String())
}

// SendInvitation emails an invitation to an event.
func (c *Client) SendInvitation(to, eventTitle, organizerName, url string) error {
	var buf bytes.Buffer
	data := map[string]string{
		"EventTitle":    eventTitle,
		"OrganizerName": organizerName,
		"URL":           url,
	}
	if err := templates.ExecuteTemplate(&buf, "invitation.html", data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	subject := "Invitation: " + eventTitle
	return c.send(to, subject, buf.String())
}

func (c *Client) send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	auth := smtp.PlainAuth("", c.cfg.User, c.cfg.Pass, c.cfg.Host)

	msg := strings.Join([]string{
		"From: " + c.cfg.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="utf-8"`,
		"",
		htmlBody,
	}, "\r\n")

	return smtp.SendMail(addr, auth, c.cfg.From, []string{to}, []byte(msg))
}
