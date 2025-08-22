# ❌ Errors Package

Structured error handling with consistent error codes, HTTP status mapping, and detailed error information for better debugging and API responses.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Error Structure](#error-structure)
- [Predefined Errors](#predefined-errors)
- [Creating Custom Errors](#creating-custom-errors)
- [Error Handling Patterns](#error-handling-patterns)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in flex-service
import "flex-service/pkg/errors"
```

## ⚡ Quick Start

### Basic Error Usage (Old Way)

```go
// Traditional way - verbose
func findUser(id string) (*User, error) {
    if id == "" {
        return nil, errors.New(errors.ErrBadRequest, "User ID is required", 400)
    }

    user := getUserFromDB(id)
    if user == nil {
        return nil, errors.New(errors.ErrNotFound, "User not found", 404)
    }

    return user, nil
}
```

### New Helper Functions (Recommended)

```go
package main

import (
    "fmt"
    "flex-service/pkg/errors"
)

func findUser(id string) (*User, error) {
    if id == "" {
        return nil, errors.BadRequest("User ID is required")
    }

    user := getUserFromDB(id)
    if user == nil {
        return nil, errors.UserNotFound()
    }

    return user, nil
}

func authenticateUser(email, password string) (*User, error) {
    user, err := getUserByEmail(email)
    if err != nil {
        return nil, errors.WrapDatabase(err, "Failed to find user")
    }

    if !verifyPassword(password, user.Password) {
        return nil, errors.InvalidCredentials()
    }

    if !user.IsActive {
        return nil, errors.AccountDisabled()
    }

    return user, nil
}

func main() {
    user, err := findUser("")
    if err != nil {
        if appErr, ok := err.(*errors.AppError); ok {
            fmt.Printf("Error Code: %s\n", appErr.Code)
            fmt.Printf("Message: %s\n", appErr.Message)
            fmt.Printf("Status: %d\n", appErr.StatusCode)
            fmt.Printf("Details: %+v\n", appErr.Details)
        }
    }
}
```

## 🏗️ Error Structure

### **AppError Structure**

```go
type AppError struct {
    Code       string      `json:"code"`        // Error code (e.g., "USER_NOT_FOUND")
    Message    string      `json:"message"`     // Human-readable message
    StatusCode int         `json:"-"`          // HTTP status code
    Details    interface{} `json:"details,omitempty"` // Additional details
    Cause      error       `json:"-"`          // Original error (for wrapping)
}
```

### **Error Implementation**

```go
func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}
```

## 📝 Predefined Errors

### **General Errors**

| Error                  | Code             | Status | Use Case                |
| ---------------------- | ---------------- | ------ | ----------------------- |
| `ErrInternalServer`    | `INTERNAL_ERROR` | 500    | System failures         |
| `ErrNotFoundError`     | `NOT_FOUND`      | 404    | Resource not found      |
| `ErrBadRequestError`   | `BAD_REQUEST`    | 400    | Invalid input           |
| `ErrUnauthorizedError` | `UNAUTHORIZED`   | 401    | Authentication required |
| `ErrForbiddenError`    | `FORBIDDEN`      | 403    | Permission denied       |

### **Authentication Errors**

| Error                        | Code                  | Status | Use Case    |
| ---------------------------- | --------------------- | ------ | ----------- |
| `ErrInvalidCredentialsError` | `INVALID_CREDENTIALS` | 401    | Wrong login |
| `ErrTokenExpiredError`       | `TOKEN_EXPIRED`       | 401    | JWT expired |
| `ErrTokenInvalidError`       | `TOKEN_INVALID`       | 401    | Invalid JWT |

### **Error Constants**

```go
const (
    // General errors
    ErrInternal     = "INTERNAL_ERROR"
    ErrNotFound     = "NOT_FOUND"
    ErrBadRequest   = "BAD_REQUEST"
    ErrUnauthorized = "UNAUTHORIZED"
    ErrForbidden    = "FORBIDDEN"
    ErrConflict     = "CONFLICT"
    ErrValidation   = "VALIDATION_ERROR"

    // Auth errors
    ErrInvalidCredentials = "INVALID_CREDENTIALS"
    ErrTokenExpired       = "TOKEN_EXPIRED"
    ErrTokenInvalid       = "TOKEN_INVALID"
    ErrUserExists         = "USER_EXISTS"
    ErrUserNotFound       = "USER_NOT_FOUND"
)
```

## 🚀 Helper Functions (NEW)

### **Convenience Functions**

| Function            | Description             | Status Code | Example                                     |
| ------------------- | ----------------------- | ----------- | ------------------------------------------- |
| `NotFound(msg)`     | Resource not found      | 404         | `errors.NotFound("User not found")`         |
| `BadRequest(msg)`   | Invalid input           | 400         | `errors.BadRequest("Invalid email format")` |
| `Unauthorized(msg)` | Authentication required | 401         | `errors.Unauthorized("Login required")`     |
| `Forbidden(msg)`    | Permission denied       | 403         | `errors.Forbidden("Admin access required")` |
| `Conflict(msg)`     | Resource conflict       | 409         | `errors.Conflict("Email already exists")`   |
| `Internal(msg)`     | Server error            | 500         | `errors.Internal("Service unavailable")`    |
| `Validation(msg)`   | Validation failed       | 400         | `errors.Validation("Invalid input data")`   |

### **Wrapping Functions**

| Function                     | Description          | Use Case                  |
| ---------------------------- | -------------------- | ------------------------- |
| `WrapInternal(err, msg)`     | Wrap as server error | Database failures         |
| `WrapNotFound(err, msg)`     | Wrap as not found    | Query returned no results |
| `WrapBadRequest(err, msg)`   | Wrap as bad request  | Parsing failures          |
| `WrapUnauthorized(err, msg)` | Wrap as unauthorized | Token validation failures |

### **Database-specific Helpers**

| Function                 | Description               | Status Code |
| ------------------------ | ------------------------- | ----------- |
| `DatabaseError(msg)`     | Database operation failed | 500         |
| `WrapDatabase(err, msg)` | Wrap database error       | 500         |

### **Auth-specific Helpers**

| Function                   | Description             | Status Code |
| -------------------------- | ----------------------- | ----------- |
| `InvalidCredentials()`     | Wrong login credentials | 401         |
| `TokenExpired()`           | JWT token expired       | 401         |
| `TokenInvalid()`           | Invalid JWT token       | 401         |
| `UserExists(field)`        | User already exists     | 409         |
| `UserNotFound()`           | User not found          | 404         |
| `AccountDisabled()`        | Account is disabled     | 401         |
| `TokenError(msg)`          | Token operation failed  | 500         |
| `WrapTokenError(err, msg)` | Wrap token error        | 500         |

### **Helper Functions Usage Examples**

```go
// Before (verbose)
return nil, errors.New(errors.ErrNotFound, "User not found", 404)

// After (clean)
return nil, errors.UserNotFound()

// Before (complex wrapping)
return nil, errors.Wrap(err, "DATABASE_ERROR", "failed to get user", 500)

// After (simple)
return nil, errors.WrapDatabase(err, "failed to get user")

// Before (manual auth errors)
return nil, errors.New(errors.ErrInvalidCredentials, "Invalid email or password", 401)

// After (predefined)
return nil, errors.InvalidCredentials()
```

## 🛠️ Creating Custom Errors

### **1. Simple Error Creation**

```go
// Create new error
func validateAge(age int) error {
    if age < 18 {
        return errors.New(
            errors.ErrValidation,
            "Age must be at least 18 years old",
            400,
        )
    }
    return nil
}
```

### **2. Error with Details**

```go
func processPayment(amount float64, cardToken string) error {
    if amount <= 0 {
        return errors.New(
            "INVALID_AMOUNT",
            "Payment amount must be greater than zero",
            400,
        ).WithDetails(map[string]interface{}{
            "provided_amount": amount,
            "minimum_amount": 0.01,
            "currency": "USD",
        })
    }

    if cardToken == "" {
        return errors.New(
            "MISSING_PAYMENT_METHOD",
            "Payment method is required",
            400,
        ).WithDetails(map[string]interface{}{
            "required_fields": []string{"card_token", "payment_method"},
        })
    }

    return nil
}
```

### **3. Wrapping Existing Errors**

```go
func getUserFromDatabase(id string) (*User, error) {
    user := &User{}
    err := db.Where("id = ?", id).First(user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, errors.New(
                errors.ErrUserNotFound,
                "User not found",
                404,
            ).WithDetails(map[string]interface{}{
                "user_id": id,
            })
        }

        return nil, errors.Wrap(
            err,
            errors.ErrInternal,
            "Database operation failed",
            500,
        )
    }

    return user, nil
}
```

## 🎯 Error Handling Patterns

### **1. Repository Layer Error Handling**

```go
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) GetByID(id string) (*User, error) {
    var user User
    err := r.db.Where("id = ?", id).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, errors.New(
                errors.ErrUserNotFound,
                "User not found",
                404,
            ).WithDetails(map[string]interface{}{
                "user_id": id,
            })
        }
        return nil, errors.Wrap(
            err,
            errors.ErrInternal,
            "Failed to retrieve user",
            500,
        )
    }
    return &user, nil
}

