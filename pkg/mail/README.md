# 📧 Mail Package

Simple email sending system with template support and SMTP configuration.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Basic Usage](#basic-usage)
- [Template Support](#template-support)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/mail"
```

## ⚡ Quick Start

### Basic Email Sending

```go
package main

import (
    "go-starter/pkg/mail"
    "go-starter/config"
)

func main() {
    // Load email configuration
    cfg := &config.EmailConfig{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        From:     "your-email@gmail.com",
        FromName: "Your App Name",
    }

    // Create mailer
    mailer, err := mail.NewGomail(cfg)
    if err != nil {
        panic(err)
    }

    // Send simple email
    err = mailer.SendEmail(
        []string{"recipient@example.com"},
        "Hello from Go Starter!",
        "This is a test email from our Go application.",
        nil, // no attachments
    )

    if err != nil {
        panic(err)
    }
}
```

## ⚙️ Configuration

### Email Configuration Structure

```go
type EmailConfig struct {
    Host               string        // SMTP host (e.g., "smtp.gmail.com")
    Port               int           // SMTP port (e.g., 587)
    Username           string        // SMTP username
    Password           string        // SMTP password or app password
    From               string        // Sender email address
    FromName           string        // Sender display name
    TemplateDir        string        // Template directory path
    MaxRetries         int           // Maximum retry attempts
    RetryDelay         time.Duration // Delay between retries
    InsecureSkipVerify bool          // Skip TLS certificate verification
}
```

### Environment Variables

```env
# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=your-email@gmail.com
SMTP_FROM_NAME=Your App Name

# Template Configuration
EMAIL_TEMPLATE_DIR=./templates
EMAIL_MAX_RETRIES=3
EMAIL_RETRY_DELAY=1s
EMAIL_INSECURE_SKIP_VERIFY=false
```

## 💡 Basic Usage

### Send Plain Text Email

```go
func SendWelcomeEmail(mailer *mail.Mailer, userEmail, userName string) error {
    subject := "Welcome to Our Platform!"
    body := fmt.Sprintf("Hello %s,\n\nWelcome to our platform! We're excited to have you on board.\n\nBest regards,\nThe Team", userName)

    return mailer.SendEmail(
        []string{userEmail},
        subject,
        body,
        nil, // no attachments
    )
}
```

### Send HTML Email

```go
func SendHTMLEmail(mailer *mail.Mailer, userEmail, userName string) error {
    subject := "Welcome to Our Platform!"
    htmlBody := fmt.Sprintf(`
        <html>
        <body>
            <h2>Welcome %s!</h2>
            <p>We're excited to have you on board.</p>
            <p><a href="https://yourapp.com/dashboard">Visit your dashboard</a></p>
            <br>
            <p>Best regards,<br>The Team</p>
        </body>
        </html>
    `, userName)

    return mailer.SendHTMLEmail(
        []string{userEmail},
        subject,
        htmlBody,
        nil, // no attachments
    )
}
```

### Send Email with Attachments

```go
func SendEmailWithAttachment(mailer *mail.Mailer) error {
    attachments := []string{
        "/path/to/document.pdf",
        "/path/to/image.jpg",
    }

    return mailer.SendEmail(
        []string{"recipient@example.com"},
        "Documents Attached",
        "Please find the attached documents.",
        attachments,
    )
}
```

## 📄 Template Support

### Create Email Templates

Create template files in your template directory:

**templates/welcome.html:**

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Welcome Email</title>
  </head>
  <body>
    <h1>Welcome {{.Name}}!</h1>
    <p>Thank you for joining {{.AppName}}.</p>
    <p>Your account details:</p>
    <ul>
      <li>Email: {{.Email}}</li>
      <li>Registration Date: {{.RegisteredAt}}</li>
    </ul>
    <p><a href="{{.DashboardURL}}">Visit your dashboard</a></p>
  </body>
</html>
```

### Send Template-Based Email

```go
func SendWelcomeEmailFromTemplate(mailer *mail.Mailer, userEmail, userName string) error {
    templateData := map[string]interface{}{
        "Name":         userName,
        "Email":        userEmail,
        "AppName":      "Go Starter",
        "RegisteredAt": time.Now().Format("January 2, 2006"),
        "DashboardURL": "https://yourapp.com/dashboard",
    }

    return mailer.SendTemplate(
        []string{userEmail},
        "Welcome to Go Starter!",
        "welcome", // template name (without .html extension)
        templateData,
        nil, // no attachments
    )
}
```

## 🎯 Real-World Examples

### 1. **User Registration Email**

```go
func SendUserRegistrationEmail(mailer *mail.Mailer, user *User) error {
    subject := fmt.Sprintf("Welcome to %s, %s!", "Go Starter", user.Name)

    body := fmt.Sprintf(`
        Hi %s,

        Welcome to Go Starter! Your account has been successfully created.

        Account Details:
        - Email: %s
        - Registration Date: %s

        You can now log in to your account and start using our services.

        If you have any questions, feel free to contact our support team.

        Best regards,
        The Go Starter Team
    `, user.Name, user.Email, time.Now().Format("January 2, 2006"))

    return mailer.SendEmail(
        []string{user.Email},
        subject,
        body,
        nil,
    )
}
```

### 2. **Password Reset Email**

```go
func SendPasswordResetEmail(mailer *mail.Mailer, userEmail, resetToken string) error {
    resetURL := fmt.Sprintf("https://yourapp.com/reset-password?token=%s", resetToken)

    subject := "Password Reset Request"
    body := fmt.Sprintf(`
        Hello,

        We received a request to reset your password for your Go Starter account.

        Click the link below to reset your password:
        %s

        This link will expire in 1 hour for security reasons.

        If you didn't request this password reset, please ignore this email.

        Best regards,
        The Go Starter Team
    `, resetURL)

    return mailer.SendEmail(
        []string{userEmail},
        subject,
        body,
        nil,
    )
}
```

### 3. **Order Confirmation Email**

```go
func SendOrderConfirmationEmail(mailer *mail.Mailer, order *Order) error {
    subject := fmt.Sprintf("Order Confirmation #%s", order.ID)

    htmlBody := fmt.Sprintf(`
        <html>
        <body>
            <h2>Order Confirmation</h2>
            <p>Thank you for your order!</p>

            <h3>Order Details:</h3>
            <ul>
                <li>Order ID: %s</li>
                <li>Order Date: %s</li>
                <li>Total Amount: $%.2f</li>
                <li>Status: %s</li>
            </ul>

            <p>We'll notify you when your order ships.</p>

            <p>Thank you for choosing our service!</p>
        </body>
        </html>
    `, order.ID, order.CreatedAt.Format("January 2, 2006"), order.TotalAmount, order.Status)

    return mailer.SendHTMLEmail(
        []string{order.User.Email},
        subject,
        htmlBody,
        nil,
    )
}
```

### 4. **Newsletter Email**

```go
func SendNewsletterEmail(mailer *mail.Mailer, subscribers []string, content NewsletterContent) error {
    subject := content.Subject

    htmlBody := fmt.Sprintf(`
        <html>
        <body>
            <h1>%s</h1>
            <div>%s</div>

            <hr>
            <p><small>
                You're receiving this because you subscribed to our newsletter.
                <a href="https://yourapp.com/unsubscribe">Unsubscribe</a>
            </small></p>
        </body>
        </html>
    `, content.Title, content.HTMLContent)

    // Send to all subscribers (in production, consider batching)
    return mailer.SendHTMLEmail(subscribers, subject, htmlBody, nil)
}
```

## 🎯 Best Practices

### 1. **Error Handling**

```go
func SafeSendEmail(mailer *mail.Mailer, to []string, subject, body string) {
    if err := mailer.SendEmail(to, subject, body, nil); err != nil {
        // Log error but don't crash the application
        log.Printf("Failed to send email to %v: %v", to, err)

        // Optionally, queue for retry or store in database
        // queueEmailForRetry(to, subject, body)
    }
}
```

### 2. **Environment-Specific Configuration**

```go
func GetEmailConfig(env string) *config.EmailConfig {
    switch env {
    case "production":
        return &config.EmailConfig{
            Host:     os.Getenv("SMTP_HOST"),
            Port:     587,
            Username: os.Getenv("SMTP_USERNAME"),
            Password: os.Getenv("SMTP_PASSWORD"),
            From:     os.Getenv("SMTP_FROM"),
            FromName: "YourApp Production",
        }
    case "staging":
        return &config.EmailConfig{
            Host:     "smtp.mailtrap.io", // Testing service
            Port:     2525,
            Username: os.Getenv("MAILTRAP_USERNAME"),
            Password: os.Getenv("MAILTRAP_PASSWORD"),
            From:     "staging@yourapp.com",
            FromName: "YourApp Staging",
        }
    default: // development
        return &config.EmailConfig{
            Host:     "localhost",
            Port:     1025, // MailHog or similar
            From:     "dev@yourapp.com",
            FromName: "YourApp Development",
        }
    }
}
```

### 3. **Template Organization**

```
templates/
├── emails/
│   ├── welcome.html
│   ├── password-reset.html
│   ├── order-confirmation.html
│   └── newsletter.html
└── layouts/
    └── base.html
```

### 4. **Testing Email Functionality**

```go
func TestEmailSending(t *testing.T) {
    // Use test SMTP server or mock
    cfg := &config.EmailConfig{
        Host: "localhost",
        Port: 1025, // MailHog test server
        From: "test@example.com",
    }

    mailer, err := mail.NewGomail(cfg)
    assert.NoError(t, err)

    err = mailer.SendEmail(
        []string{"test@example.com"},
        "Test Subject",
        "Test Body",
        nil,
    )
    assert.NoError(t, err)
}
```

### 5. **Production Considerations**

```go
// In production, consider:
// 1. Rate limiting email sending
// 2. Queue-based email processing
// 3. Email delivery tracking
// 4. Bounce handling
// 5. Unsubscribe management

func SendEmailAsync(mailer *mail.Mailer, emailJob EmailJob) {
    go func() {
        if err := mailer.SendEmail(
            emailJob.Recipients,
            emailJob.Subject,
            emailJob.Body,
            emailJob.Attachments,
        ); err != nil {
            // Log error and possibly retry
            log.Printf("Email sending failed: %v", err)
        }
    }()
}
```

## 📧 Email Providers

### **Gmail Configuration**

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-gmail@gmail.com
SMTP_PASSWORD=your-app-password  # Generate in Google Account settings
```

### **SendGrid Configuration**

```env
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
```

### **Mailgun Configuration**

```env
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USERNAME=your-mailgun-username
SMTP_PASSWORD=your-mailgun-password
```

## 🚨 Common Issues

### **Gmail "Less Secure App" Error**

```bash
# Solution: Use App Passwords
# 1. Enable 2FA in Google Account
# 2. Generate App Password
# 3. Use App Password instead of regular password
```

### **Email Not Delivered**

```go
// Check these common issues:
// 1. SMTP credentials are correct
// 2. Firewall allows SMTP port
// 3. Email content isn't flagged as spam
// 4. From email is properly configured
```

## 🔗 Related Packages

- [`config`](../../config/) - Email configuration
- [`pkg/logger`](../logger/) - Error logging
- [`pkg/validator`](../validator/) - Email validation

## 📚 Additional Resources

- [Go Mail Documentation](https://pkg.go.dev/gopkg.in/gomail.v2)
- [SMTP Configuration Guide](https://support.google.com/mail/answer/7126229)
- [Email Best Practices](https://sendgrid.com/blog/email-best-practices/)
