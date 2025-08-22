package middleware

import (
	"net/http"

	"flex-service/pkg/errors"
	"flex-service/pkg/logger"
	"flex-service/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler middleware handles errors and returns standardized responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			switch e := err.Err.(type) {
			case *errors.AppError:
				// Handle application errors
				logger.Error("Application error",
					zap.String("code", e.Code),
					zap.String("message", e.Message),
					zap.Int("status", e.StatusCode),
					zap.String("path", c.Request.URL.Path),
					zap.Error(e.Cause),
				)

				response.Error(c, e.StatusCode, e.Code, e.Message, e.Details)
			default:
				// Handle unknown errors
				logger.Error("Unknown error",
					zap.String("path", c.Request.URL.Path),
					zap.Error(err.Err),
				)

				response.Error(c, http.StatusInternalServerError,
					errors.ErrInternal, "Internal server error", nil)
			}

			c.Abort()
		}
	}
}

// HandleError is a helper function to add errors to context
func HandleError(c *gin.Context, err error) {
	c.Error(err)
}

// HandleAppError is a helper function to add app errors to context
func HandleAppError(c *gin.Context, appErr *errors.AppError) {
	c.Error(appErr)
}
