# 📤 Response Package

Standardized HTTP response utilities for consistent API responses with JSON formatting, error handling, and pagination support.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Response Structure](#response-structure)
- [Success Responses](#success-responses)
- [Error Responses](#error-responses)
- [Pagination](#pagination)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in flex-service
import "flex-service/pkg/response"
```

## ⚡ Quick Start

### Basic Success Response

```go
package main

import (
    "flex-service/pkg/response"
    "github.com/gin-gonic/gin"
)

func GetUserHandler(c *gin.Context) {
    user := User{
        ID:    1,
        Name:  "John Doe",
        Email: "john@example.com",
    }

    response.Success(c, 200, "User retrieved successfully", user)
}

// Response output:
// {
//   "status_code": 200,
//   "message": "User retrieved successfully",
//   "data": {
//     "id": 1,
//     "name": "John Doe",
//     "email": "john@example.com"
//   },
//   "timestamp": "2024-01-15T10:30:00Z"
// }
```

## 📊 Response Structure

### **Standard Response Format**

All API responses follow this consistent structure:

```go
type Response struct {
    StatusCode int         `json:"status_code"`
    Message    string      `json:"message"`
    Data       interface{} `json:"data,omitempty"`
    Error      *ErrorInfo  `json:"error,omitempty"`
    Meta       *Meta       `json:"meta,omitempty"`
    Timestamp  time.Time   `json:"timestamp"`
}
```

### **Error Information Structure**

```go
type ErrorInfo struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Details interface{}       `json:"details,omitempty"`
    Fields  map[string]string `json:"fields,omitempty"`
}
```

### **Pagination Metadata**

```go
type Meta struct {
    Page        int   `json:"page,omitempty"`
    Limit       int   `json:"limit,omitempty"`
    Total       int64 `json:"total,omitempty"`
    TotalPages  int   `json:"total_pages,omitempty"`
    HasNext     bool  `json:"has_next,omitempty"`
    HasPrevious bool  `json:"has_previous,omitempty"`
}
```

## ✅ Success Responses

### **1. Basic Success Response**

```go
func CreateUserHandler(c *gin.Context) {
    // Create user logic...

    response.Success(c, 201, "User created successfully", gin.H{
        "user_id": userID,
        "name":    user.Name,
        "email":   user.Email,
    })
}

// Output:
// {
//   "status_code": 201,
//   "message": "User created successfully",
//   "data": {
//     "user_id": "uuid-here",
//     "name": "John Doe",
//     "email": "john@example.com"
//   },
//   "timestamp": "2024-01-15T10:30:00Z"
// }
```

### **2. Success with Pagination**

```go
func ListUsersHandler(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    limit := c.DefaultQuery("limit", "10")

    users, total, err := getUsersPaginated(page, limit)
    if err != nil {
        response.Error(c, 500, "DATABASE_ERROR", "Failed to retrieve users", nil)
        return
    }

    meta := response.Pagination(
        parseIntOrDefault(page, 1),
        parseIntOrDefault(limit, 10),
        total,
    )

    response.SuccessWithMeta(c, 200, "Users retrieved successfully", users, meta)
}

// Output:
// {
//   "status_code": 200,
//   "message": "Users retrieved successfully",
//   "data": [
//     {"id": 1, "name": "John", "email": "john@example.com"},
//     {"id": 2, "name": "Jane", "email": "jane@example.com"}
//   ],
//   "meta": {
//     "page": 1,
//     "limit": 10,
//     "total": 25,
//     "total_pages": 3,
//     "has_next": true,
//     "has_previous": false
//   },
//   "timestamp": "2024-01-15T10:30:00Z"
// }
```

### **3. Success without Data**

```go
func DeleteUserHandler(c *gin.Context) {
    userID := c.Param("id")

    err := deleteUser(userID)
    if err != nil {
        response.Error(c, 500, "DELETE_FAILED", "Failed to delete user", nil)
        return
    }

    response.Success(c, 200, "User deleted successfully", nil)
}

// Output:
// {
//   "status_code": 200,
//   "message": "User deleted successfully",
//   "timestamp": "2024-01-15T10:30:00Z"
// }
```

## ❌ Error Responses

### **1. Basic Error Response**

```go
func GetUserHandler(c *gin.Context) {
    userID := c.Param("id")

    user, err := findUserByID(userID)
    if err != nil {
        if err == ErrUserNotFound {
            response.Error(c, 404, "USER_NOT_FOUND", "User not found", nil)
            return
        }
        response.Error(c, 500, "DATABASE_ERROR", "Internal server error", nil)
        return
    }

    response.Success(c, 200, "User retrieved successfully", user)
}

// Error Output:
// {
//   "status_code": 404,
//   "message": "Request failed",
//   "error": {
//     "code": "USER_NOT_FOUND",
//     "message": "User not found"
//   },
//   "timestamp": "2024-01-15T10:30:00Z"
// }
```

### **2. Validation Error Response**

```go
func CreateUserHandler(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid JSON format", nil)
        return
    }

    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    // Create user...
}

// Validation Error Output:
// {
//   "status_code": 400,
//   "message": "Validation failed",
//   "error": {
//     "code": "VALIDATION_ERROR",
//     "message": "Validation failed",
//     "fields": {
//       "name": "name is required",
//       "email": "email must be a valid email",
//       "password": "password must be at least 8 characters"
//     }
//   },
//   "timestamp": "2024-01-15T10:30:00Z"
// }
```

### **3. Error with Details**

```go
func ProcessPaymentHandler(c *gin.Context) {
    payment, err := processPayment(paymentData)
    if err != nil {
        response.Error(c, 400, "PAYMENT_FAILED", "Payment processing failed", gin.H{
            "transaction_id": paymentData.TransactionID,
            "reason":         err.Error(),
            "retry_after":    300, // seconds
        })
        return
    }

    response.Success(c, 200, "Payment processed successfully", payment)
}

// Error with Details Output:
// {
//   "status_code": 400,
//   "message": "Request failed",
//   "error": {
//     "code": "PAYMENT_FAILED",
//     "message": "Payment processing failed",
//     "details": {
//       "transaction_id": "txn_123456",
//       "reason": "Insufficient funds",
//       "retry_after": 300
//     }
//   },
//   "timestamp": "2024-01-15T10:30:00Z"
// }
```

## 📊 Pagination

### **Create Pagination Metadata**

```go
func ListProductsHandler(c *gin.Context) {
    page := parseIntOrDefault(c.Query("page"), 1)
    limit := parseIntOrDefault(c.Query("limit"), 10)

    products, total, err := getProducts(page, limit)
    if err != nil {
        response.Error(c, 500, "DATABASE_ERROR", "Failed to retrieve products", nil)
        return
    }

    meta := response.Pagination(page, limit, total)
    response.SuccessWithMeta(c, 200, "Products retrieved successfully", products, meta)
}

// Helper function
func parseIntOrDefault(s string, defaultVal int) int {
    if i, err := strconv.Atoi(s); err == nil && i > 0 {
        return i
    }
    return defaultVal
}
```

### **Advanced Pagination Example**

```go
func ListOrdersHandler(c *gin.Context) {
    // Parse query parameters
    page := parseIntOrDefault(c.Query("page"), 1)
    limit := parseIntOrDefault(c.Query("limit"), 10)
    status := c.Query("status")
    userID := c.Query("user_id")

    // Validate limit
    if limit > 100 {
        limit = 100 // Maximum limit
    }

    orders, total, err := getOrdersWithFilters(page, limit, status, userID)
    if err != nil {
        response.Error(c, 500, "DATABASE_ERROR", "Failed to retrieve orders", nil)
        return
    }

    meta := response.Pagination(page, limit, total)

    response.SuccessWithMeta(c, 200, "Orders retrieved successfully", gin.H{
        "orders": orders,
        "filters": gin.H{
            "status":  status,
            "user_id": userID,
        },
    }, meta)
}
```

## 🎯 Real-World Examples

### **1. Authentication Endpoints**

```go
// Login endpoint
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request format", nil)
        return
    }

    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    user, token, err := authenticateUser(req.Email, req.Password)
    if err != nil {
        response.Error(c, 401, "INVALID_CREDENTIALS", "Invalid email or password", nil)
        return
    }

    response.Success(c, 200, "Login successful", gin.H{
        "user":  user,
        "token": token,
        "expires_at": time.Now().Add(24 * time.Hour),
    })
}

