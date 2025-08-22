package middleware

import (
	"bytes"
	"encoding/json"
	"flex-service/pkg/logger"
	"io"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = maskSensitiveData(string(bodyBytes))
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		c.Next()

		responseBody := blw.body.String()

		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if shouldLogRequestBody(c.Request.Method) && requestBody != "" {
			logFields = append(logFields, zap.String("request_body", requestBody))
		}

		if c.Writer.Status() >= 400 {
			if errorCode, errorMessage := extractErrorDetails(responseBody); errorCode != "" {
				logFields = append(logFields,
					zap.String("error_code", errorCode),
					zap.String("error_message", errorMessage),
				)
			}
		} else if c.Writer.Status() >= 200 && c.Writer.Status() < 400 {
			logFields = append(logFields, zap.String("response_body", responseBody))
		}

		logger.Info("HTTP Request", logFields...)
	}
}

func shouldLogRequestBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

func extractErrorDetails(responseBody string) (string, string) {
	var errorDetails map[string]interface{}
	if json.Unmarshal([]byte(responseBody), &errorDetails) == nil {
		if err, ok := errorDetails["error"].(map[string]interface{}); ok {
			code, _ := err["code"].(string)
			message, _ := err["message"].(string)
			return code, message
		}
	}
	return "", ""
}

func maskSensitiveData(body string) string {
	sensitiveFields := []string{}

	for _, field := range sensitiveFields {
		pattern := `"` + field + `"\s*:\s*"[^"]*"`
		re := regexp.MustCompile(`(?i)` + pattern)
		body = re.ReplaceAllString(body, `"`+field+`": "***MASKED***"`)
	}

	return body
}
