package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents the standard API response structure
type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      *ErrorInfo  `json:"error,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// ErrorInfo represents error details
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details interface{}       `json:"details,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Meta represents pagination and additional metadata
type Meta struct {
	Page        int   `json:"page,omitempty"`
	Limit       int   `json:"limit,omitempty"`
	Total       int64 `json:"total,omitempty"`
	TotalPages  int   `json:"total_pages,omitempty"`
	HasNext     bool  `json:"has_next,omitempty"`
	HasPrevious bool  `json:"has_previous,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
		Timestamp:  time.Now().UTC(),
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, statusCode int, message string, data interface{}, meta *Meta) {
	c.JSON(statusCode, Response{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
		Meta:       meta,
		Timestamp:  time.Now().UTC(),
	})
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, Response{
		StatusCode: statusCode,
		Message:    "Request failed",
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string, fields map[string]string) {
	c.JSON(http.StatusBadRequest, Response{
		StatusCode: http.StatusBadRequest,
		Message:    "Validation failed",
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Fields:  fields,
		},
		Timestamp: time.Now().UTC(),
	})
}

// Pagination creates pagination metadata
func Pagination(page, limit int, total int64) *Meta {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &Meta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}