// Register endpoint
func RegisterHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request format", nil)
        return
    }

    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    if userExists(req.Email) {
        response.Error(c, 409, "USER_EXISTS", "Email already registered", gin.H{
            "email": req.Email,
        })
        return
    }

    user, err := createUser(req)
    if err != nil {
        response.Error(c, 500, "CREATION_FAILED", "Failed to create user", nil)
        return
    }

    response.Success(c, 201, "User registered successfully", gin.H{
        "user_id": user.ID,
        "email":   user.Email,
    })
}
```

### **2. CRUD Operations**

```go
// Create resource
func CreatePostHandler(c *gin.Context) {
    var req CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request format", nil)
        return
    }

    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    post, err := createPost(req, getUserFromContext(c))
    if err != nil {
        response.Error(c, 500, "CREATION_FAILED", "Failed to create post", nil)
        return
    }

    response.Success(c, 201, "Post created successfully", post)
}

// Update resource
func UpdatePostHandler(c *gin.Context) {
    postID := c.Param("id")

    var req UpdatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request format", nil)
        return
    }

    if errors := validator.ValidateStruct(req); errors != nil {
        response.ValidationError(c, "Validation failed", errors)
        return
    }

    post, err := updatePost(postID, req, getUserFromContext(c))
    if err != nil {
        if err == ErrPostNotFound {
            response.Error(c, 404, "POST_NOT_FOUND", "Post not found", nil)
            return
        }
        if err == ErrUnauthorized {
            response.Error(c, 403, "FORBIDDEN", "You don't have permission to update this post", nil)
            return
        }
        response.Error(c, 500, "UPDATE_FAILED", "Failed to update post", nil)
        return
    }

    response.Success(c, 200, "Post updated successfully", post)
}

