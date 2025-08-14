package mail

import (
	"bytes"
	"crypto/tls"
	"go-starter/config"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer        *gomail.Dialer
	templateCache map[string]*template.Template
	cacheMutex    sync.RWMutex
	config        *config.EmailConfig
}

func NewGomail(cfg *config.EmailConfig) (*Mailer, error) {
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		ServerName:         cfg.Host,
	}

	return &Mailer{
		dialer:        d,
		templateCache: make(map[string]*template.Template),
		cacheMutex:    sync.RWMutex{},
		config:        cfg,
	}, nil
}

// SendEmail sends an email with retry logic and better error handling
func (m *Mailer) SendEmail(to []string, subject string, body string, attachments []string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	if m.config.From == "" {
		return fmt.Errorf("MAILER_FROM environment variable is not set")
	}

	// Validate recipients
	if err := m.validateRecipients(to); err != nil {
		return fmt.Errorf("invalid recipients: %v", err)
	}

	message := m.createMessage(to, subject, body, attachments)
	if message == nil {
		return fmt.Errorf("failed to create email message")
	}

	// Send with retry logic
	return m.sendWithRetry(message)
}

// createMessage creates a gomail message with proper configuration
func (m *Mailer) createMessage(to []string, subject string, body string, attachments []string) *gomail.Message {
	message := gomail.NewMessage()

	// Set headers
	message.SetHeader("From", message.FormatAddress(m.config.From, m.config.FromName))
	message.SetHeader("To", to...)
	message.SetHeader("Subject", subject)

	// Set message ID for tracking
	message.SetHeader("Message-ID", fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), m.config.Host))

	// Set body with proper charset
	message.SetBody("text/html; charset=UTF-8", body)

	// Add attachments with validation
	for _, attachment := range attachments {
		if err := m.validateAttachment(attachment); err != nil {
			// Log warning but continue with other attachments
			continue
		}
		message.Attach(attachment)
	}

	return message
}

// validateRecipients validates email addresses
func (m *Mailer) validateRecipients(recipients []string) error {
	for _, email := range recipients {
		if email == "" {
			return fmt.Errorf("empty email address")
		}
		// Basic email validation (you might want to use a more robust validator)
		if len(email) < 5 || !contains(email, "@") || !contains(email, ".") {
			return fmt.Errorf("invalid email format: %s", email)
		}
	}
	return nil
}

// validateAttachment checks if attachment exists and is readable
func (m *Mailer) validateAttachment(path string) error {
	if path == "" {
		return fmt.Errorf("empty attachment path")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("attachment not found: %s", path)
	}

	if info.IsDir() {
		return fmt.Errorf("attachment is a directory: %s", path)
	}

	// Check file size (max 25MB)
	const maxSize = 25 * 1024 * 1024
	if info.Size() > maxSize {
		return fmt.Errorf("attachment too large: %s (%d bytes)", path, info.Size())
	}

	return nil
}

// sendWithRetry implements retry logic for sending emails
func (m *Mailer) sendWithRetry(message *gomail.Message) error {
	var lastErr error

	for attempt := 0; attempt <= m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(m.config.RetryDelay * time.Duration(attempt))
		}

		if err := m.dialer.DialAndSend(message); err != nil {
			lastErr = err
			continue
		}

		return nil // Success
	}

	return fmt.Errorf("failed to send email after %d attempts: %v", m.config.MaxRetries+1, lastErr)
}

// TestConnection tests the SMTP connection
func (m *Mailer) TestConnection() error {
	sender, err := m.dialer.Dial()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer sender.Close()

	return nil
}

// SendEmailWithTemplate sends an email using a template with caching
func (m *Mailer) SendEmailWithTemplate(to []string, subject string, templateName string, data interface{}, attachments []string) error {
	// Get template from cache or load it
	tmpl, err := m.getTemplate(templateName)
	if err != nil {
		return fmt.Errorf("failed to load template: %v", err)
	}

	// Execute template
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return m.SendEmail(to, subject, buffer.String(), attachments)
}

// getTemplate retrieves template from cache or loads it
func (m *Mailer) getTemplate(templateName string) (*template.Template, error) {
	// Check cache first (read lock)
	m.cacheMutex.RLock()
	if tmpl, exists := m.templateCache[templateName]; exists {
		m.cacheMutex.RUnlock()
		return tmpl, nil
	}
	m.cacheMutex.RUnlock()

	// Load template (write lock)
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Double-check in case another goroutine loaded it
	if tmpl, exists := m.templateCache[templateName]; exists {
		return tmpl, nil
	}

	// Validate template path
	if !filepath.IsAbs(templateName) {
		// Make relative paths relative to working directory
		wd, _ := os.Getwd()
		templateName = filepath.Join(wd, m.config.TemplateDir, templateName+".html")
	}

	// Load and parse template
	tmpl, err := template.ParseFiles(templateName)
	if err != nil {
		return nil, err
	}

	// Cache the template
	m.templateCache[templateName] = tmpl

	return tmpl, nil
}

// ClearTemplateCache clears the template cache (useful for development)
func (m *Mailer) ClearTemplateCache() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.templateCache = make(map[string]*template.Template)
}

// SendBulkEmail sends emails to multiple recipients efficiently
func (m *Mailer) SendBulkEmail(recipients []string, subject string, body string, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	// Send in batches to avoid overwhelming the SMTP server
	for i := 0; i < len(recipients); i += batchSize {
		end := i + batchSize
		if end > len(recipients) {
			end = len(recipients)
		}

		batch := recipients[i:end]
		if err := m.SendEmail(batch, subject, body, []string{}); err != nil {
			return fmt.Errorf("failed to send batch %d-%d: %v", i, end-1, err)
		}

		// Small delay between batches to be respectful to the SMTP server
		if end < len(recipients) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
