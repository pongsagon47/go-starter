# 🎯 Validator Package

Simple input validation wrapper around go-playground/validator with formatted error messages.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Validation Tags](#validation-tags)
- [Custom Error Messages](#custom-error-messages)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/validator"
```

## ⚡ Quick Start

### Basic Struct Validation

```go
package main

import (
    "fmt"
    "go-starter/pkg/validator"
)

type User struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18,lte=100"`
    Password string `json:"password" validate:"required,min=8"`
}

func main() {
    user := User{
        Name:     "",           // Invalid: required
        Email:    "invalid",    // Invalid: not email format
        Age:      15,           // Invalid: less than 18
        Password: "123",        // Invalid: less than 8 characters
    }

    errors := validator.ValidateStruct(user)
    if errors != nil {
        for field, message := range errors {
            fmt.Printf("%s: %s\n", field, message)
        }
    }

    // Output:
    // name: name is required
    // email: email must be a valid email
    // age: age must be greater than or equal to 18
    // password: password must be at least 8 characters
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
    Username string `validate:"required,min=3,max=20,alphanum"`
    URL      string `validate:"omitempty,url"`
    UUID     string `validate:"required,uuid"`
    Phone    string `validate:"omitempty,e164"` // International format
}
```

### **Number Validation**

```go
type NumberValidation struct {
    Age     int     `validate:"required,gte=0,lte=150"`
    Price   float64 `validate:"required,gt=0"`
    Rating  int     `validate:"omitempty,gte=1,lte=5"`
    Percent float64 `validate:"gte=0,lte=100"`
}
```

### **Array/Slice Validation**

```go
type ArrayValidation struct {
    Tags      []string `validate:"required,min=1,max=10,dive,min=1,max=20"`
    Numbers   []int    `validate:"omitempty,dive,gte=1,lte=100"`
    Emails    []string `validate:"required,dive,email"`
}
```

## 💡 Examples

### **1. User Registration Validation**

```go
type RegisterRequest struct {
    Name            string `json:"name" validate:"required,min=2,max=50"`
    Email           string `json:"email" validate:"required,email"`
    Password        string `json:"password" validate:"required,min=8,max=128"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
    Age             int    `json:"age" validate:"required,gte=18"`
    Terms           bool   `json:"terms" validate:"required,eq=true"`
}

func RegisterHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    // Validate request
    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // Process registration...
    response.Success(c, 201, "User registered successfully", gin.H{
        "user_id": userID,
    })
}
```

### **2. Product Creation Validation**

```go
type Product struct {
    Name        string   `json:"name" validate:"required,min=1,max=100"`
    Description string   `json:"description" validate:"omitempty,max=1000"`
    Price       float64  `json:"price" validate:"required,gt=0"`
    CategoryID  string   `json:"category_id" validate:"required,uuid"`
    Tags        []string `json:"tags" validate:"omitempty,max=10,dive,min=1,max=30"`
    Stock       int      `json:"stock" validate:"required,gte=0"`
    SKU         string   `json:"sku" validate:"required,min=3,max=50,alphanum"`
}

func CreateProductHandler(c *gin.Context) {
    var product Product
    if err := c.ShouldBindJSON(&product); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    if errors := validator.ValidateStruct(product); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // Create product...
}
```

### **3. Query Parameter Validation**

```go
type ListProductsQuery struct {
    Page     int    `form:"page" validate:"omitempty,gte=1"`
    Limit    int    `form:"limit" validate:"omitempty,gte=1,lte=100"`
    Category string `form:"category" validate:"omitempty,uuid"`
    Sort     string `form:"sort" validate:"omitempty,oneof=name price created_at"`
    Order    string `form:"order" validate:"omitempty,oneof=asc desc"`
}

func ListProductsHandler(c *gin.Context) {
    var query ListProductsQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        response.ValidationError(c, "Invalid query parameters", nil)
        return
    }

    // Set defaults
    if query.Page == 0 {
        query.Page = 1
    }
    if query.Limit == 0 {
        query.Limit = 10
    }
    if query.Sort == "" {
        query.Sort = "created_at"
    }
    if query.Order == "" {
        query.Order = "desc"
    }

    if errors := validator.ValidateStruct(query); errors != nil {
        response.ValidationError(c, "Invalid query parameters", errors)
        return
    }

    // Get products...
}
```

### **4. File Upload Validation**

```go
type FileUpload struct {
    Title       string `json:"title" validate:"required,min=1,max=100"`
    Description string `json:"description" validate:"omitempty,max=500"`
    FileType    string `json:"file_type" validate:"required,oneof=image document video"`
    MaxSize     int64  `json:"max_size" validate:"omitempty,lte=52428800"` // 50MB
}

func UploadFileHandler(c *gin.Context) {
    var req FileUpload
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request format", nil)
        return
    }

    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // Handle file upload...
}
```

### **5. Update Profile Validation**

```go
type UpdateProfileRequest struct {
    Name    *string `json:"name" validate:"omitempty,min=2,max=50"`
    Email   *string `json:"email" validate:"omitempty,email"`
    Phone   *string `json:"phone" validate:"omitempty,e164"`
    Bio     *string `json:"bio" validate:"omitempty,max=500"`
    Website *string `json:"website" validate:"omitempty,url"`
}

func UpdateProfileHandler(c *gin.Context) {
    var req UpdateProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // Update only provided fields...
    updateData := make(map[string]interface{})
    if req.Name != nil {
        updateData["name"] = *req.Name
    }
    if req.Email != nil {
        updateData["email"] = *req.Email
    }
    // ... update other fields
}
```

## 🎯 Custom Error Messages

The validator automatically formats error messages, but you can customize them:

### **Current Error Messages**

- `required` → "field is required"
- `email` → "field must be a valid email"
- `min` → "field must be at least X characters"
- `max` → "field must be at most X characters"
- `gte` → "field must be greater than or equal to X"
- `lte` → "field must be less than or equal to X"

### **Custom Validation Function**

```go
func ValidateWithCustomMessages(s interface{}) map[string]string {
    errors := validator.ValidateStruct(s)
    if errors == nil {
        return nil
    }

    // Customize specific error messages
    customErrors := make(map[string]string)
    for field, message := range errors {
        switch field {
        case "password":
            if strings.Contains(message, "at least") {
                customErrors[field] = "Password must be at least 8 characters long"
            } else {
                customErrors[field] = message
            }
        case "email":
            customErrors[field] = "Please provide a valid email address"
        default:
            customErrors[field] = message
        }
    }

    return customErrors
}
```

## 🛠️ Advanced Usage

### **Conditional Validation**

```go
type ConditionalValidation struct {
    Type     string  `json:"type" validate:"required,oneof=individual business"`
    Name     string  `json:"name" validate:"required"`
    TaxID    *string `json:"tax_id" validate:"required_if=Type business"`
    PersonID *string `json:"person_id" validate:"required_if=Type individual"`
}
```

### **Cross-Field Validation**

```go
type DateRange struct {
    StartDate time.Time `json:"start_date" validate:"required"`
    EndDate   time.Time `json:"end_date" validate:"required,gtefield=StartDate"`
}
```

### **Nested Struct Validation**

```go
type Address struct {
    Street  string `json:"street" validate:"required,min=5,max=100"`
    City    string `json:"city" validate:"required,min=2,max=50"`
    ZipCode string `json:"zip_code" validate:"required,min=5,max=10"`
    Country string `json:"country" validate:"required,len=2"` // ISO country code
}

type UserWithAddress struct {
    Name    string  `json:"name" validate:"required,min=2,max=50"`
    Email   string  `json:"email" validate:"required,email"`
    Address Address `json:"address" validate:"required"`
}
```

## 🎯 Best Practices

### **1. Validate Early and Often**

```go
// Validate input at API boundaries
func CreateUserHandler(c *gin.Context) {
    var req CreateUserRequest

    // 1. Bind and validate JSON structure
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    // 2. Validate business rules
    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // 3. Additional business logic validation
    if userExists(req.Email) {
        response.Error(c, 409, "USER_EXISTS", "Email already registered", nil)
        return
    }

    // Process request...
}
```

### **2. Use Consistent Validation Tags**

```go
// Good: Consistent validation patterns
type User struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
}

type Product struct {
    Name string `json:"name" validate:"required,min=2,max=50"`  // Same pattern
    SKU  string `json:"sku" validate:"required,min=3,max=20"`
}
```

### **3. Document Validation Rules**

```go
// Document validation requirements
type CreatePostRequest struct {
    Title   string   `json:"title" validate:"required,min=5,max=200"`   // 5-200 chars
    Content string   `json:"content" validate:"required,min=10"`        // Min 10 chars
    Tags    []string `json:"tags" validate:"omitempty,max=5,dive,min=1"` // Max 5 tags
    Public  bool     `json:"public"`                                     // Optional
}
```

### **4. Handle Validation in Middleware**

```go
func ValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip validation for certain routes
        if c.Request.Method == "GET" || c.FullPath() == "/health" {
            c.Next()
            return
        }

        // Add validation context
        c.Set("validator", validator.GetValidator())
        c.Next()
    }
}
```

### **5. Testing Validation**

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        wantErr bool
        errors  []string
    }{
        {
            name: "valid user",
            user: User{
                Name:  "John Doe",
                Email: "john@example.com",
                Age:   25,
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            user: User{
                Name:  "John Doe",
                Email: "invalid-email",
                Age:   25,
            },
            wantErr: true,
            errors:  []string{"email"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errors := validator.ValidateStruct(tt.user)

            if tt.wantErr {
                assert.NotNil(t, errors)
                for _, field := range tt.errors {
                    assert.Contains(t, errors, field)
                }
            } else {
                assert.Nil(t, errors)
            }
        })
    }
}
```

## 📚 Validation Tag Reference

### **String Validation**

```go
type StringTags struct {
    Required    string `validate:"required"`           // Must be present
    Email       string `validate:"email"`              // Valid email format
    URL         string `validate:"url"`                // Valid URL format
    UUID        string `validate:"uuid"`               // Valid UUID format
    Alpha       string `validate:"alpha"`              // Only letters
    AlphaNum    string `validate:"alphanum"`           // Letters and numbers
    Numeric     string `validate:"numeric"`            // Only numbers
    Lowercase   string `validate:"lowercase"`          // All lowercase
    Uppercase   string `validate:"uppercase"`          // All uppercase
    Base64      string `validate:"base64"`             // Valid base64
    JSON        string `validate:"json"`               // Valid JSON string
}
```

### **Number Validation**

```go
type NumberTags struct {
    Min         int `validate:"min=1"`          // Minimum value
    Max         int `validate:"max=100"`        // Maximum value
    GreaterThan int `validate:"gt=0"`           // Greater than
    LessThan    int `validate:"lt=200"`         // Less than
    GreaterEq   int `validate:"gte=18"`         // Greater than or equal
    LessEq      int `validate:"lte=65"`         // Less than or equal
    OneOf       int `validate:"oneof=1 2 3"`    // One of specific values
}
```

### **Array/Slice Validation**

```go
type ArrayTags struct {
    MinItems    []string `validate:"min=1"`              // Minimum array length
    MaxItems    []string `validate:"max=10"`             // Maximum array length
    DiveMin     []string `validate:"dive,min=1"`         // Each element min length
    DiveMax     []string `validate:"dive,max=50"`        // Each element max length
    DiveEmail   []string `validate:"dive,email"`         // Each element is email
    Unique      []string `validate:"unique"`             // All elements unique
}
```

## 🔗 Related Packages

- [`pkg/response`](../response/) - API response formatting
- [`pkg/errors`](../errors/) - Error handling
- [`internal/middleware`](../../internal/middleware/) - Request validation middleware

## 📚 Additional Resources

- [Go Playground Validator Documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Validation Tag Reference](https://github.com/go-playground/validator#baked-in-validations)
- [Custom Validators Guide](https://github.com/go-playground/validator#custom-validation-functions)