// Delete resource
func DeletePostHandler(c *gin.Context) {
    postID := c.Param("id")

    err := deletePost(postID, getUserFromContext(c))
    if err != nil {
        if err == ErrPostNotFound {
            response.Error(c, 404, "POST_NOT_FOUND", "Post not found", nil)
            return
        }
        if err == ErrUnauthorized {
            response.Error(c, 403, "FORBIDDEN", "You don't have permission to delete this post", nil)
            return
        }
        response.Error(c, 500, "DELETE_FAILED", "Failed to delete post", nil)
        return
    }

    response.Success(c, 200, "Post deleted successfully", nil)
}
```

### **3. File Upload Endpoint**

```go
func UploadFileHandler(c *gin.Context) {
    // Parse multipart form
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        response.Error(c, 400, "NO_FILE", "No file uploaded", nil)
        return
    }
    defer file.Close()

    // Validate file size (5MB limit)
    if header.Size > 5*1024*1024 {
        response.Error(c, 400, "FILE_TOO_LARGE", "File size exceeds 5MB limit", gin.H{
            "max_size": "5MB",
            "file_size": fmt.Sprintf("%.2fMB", float64(header.Size)/(1024*1024)),
        })
        return
    }

    // Validate file type
    if !isValidFileType(header.Filename) {
        response.Error(c, 400, "INVALID_FILE_TYPE", "File type not allowed", gin.H{
            "allowed_types": []string{".jpg", ".jpeg", ".png", ".pdf", ".doc", ".docx"},
            "file_type": filepath.Ext(header.Filename),
        })
        return
    }

    // Save file
    uploadedFile, err := saveFile(file, header)
    if err != nil {
        response.Error(c, 500, "UPLOAD_FAILED", "Failed to save file", nil)
        return
    }

    response.Success(c, 201, "File uploaded successfully", gin.H{
        "file_id":   uploadedFile.ID,
        "filename":  uploadedFile.Filename,
        "size":      uploadedFile.Size,
        "url":       uploadedFile.URL,
        "mime_type": uploadedFile.MimeType,
    })
}
```

### **4. Search and Filter Endpoint**

```go
func SearchUsersHandler(c *gin.Context) {
    // Parse query parameters
    query := c.Query("q")
    page := parseIntOrDefault(c.Query("page"), 1)
    limit := parseIntOrDefault(c.Query("limit"), 10)
    role := c.Query("role")
    status := c.Query("status")
    sortBy := c.DefaultQuery("sort", "created_at")
    sortOrder := c.DefaultQuery("order", "desc")

    // Validate search query
    if query == "" {
        response.Error(c, 400, "MISSING_QUERY", "Search query is required", nil)
        return
    }

    if len(query) < 2 {
        response.Error(c, 400, "QUERY_TOO_SHORT", "Search query must be at least 2 characters", nil)
        return
    }

    // Validate sort parameters
    validSortFields := []string{"name", "email", "created_at", "updated_at"}
    if !contains(validSortFields, sortBy) {
        response.Error(c, 400, "INVALID_SORT_FIELD", "Invalid sort field", gin.H{
            "valid_fields": validSortFields,
            "provided": sortBy,
        })
        return
    }

    // Perform search
    searchParams := SearchParams{
        Query:     query,
        Page:      page,
        Limit:     limit,
        Role:      role,
        Status:    status,
        SortBy:    sortBy,
        SortOrder: sortOrder,
    }

    users, total, err := searchUsers(searchParams)
    if err != nil {
        response.Error(c, 500, "SEARCH_FAILED", "Search operation failed", nil)
        return
    }

    meta := response.Pagination(page, limit, total)

    response.SuccessWithMeta(c, 200, "Search completed successfully", gin.H{
        "users": users,
        "search_params": gin.H{
            "query":      query,
            "role":       role,
            "status":     status,
            "sort_by":    sortBy,
            "sort_order": sortOrder,
        },
    }, meta)
}
```

## 🎯 Best Practices

### **1. Consistent Status Codes**

```go
// Use appropriate HTTP status codes
var StatusCodes = map[string]int{
    "SUCCESS":           200, // GET, PUT, PATCH
    "CREATED":           201, // POST
    "NO_CONTENT":        204, // DELETE
    "BAD_REQUEST":       400, // Invalid input
    "UNAUTHORIZED":      401, // Authentication required
    "FORBIDDEN":         403, // Permission denied
    "NOT_FOUND":         404, // Resource not found
    "CONFLICT":          409, // Resource already exists
    "VALIDATION_ERROR":  422, // Validation failed
    "INTERNAL_ERROR":    500, // Server error
}
```

### **2. Standardized Error Codes**

```go
// Use consistent error codes across your API
const (
    // Authentication & Authorization
    ErrInvalidCredentials = "INVALID_CREDENTIALS"
    ErrTokenExpired      = "TOKEN_EXPIRED"
    ErrTokenInvalid      = "TOKEN_INVALID"
    ErrUnauthorized      = "UNAUTHORIZED"
    ErrForbidden         = "FORBIDDEN"

    // Validation
    ErrValidationFailed  = "VALIDATION_ERROR"
    ErrMissingField      = "MISSING_FIELD"
    ErrInvalidFormat     = "INVALID_FORMAT"

    // Resources
    ErrNotFound          = "NOT_FOUND"
    ErrAlreadyExists     = "ALREADY_EXISTS"
    ErrCannotDelete      = "CANNOT_DELETE"

    // System
    ErrInternalServer    = "INTERNAL_ERROR"
    ErrDatabaseError     = "DATABASE_ERROR"
    ErrExternalService   = "EXTERNAL_SERVICE_ERROR"
)
```

### **3. Response Wrapper Functions**

```go
// Create helper functions for common responses
func NotFound(c *gin.Context, resource string) {
    response.Error(c, 404, "NOT_FOUND", fmt.Sprintf("%s not found", resource), nil)
}