func (r *UserRepository) Create(user *User) error {
    err := r.db.Create(user).Error
    if err != nil {
        if isDuplicateError(err) {
            return errors.New(
                errors.ErrUserExists,
                "User with this email already exists",
                409,
            ).WithDetails(map[string]interface{}{
                "email": user.Email,
            })
        }
        return errors.Wrap(
            err,
            errors.ErrInternal,
            "Failed to create user",
            500,
        )
    }
    return nil
}
```

### **2. Service Layer Error Handling**

```go
type UserService struct {
    repo UserRepository
}

func (s *UserService) RegisterUser(req RegisterRequest) (*User, error) {
    // Validate business rules
    if err := s.validateRegistration(req); err != nil {
        return nil, err // Already an AppError
    }

    // Check if user exists
    existingUser, err := s.repo.GetByEmail(req.Email)
    if err != nil {
        // Only return error if it's not "not found"
        if appErr, ok := err.(*errors.AppError); ok {
            if appErr.Code != errors.ErrUserNotFound {
                return nil, err
            }
        } else {
            return nil, errors.Wrap(
                err,
                errors.ErrInternal,
                "Failed to check user existence",
                500,
            )
        }
    }

    if existingUser != nil {
        return nil, errors.New(
            errors.ErrUserExists,
            "User with this email already exists",
            409,
        ).WithDetails(map[string]interface{}{
            "email": req.Email,
        })
    }

    // Create user
    user := &User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashPassword(req.Password),
    }

    if err := s.repo.Create(user); err != nil {
        return nil, err // Already wrapped by repository
    }

    return user, nil
}

