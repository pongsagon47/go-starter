# 🎯 Validator Package

Simple input validation and XSS sanitization wrapper around go-playground/validator with formatted error messages and automatic security protection.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [XSS Sanitization](#xss-sanitization)
- [Validation Tags](#validation-tags)
- [Sanitization Tags](#sanitization-tags)
- [Custom Error Messages](#custom-error-messages)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in flex-service
import "flex-service/pkg/validator"

# Required dependency
go get github.com/microcosm-cc/bluemonday
```

## ⚡ Quick Start

### Basic Struct Validation with XSS Protection

```go
package main

import (
    "fmt"
    "flex-service/pkg/validator"
)

type User struct {
    Name     string `json:"name" validate:"required,min=2,max=50" sanitize:"strict"`
    Email    string `json:"email" validate:"required,email" sanitize:"strict"`
    Bio      string `json:"bio,omitempty" validate:"max=1000" sanitize:"rich"`
    Age      int    `json:"age" validate:"required,gte=18,lte=100"`
    Password string `json:"password" validate:"required,min=8"`
}

func main() {
    user := User{
        Name:     "<script>alert('xss')</script>John",  // Will be sanitized
        Email:    "john@example.com",
        Bio:      "<p>Hello <strong>world</strong></p>", // HTML preserved
        Age:      25,
        Password: "securepass123",
    }

    // Single function call: sanitizes then validates
    errors := validator.ValidateStruct(&user)
    if errors != nil {
        for field, message := range errors {
            fmt.Printf("%s: %s\n", field, message)
        }
    }

    // After sanitization:
    // user.Name = "John"  (script removed)
    // user.Bio = "<p>Hello <strong>world</strong></p>"  (safe HTML preserved)
}
```

## 🛡️ XSS Sanitization

The validator automatically sanitizes input based on `sanitize` tags **before** validation to prevent XSS attacks.

### Sanitization Modes

| Mode     | Description                           | Use Case                    |
| -------- | ------------------------------------- | --------------------------- |
| `strict` | Remove all HTML (default)             | Names, emails, titles       |
| `ugc`    | Allow basic HTML tags                 | Comments, descriptions      |
| `rich`   | Allow full HTML (tables, images, etc) | Articles, rich text content |
| `none`   | No sanitization                       | Code, CSS, system data      |

### Default Behavior

```go
type Example struct {
    Name    string // Default: sanitize="strict" (secure by default)
    Bio     string `sanitize:"rich"`  // Explicitly allow rich HTML
    Comment string `sanitize:"ugc"`   // Allow basic HTML
    Code    string `sanitize:"none"`  // Don't sanitize code
}
```

## 🏷️ Validation Tags

### **Basic Tags**

| Tag        | Description           | Example               |
| ---------- | --------------------- | --------------------- |
| `required` | Field must be present | `validate:"required"` |
| `email`    | Must be valid email   | `validate:"email"`    |
| `min`      | Minimum length/value  | `validate:"min=3"`    |
| `max`      | Maximum length/value  | `validate:"max=50"`   |
| `gte`      | Greater than or equal | `validate:"gte=18"`   |
| `lte`      | Less than or equal    | `validate:"lte=100"`  |
| `len`      | Exact length          | `validate:"len=10"`   |

### **String Validation**

```go
type StringValidation struct {
    Username string `validate:"required,min=3,max=20,alphanum" sanitize:"strict"`
    URL      string `validate:"omitempty,url" sanitize:"strict"`
    UUID     string `validate:"required,uuid" sanitize:"strict"`
    Phone    string `validate:"omitempty,e164" sanitize:"strict"`
}
```

## 🔒 Sanitization Tags

### **Strict Mode (Default)**

Removes all HTML and escapes special characters. Perfect for user input that should never contain markup.

```go
type UserProfile struct {
    Name     string `sanitize:"strict"`  // <script>alert('xss')</script>John → John
    Email    string `sanitize:"strict"`  // test<b>@example.com → test@example.com
    Username string `sanitize:"strict"`  // user<img src=x> → user
    Phone    string                      // Default: strict mode
}
```

### **UGC Mode (User Generated Content)**

Allows basic HTML tags like paragraphs, links, and text formatting.

```go
type BlogComment struct {
    Content string `sanitize:"ugc"` // Allows: <p>, <strong>, <em>, <a>, <ul>, <li>
    Author  string `sanitize:"strict"`
}

// Input:  "<p>Great post! <script>alert('xss')</script> <strong>Thanks!</strong></p>"
// Output: "<p>Great post!  <strong>Thanks!</strong></p>"
```

### **Rich Mode**

Allows comprehensive HTML including headings, images, tables, and complex formatting.

```go
type Article struct {
    Title   string `sanitize:"strict"`
    Content string `sanitize:"rich"`  // Allows: <h1>-<h6>, <img>, <table>, <div>, etc.
    Summary string `sanitize:"ugc"`   // Basic HTML only
}

// Rich mode preserves:
// <h1>Heading</h1><p>Content with <img src="safe.jpg"></p><table>...</table>
```

### **None Mode**

Skips sanitization entirely. Use for system data, code snippets, or pre-validated content.

```go
type AdminSettings struct {
    CustomCSS  string `sanitize:"none"`  // CSS code preserved
    CustomJS   string `sanitize:"none"`  // JavaScript preserved
    Template   string `sanitize:"none"`  // Template syntax preserved
    SQLQuery   string `sanitize:"none"`  // SQL preserved
}
```

## 💡 Examples

### **1. User Registration with XSS Protection**

```go
type RegisterRequest struct {
    Name            string `json:"name" validate:"required,min=2,max=50" sanitize:"strict"`
    Email           string `json:"email" validate:"required,email" sanitize:"strict"`
    Password        string `json:"password" validate:"required,min=8,max=128"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
    Bio             string `json:"bio,omitempty" validate:"max=500" sanitize:"ugc"`
    Age             int    `json:"age" validate:"required,gte=18"`
    Terms           bool   `json:"terms" validate:"required,eq=true"`
}

func RegisterHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    // Auto-sanitizes then validates
    if errors := validator.ValidateStruct(&req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // req.Name and req.Email are now XSS-safe
    // req.Bio allows safe HTML formatting
}
```

### **2. Blog Article Creation**

```go
type CreateArticleRequest struct {
    Title       string   `json:"title" validate:"required,min=5,max=200" sanitize:"strict"`
    Content     string   `json:"content" validate:"required,min=50" sanitize:"rich"`
    Summary     string   `json:"summary,omitempty" validate:"max=300" sanitize:"ugc"`
    Tags        []string `json:"tags" validate:"omitempty,max=10,dive,min=1,max=30" sanitize:"strict"`
    MetaDesc    string   `json:"meta_description,omitempty" validate:"max=160" sanitize:"strict"`
    CustomCSS   string   `json:"custom_css,omitempty" sanitize:"none"`  // Admin only
    CategoryID  string   `json:"category_id" validate:"required,uuid"`
}

func CreateArticleHandler(c *gin.Context) {
    var req CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    if errors := validator.ValidateStruct(&req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // Safe to use:
    // req.Title - XSS-safe text
    // req.Content - Safe rich HTML
    // req.Summary - Safe basic HTML
    // req.CustomCSS - Untouched (admin feature)
}
```

### **3. Comment System with Multi-level Sanitization**

```go
type Comment struct {
    Content    string `json:"content" validate:"required,min=1,max=1000" sanitize:"ugc"`
    AuthorName string `json:"author_name" validate:"required,min=1,max=50" sanitize:"strict"`
    Email      string `json:"email" validate:"required,email" sanitize:"strict"`
    Website    string `json:"website,omitempty" validate:"omitempty,url" sanitize:"strict"`
    UserAgent  string `json:"user_agent,omitempty" sanitize:"strict"`
    Metadata   string `json:"metadata,omitempty" sanitize:"none"`  // System use
}

// Input example:
comment := Comment{
    Content:    "<p>Great article! <script>alert('xss')</script> <strong>Well done!</strong></p>",
    AuthorName: "<b>John</b> Doe<script>alert('hack')</script>",
    Email:      "john@example.com",
}

// After ValidateStruct(&comment):
// comment.Content = "<p>Great article!  <strong>Well done!</strong></p>"
// comment.AuthorName = "John Doe"
// comment.Email = "john@example.com"
```

### **4. Admin Configuration with Selective Sanitization**

```go
type SiteConfig struct {
    SiteName     string `json:"site_name" validate:"required,min=1,max=100" sanitize:"strict"`
    Description  string `json:"description" validate:"max=500" sanitize:"ugc"`
    CustomHeader string `json:"custom_header,omitempty" sanitize:"none"`     // Raw HTML
    CustomCSS    string `json:"custom_css,omitempty" sanitize:"none"`       // Raw CSS
    CustomJS     string `json:"custom_js,omitempty" sanitize:"none"`        // Raw JS
    FooterText   string `json:"footer_text,omitempty" sanitize:"rich"`      // Rich HTML
    ContactInfo  string `json:"contact_info" validate:"required" sanitize:"ugc"`
}

func UpdateSiteConfigHandler(c *gin.Context) {
    // Only allow admins to access this endpoint
    if !isAdmin(c) {
        response.Forbidden(c, "Admin access required")
        return
    }

    var config SiteConfig
    if err := c.ShouldBindJSON(&config); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    if errors := validator.ValidateStruct(&config); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // Admins can inject custom code via CustomHeader, CustomCSS, CustomJS
    // Other fields are sanitized appropriately
}
```

### **5. File Upload Metadata with Security**

```go
type FileUploadRequest struct {
    FileName    string `json:"file_name" validate:"required,min=1,max=255" sanitize:"strict"`
    Description string `json:"description,omitempty" validate:"max=1000" sanitize:"ugc"`
    AltText     string `json:"alt_text,omitempty" validate:"max=255" sanitize:"strict"`
    Title       string `json:"title,omitempty" validate:"max=255" sanitize:"strict"`
    Caption     string `json:"caption,omitempty" sanitize:"rich"`  // Allow formatting
    Metadata    string `json:"metadata,omitempty" sanitize:"none"` // JSON metadata
}

func UploadFileHandler(c *gin.Context) {
    var req FileUploadRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    if errors := validator.ValidateStruct(&req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // req.FileName is XSS-safe
    // req.Description allows basic HTML
    // req.Caption allows rich formatting
    // req.Metadata preserves JSON structure
}
```

## 🛠️ Advanced Security Features

### **Manual Sanitization**

```go
import "flex-service/pkg/validator"

// Direct sanitization without validation
text := "<script>alert('xss')</script><p>Hello <strong>world</strong></p>"

strict := validator.GetSanitizer().SanitizeWithMode(text, "strict")
// Result: "Hello world"

ugc := validator.GetSanitizer().SanitizeWithMode(text, "ugc")
// Result: "<p>Hello <strong>world</strong></p>"

rich := validator.GetSanitizer().SanitizeWithMode(text, "rich")
// Result: "<p>Hello <strong>world</strong></p>"

none := validator.GetSanitizer().SanitizeWithMode(text, "none")
// Result: "<script>alert('xss')</script><p>Hello <strong>world</strong></p>"
```

### **Conditional Sanitization**

```go
type DynamicContent struct {
    Content  string `json:"content"`
    UserRole string `json:"-"` // Set by middleware
}

func (d *DynamicContent) CustomSanitize() {
    sanitizer := validator.GetSanitizer()

    switch d.UserRole {
    case "admin":
        // Admins can use rich HTML
        d.Content = sanitizer.SanitizeWithMode(d.Content, "rich")
    case "editor":
        // Editors can use basic HTML
        d.Content = sanitizer.SanitizeWithMode(d.Content, "ugc")
    default:
        // Regular users get strict sanitization
        d.Content = sanitizer.SanitizeWithMode(d.Content, "strict")
    }
}
```

## 🎯 Security Best Practices

### **1. Secure by Default**

```go
// ✅ Good: Explicit sanitization tags
type User struct {
    Name string `sanitize:"strict"`  // Safe for display anywhere
    Bio  string `sanitize:"rich"`    // Intentionally allows HTML
}

// ✅ Also Good: Default is strict (secure)
type User struct {
    Name string                      // Defaults to strict mode
    Bio  string `sanitize:"rich"`    // Explicitly allow HTML
}
```

### **2. Progressive Permission Model**

```go
// Regular users - strict sanitization
type UserComment struct {
    Content string `sanitize:"ugc"`  // Basic HTML only
}

// Trusted users - more HTML allowed
type EditorPost struct {
    Content string `sanitize:"rich"` // Full HTML support
}

// Administrators - minimal restrictions
type AdminConfig struct {
    CustomCode string `sanitize:"none"` // Raw code allowed
}
```

### **3. Context-Aware Sanitization**

```go
type BlogPost struct {
    Title       string `sanitize:"strict"`  // Never allow HTML in titles
    Excerpt     string `sanitize:"ugc"`     // Basic formatting in excerpts
    Content     string `sanitize:"rich"`    // Full editor in content
    MetaDesc    string `sanitize:"strict"`  // SEO fields stay clean
    CustomCSS   string `sanitize:"none"`    // Admin-only features
}
```

### **4. Validation Order**

The validator automatically follows the secure pattern:

1. **Sanitize** input based on `sanitize` tags
2. **Validate** sanitized input based on `validate` tags
3. **Return** any validation errors

```go
func SecureHandler(c *gin.Context) {
    var req MyRequest
    c.ShouldBindJSON(&req)

    // This single call:
    // 1. Sanitizes req based on sanitize tags
    // 2. Validates req based on validate tags
    // 3. Returns formatted errors if any
    if errors := validator.ValidateStruct(&req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // req is now safe to use - sanitized and validated
}
```

## 🔍 XSS Protection Examples

### **Input vs Output Examples**

```go
// Example malicious inputs and their sanitized outputs

type SecurityTest struct {
    Strict string `sanitize:"strict"`
    UGC    string `sanitize:"ugc"`
    Rich   string `sanitize:"rich"`
    None   string `sanitize:"none"`
}

test := SecurityTest{
    Strict: "<script>alert('xss')</script>Hello <b>World</b>",
    UGC:    "<p>Hello <script>alert('xss')</script> <strong>World</strong></p>",
    Rich:   "<h1>Title</h1><script>alert('xss')</script><img src='image.jpg'>",
    None:   "<script>alert('xss')</script>Raw content",
}

validator.ValidateStruct(&test)

// Results after sanitization:
// test.Strict = "Hello World"
// test.UGC = "<p>Hello  <strong>World</strong></p>"
// test.Rich = "<h1>Title</h1><img src='image.jpg'>"
// test.None = "<script>alert('xss')</script>Raw content"
```

## 📚 Sanitization Tag Reference

### **Allowed HTML by Mode**

#### Strict Mode

- **Allowed**: None (all HTML removed)
- **Use for**: Names, emails, titles, metadata

#### UGC Mode

- **Allowed**: `<p>`, `<br>`, `<strong>`, `<em>`, `<b>`, `<i>`, `<u>`, `<a>`, `<ul>`, `<ol>`, `<li>`, `<blockquote>`
- **Use for**: Comments, descriptions, basic content

#### Rich Mode

- **Allowed**: All UGC tags plus `<h1>-<h6>`, `<img>`, `<table>`, `<div>`, `<span>`, `<pre>`, `<code>`, style attributes
- **Use for**: Articles, rich text content, full editor content

#### None Mode

- **Allowed**: Everything (no sanitization)
- **Use for**: Admin content, code snippets, system data

## 🧪 Testing Security

```go
func TestXSSSanitization(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        mode     string
        expected string
    }{
        {
            name:     "script removal in strict mode",
            input:    "<script>alert('xss')</script>Hello",
            mode:     "strict",
            expected: "Hello",
        },
        {
            name:     "safe HTML preservation in UGC mode",
            input:    "<p>Hello <strong>world</strong></p>",
            mode:     "ugc",
            expected: "<p>Hello <strong>world</strong></p>",
        },
        {
            name:     "malicious script removal in UGC mode",
            input:    "<p>Hello <script>alert('xss')</script> world</p>",
            mode:     "ugc",
            expected: "<p>Hello  world</p>",
        },
    }

    sanitizer := validator.GetSanitizer()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := sanitizer.SanitizeWithMode(tt.input, tt.mode)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## 🎯 Best Practices

### **1. Choose the Right Sanitization Mode**

```go
// ✅ Good: Appropriate modes for each field
type BlogPost struct {
    Title    string `sanitize:"strict"`  // Titles should never have HTML
    Content  string `sanitize:"rich"`    // Content needs rich formatting
    Summary  string `sanitize:"ugc"`     // Summary needs basic formatting
    Slug     string `sanitize:"strict"`  // Slugs must be clean
}

// ❌ Bad: Using rich mode everywhere
type BlogPost struct {
    Title   string `sanitize:"rich"`    // Dangerous for titles
    Slug    string `sanitize:"rich"`    // Dangerous for URLs
    Content string `sanitize:"rich"`    // OK for content
}
```

### **2. Validate Before Database Operations**

```go
// ✅ Good: Sanitize and validate at API boundary
func CreatePostHandler(c *gin.Context) {
    var post BlogPost
    c.ShouldBindJSON(&post)

    // Sanitize + validate in one call
    if errors := validator.ValidateStruct(&post); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // post is now safe to store in database
    database.Create(&post)
}
```

### **3. Document Security Requirements**

```go
// Document why each field uses its sanitization mode
type UserProfile struct {
    // Personal info - must be clean for display in all contexts
    Name     string `json:"name" sanitize:"strict" validate:"required,min=1,max=50"`
    Email    string `json:"email" sanitize:"strict" validate:"required,email"`

    // User content - allow basic formatting for better UX
    Bio      string `json:"bio" sanitize:"ugc" validate:"max=1000"`

    // Rich content - for users who need advanced formatting
    Signature string `json:"signature" sanitize:"rich" validate:"max=500"`

    // System fields - never sanitize system-generated content
    Metadata  string `json:"metadata" sanitize:"none"`
}
```

### **4. Handle Edge Cases**

```go
// Handle special characters and edge cases
type SpecialContent struct {
    MathFormula  string `sanitize:"none"`     // Preserve < > in formulas
    CodeSnippet  string `sanitize:"none"`     // Preserve HTML in code examples
    Email        string `sanitize:"strict"`   // Clean email addresses
    DisplayName  string `sanitize:"strict"`   // Clean display names
}
```

## 🔗 Related Packages

- [`pkg/response`](../response/) - API response formatting
- [`pkg/errors`](../errors/) - Error handling
- [`internal/middleware`](../../internal/middleware/) - Request validation middleware

## 📚 Additional Resources

- [Go Playground Validator Documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Bluemonday HTML Sanitizer](https://github.com/microcosm-cc/bluemonday)
- [OWASP XSS Prevention](https://owasp.org/www-project-cheat-sheets/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [HTML Sanitization Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
