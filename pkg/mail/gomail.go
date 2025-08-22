package mail

import (
	"bytes"
	"crypto/tls"
	"flex-service/config"
	"fmt"
	"html/template"
	"path/filepath"

	"gopkg.in/gomail.v2"
)

// Mailer provides basic email functionality without complex caching
type Mailer struct {
	dialer *gomail.Dialer
	config *config.EmailConfig
}

// NewGomail creates a new mailer instance
func NewGomail(cfg *config.EmailConfig) (*Mailer, error) {
	if cfg.From == "" {
		return nil, fmt.Errorf("email from address is required")
	}

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		ServerName:         cfg.Host,
	}

	return &Mailer{
		dialer: d,
		config: cfg,
	}, nil
}

// SendEmail sends a plain text email
func (m *Mailer) SendEmail(to []string, subject, body string, attachments []string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.config.From)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	// Add attachments if any
	for _, attachment := range attachments {
		msg.Attach(attachment)
	}

	return m.dialer.DialAndSend(msg)
}

// SendHTMLEmail sends an HTML email
func (m *Mailer) SendHTMLEmail(to []string, subject, htmlBody string, attachments []string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.config.From)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", htmlBody)

	// Add attachments if any
	for _, attachment := range attachments {
		msg.Attach(attachment)
	}

	return m.dialer.DialAndSend(msg)
}

// SendTemplate sends an email using a template file (simplified - no caching)
func (m *Mailer) SendTemplate(to []string, subject, templateName string, data interface{}, attachments []string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	if m.config.TemplateDir == "" {
		return fmt.Errorf("template directory not configured")
	}

	// Load and parse template (fresh each time - simpler for starter)
	templatePath := filepath.Join(m.config.TemplateDir, templateName+".html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	// Send HTML email
	return m.SendHTMLEmail(to, subject, buf.String(), attachments)
}

// TestConnection tests the SMTP connection
func (m *Mailer) TestConnection() error {
	// Create a test message without sending
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.config.From)
	msg.SetHeader("To", m.config.From) // Send to self for test
	msg.SetHeader("Subject", "Test Connection")
	msg.SetBody("text/plain", "This is a test connection email")

	// Try to dial (connect but don't send)
	sender, err := m.dialer.Dial()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer sender.Close()

	return nil
}