func (s *UserService) validateRegistration(req RegisterRequest) error {
    var errors []string

    if len(req.Password) < 8 {
        errors = append(errors, "Password must be at least 8 characters")
    }

    if !isValidEmail(req.Email) {
        errors = append(errors, "Invalid email format")
    }

    if len(errors) > 0 {
        return errors.New(
            errors.ErrValidation,
            "Registration validation failed",
            400,
        ).WithDetails(map[string]interface{}{
            "validation_errors": errors,
        })
    }

    return nil
}
```

### **3. Handler Layer Error Handling**

```go
func RegisterHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := errors.New(
            errors.ErrValidation,
            "Invalid request format",
            400,
        ).WithDetails(map[string]interface{}{
            "parse_error": err.Error(),
        })

        handleError(c, appErr)
        return
    }

    user, err := userService.RegisterUser(req)
    if err != nil {
        handleError(c, err)
        return
    }

    response.Success(c, 201, "User registered successfully", user)
}

// Generic error handler
func handleError(c *gin.Context, err error) {
    if appErr, ok := err.(*errors.AppError); ok {
        response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
        return
    }

    // Fallback for unexpected errors
    logger.Error("Unexpected error", zap.Error(err))
    response.Error(c, 500, errors.ErrInternal, "Internal server error", nil)
}
```

## 💡 Real-World Examples

### **1. E-commerce Order Processing**

```go
func (s *OrderService) CreateOrder(req CreateOrderRequest) (*Order, error) {
    // Validate cart
    if len(req.Items) == 0 {
        return nil, errors.New(
            "EMPTY_CART",
            "Cart cannot be empty",
            400,
        )
    }

    // Check product availability
    for _, item := range req.Items {
        product, err := s.productRepo.GetByID(item.ProductID)
        if err != nil {
            return nil, err
        }

        if product.Stock < item.Quantity {
            return nil, errors.New(
                "INSUFFICIENT_STOCK",
                "Insufficient product stock",
                409,
            ).WithDetails(map[string]interface{}{
                "product_id": item.ProductID,
                "requested_quantity": item.Quantity,
                "available_stock": product.Stock,
            })
        }
    }

    // Calculate total
    total, err := s.calculateTotal(req.Items)
    if err != nil {
        return nil, errors.Wrap(
            err,
            "CALCULATION_ERROR",
            "Failed to calculate order total",
            500,
        )
    }

    // Process payment
    payment, err := s.paymentService.ProcessPayment(req.PaymentMethod, total)
    if err != nil {
        return nil, errors.Wrap(
            err,
            "PAYMENT_FAILED",
            "Payment processing failed",
            402,
        ).WithDetails(map[string]interface{}{
            "amount": total,
            "payment_method": req.PaymentMethod,
        })
    }

    // Create order
    order := &Order{
        UserID:    req.UserID,
        Items:     req.Items,
        Total:     total,
        PaymentID: payment.ID,
        Status:    "confirmed",
    }

    if err := s.orderRepo.Create(order); err != nil {
        // Rollback payment if order creation fails
        s.paymentService.RefundPayment(payment.ID)
        return nil, err
    }

    return order, nil
}
```

### **2. File Upload with Validation**

```go
func (s *FileService) UploadFile(file multipart.File, header *multipart.FileHeader) (*UploadedFile, error) {
    // Validate file size (10MB limit)
    if header.Size > 10*1024*1024 {
        return nil, errors.New(
            "FILE_TOO_LARGE",
            "File size exceeds maximum limit",
            413,
        ).WithDetails(map[string]interface{}{
            "file_size": header.Size,
            "max_size": 10*1024*1024,
            "max_size_mb": 10,
        })
    }

    // Validate file type
    allowedTypes := []string{".jpg", ".jpeg", ".png", ".pdf", ".doc", ".docx"}
    ext := filepath.Ext(header.Filename)
    if !contains(allowedTypes, ext) {
        return nil, errors.New(
            "INVALID_FILE_TYPE",
            "File type not allowed",
            415,
        ).WithDetails(map[string]interface{}{
            "file_extension": ext,
            "allowed_types": allowedTypes,
        })
    }

    // Check for malware (example)
    if s.containsMalware(file) {
        return nil, errors.New(
            "MALWARE_DETECTED",
            "File failed security scan",
            422,
        ).WithDetails(map[string]interface{}{
            "filename": header.Filename,
            "scan_result": "malware_detected",
        })
    }

    // Save file
    savedFile, err := s.saveToStorage(file, header)
    if err != nil {
        return nil, errors.Wrap(
            err,
            "STORAGE_ERROR",
            "Failed to save file to storage",
            500,
        ).WithDetails(map[string]interface{}{
            "filename": header.Filename,
            "storage_provider": "aws_s3",
        })
    }

    return savedFile, nil
}
```

### **3. Authentication Service**

```go
func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
    // Find user
    user, err := s.userRepo.GetByEmail(email)
    if err != nil {
        if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.ErrUserNotFound {
            return nil, errors.ErrInvalidCredentialsError
        }
        return nil, err
    }

    // Check if account is active
    if !user.IsActive {
        return nil, errors.New(
            "ACCOUNT_SUSPENDED",
            "Account has been suspended",
            403,
        ).WithDetails(map[string]interface{}{
            "user_id": user.ID,
            "suspension_reason": user.SuspensionReason,
        })
    }

    // Verify password
    if !s.verifyPassword(password, user.Password) {
        // Log failed attempt
        s.logFailedLogin(user.ID, "invalid_password")

        return nil, errors.ErrInvalidCredentialsError
    }

    // Check for too many failed attempts
    if s.isAccountLocked(user.ID) {
        return nil, errors.New(
            "ACCOUNT_LOCKED",
            "Account temporarily locked due to multiple failed login attempts",
            423,
        ).WithDetails(map[string]interface{}{
            "user_id": user.ID,
            "unlock_time": time.Now().Add(30 * time.Minute),
        })
    }

    // Generate tokens
    accessToken, err := s.generateAccessToken(user)
    if err != nil {
        return nil, errors.Wrap(
            err,
            "TOKEN_GENERATION_ERROR",
            "Failed to generate access token",
            500,
        )
    }

    refreshToken, err := s.generateRefreshToken(user)
    if err != nil {
        return nil, errors.Wrap(
            err,
            "TOKEN_GENERATION_ERROR",
            "Failed to generate refresh token",
            500,
        )
    }

    return &LoginResponse{
        User:         user,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    time.Now().Add(24 * time.Hour),
    }, nil
}
```

## 🎯 Best Practices

### **1. Consistent Error Codes**

```go
// Use hierarchical error codes
const (
    // User-related errors
    ErrUserNotFound     = "USER_NOT_FOUND"
    ErrUserExists       = "USER_EXISTS"
    ErrUserSuspended    = "USER_SUSPENDED"

    // Product-related errors
    ErrProductNotFound  = "PRODUCT_NOT_FOUND"
    ErrProductOutOfStock = "PRODUCT_OUT_OF_STOCK"
    ErrProductInactive  = "PRODUCT_INACTIVE"

    // Order-related errors
    ErrOrderNotFound    = "ORDER_NOT_FOUND"
    ErrOrderCancelled   = "ORDER_CANCELLED"
    ErrOrderProcessed   = "ORDER_ALREADY_PROCESSED"
)
```

### **2. Error Context and Tracing**

```go
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    span := trace.SpanFromContext(ctx)

    user, err := s.repo.GetByID(id)
    if err != nil {
        span.RecordError(err)

        if appErr, ok := err.(*errors.AppError); ok {
            // Add tracing information
            appErr.Details = map[string]interface{}{
                "trace_id": span.SpanContext().TraceID().String(),
                "span_id":  span.SpanContext().SpanID().String(),
                "user_id":  id,
            }
        }

        return nil, err
    }

    return user, nil
}
```

### **3. Error Logging and Monitoring**

```go
func logError(err error, ctx context.Context) {
    if appErr, ok := err.(*errors.AppError); ok {
        logger.Error("Application error",
            zap.String("code", appErr.Code),
            zap.String("message", appErr.Message),
            zap.Int("status_code", appErr.StatusCode),
            zap.Any("details", appErr.Details),
            zap.Error(appErr.Cause),
        )

        // Send to monitoring service
        if appErr.StatusCode >= 500 {
            monitoring.RecordError(appErr)
        }
    } else {
        logger.Error("Unexpected error", zap.Error(err))
        monitoring.RecordError(err)
    }
}
```

### **4. Error Recovery and Fallbacks**

```go
func (s *NotificationService) SendEmail(to, subject, body string) error {
    // Try primary email service
    err := s.primaryEmailService.Send(to, subject, body)
    if err != nil {
        logger.Warn("Primary email service failed, trying fallback",
            zap.Error(err),
            zap.String("recipient", to),
        )

        // Try fallback service
        err = s.fallbackEmailService.Send(to, subject, body)
        if err != nil {
            return errors.Wrap(
                err,
                "EMAIL_DELIVERY_FAILED",
                "Failed to send email via all providers",
                500,
            ).WithDetails(map[string]interface{}{
                "recipient": to,
                "attempted_providers": []string{"primary", "fallback"},
            })
        }
    }

    return nil
}
```

### **5. Testing Error Scenarios**

```go
func TestUserService_CreateUser_EmailExists(t *testing.T) {
    userRepo := &MockUserRepository{}
    userService := NewUserService(userRepo)

    // Setup: existing user
    existingUser := &User{Email: "test@example.com"}
    userRepo.On("GetByEmail", "test@example.com").Return(existingUser, nil)

    // Test
    req := RegisterRequest{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "password123",
    }

    user, err := userService.RegisterUser(req)

    // Assert
    assert.Nil(t, user)
    assert.NotNil(t, err)

    appErr, ok := err.(*errors.AppError)
    assert.True(t, ok)
    assert.Equal(t, errors.ErrUserExists, appErr.Code)
    assert.Equal(t, 409, appErr.StatusCode)
    assert.Contains(t, appErr.Details, "email")
}
```

## 🔄 Error Middleware

### **Global Error Handler**

```go
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Handle any errors that occurred during request processing
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            if appErr, ok := err.(*errors.AppError); ok {
                response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
            } else {
                logger.Error("Unhandled error", zap.Error(err))
                response.Error(c, 500, errors.ErrInternal, "Internal server error", nil)
            }

            c.Abort()
        }
    }
}
```

## 🔗 Related Packages

- [`pkg/response`](../response/) - API response formatting
- [`pkg/validator`](../validator/) - Input validation
- [`pkg/logger`](../logger/) - Error logging

## 📚 Additional Resources

- [Error Handling in Go](https://blog.golang.org/error-handling-and-go)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [API Error Design Guidelines](https://cloud.google.com/apis/design/errors)