func Unauthorized(c *gin.Context, message string) {
    response.Error(c, 401, "UNAUTHORIZED", message, nil)
}

func BadRequest(c *gin.Context, message string, details interface{}) {
    response.Error(c, 400, "BAD_REQUEST", message, details)
}

func InternalServerError(c *gin.Context, err error) {
    // Log the actual error
    logger.Error("Internal server error", zap.Error(err))

    // Return generic error to client
    response.Error(c, 500, "INTERNAL_ERROR", "Internal server error", nil)
}
```

### **4. Testing Response Helpers**

```go
func TestSuccessResponse(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    testData := gin.H{"test": "data"}
    response.Success(c, 200, "Test successful", testData)

    var resp response.Response
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NoError(t, err)

    assert.Equal(t, 200, resp.StatusCode)
    assert.Equal(t, "Test successful", resp.Message)
    assert.NotNil(t, resp.Data)
    assert.Nil(t, resp.Error)
}

func TestErrorResponse(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    response.Error(c, 404, "NOT_FOUND", "Resource not found", nil)

    var resp response.Response
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NoError(t, err)

    assert.Equal(t, 404, resp.StatusCode)
    assert.NotNil(t, resp.Error)
    assert.Equal(t, "NOT_FOUND", resp.Error.Code)
}
```

### **5. API Documentation**

```go
// Document your response formats in OpenAPI/Swagger
// Example response documentation:

// @Success 200 {object} response.Response{data=User} "User retrieved successfully"
// @Failure 404 {object} response.Response{error=response.ErrorInfo} "User not found"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
func GetUserHandler(c *gin.Context) {
    // Implementation...
}
```

## 🔗 Related Packages

- [`pkg/validator`](../validator/) - Input validation
- [`pkg/errors`](../errors/) - Error handling
- [`internal/middleware`](../../internal/middleware/) - HTTP middleware

## 📚 Additional Resources

- [HTTP Status Codes Reference](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [REST API Design Guidelines](https://restfulapi.net/)
- [JSON API Specification](https://jsonapi.org/)
