package middleware

import (
	"net/http"
	"runtime/debug"

	"flex-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("stack", string(debug.Stack())),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Something went wrong",
		})
		c.Abort()
	})
}
