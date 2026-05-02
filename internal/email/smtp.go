package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"github.com/evermeet/evermeet/internal/config"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

// Config holds envelope/from and SMTP settings (Host empty when using sendmail).
type Config struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

// Client sends transactional email via SMTP or local sendmail.
type Client struct {
	transport    string // "smtp" | "sendmail"
	sendmailPath string
	cfg          Config
}

// NewClient picks SMTP when smtp_host is set; otherwise local sendmail when a
// sendmail binary exists. For sendmail, From is [email] from, or if empty
// defaultMailHost (hostname from the public base URL, e.g. "meet.example.com")
// is used as evermeet@<host>.
func NewClient(ec config.EmailConfig, defaultMailHost string) (*Client, string) {
	if strings.TrimSpace(ec.SMTPHost) != "" {
		port := ec.SMTPPort
		if port <= 0 {
			port = 587
		}
		c := &Client{
			transport: "smtp",
			cfg: Config{
				Host: ec.SMTPHost,
				Port: port,
				User: ec.SMTPUser,
				Pass: ec.SMTPPass,
				From: ec.From,
			},
		}
		return c, fmt.Sprintf("using SMTP %s:%d", ec.SMTPHost, port)
	}
	path := findSendmail()
	if path == "" {
		return nil, ""
	}
	from := strings.TrimSpace(ec.From)
	if from == "" {
		host := strings.TrimSpace(defaultMailHost)
		if host == "" {
			return nil, ""
		}
		from = "evermeet@" + host
	}
	c := &Client{
		transport:    "sendmail",
		sendmailPath: path,
		cfg: Config{
			From: from,
		},
	}
	if strings.TrimSpace(ec.From) == "" {
		return c, fmt.Sprintf("using local sendmail at %s (default From %s)", path, from)
	}
	return c, fmt.Sprintf("using local sendmail at %s", path)
}

// Transport returns "smtp" or "sendmail".
func (c *Client) Transport() string {
	return c.transport
}

// SendmailPath returns the sendmail binary path when Transport is sendmail.
func (c *Client) SendmailPath() string {
	if c.transport != "sendmail" {
		return ""
	}
	return c.sendmailPath
}

// FromAddress returns the From header / envelope address in use.
func (c *Client) FromAddress() string {
	return c.cfg.From
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

// SendTest sends a minimal HTML message to verify SMTP settings.
func (c *Client) SendTest(to string) error {
	const subject = "Evermeet mail test"
	const body = "<p>This is a test message from your Evermeet instance mail settings.</p>"
	return c.send(to, subject, body)
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
	msg := strings.Join([]string{
		"From: " + c.cfg.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="utf-8"`,
		"",
		htmlBody,
	}, "\r\n")

	if c.transport == "sendmail" {
		return deliverSendmail(c.sendmailPath, c.cfg.From, []byte(msg))
	}

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	auth := smtp.PlainAuth("", c.cfg.User, c.cfg.Pass, c.cfg.Host)
	return smtp.SendMail(addr, auth, c.cfg.From, []string{to}, []byte(msg))
}
