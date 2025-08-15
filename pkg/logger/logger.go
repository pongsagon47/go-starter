package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init(level, format string) error {
	var config zap.Config

	switch format {
	case "json":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Set output paths
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
	os.Exit(1)
}

func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Common field constants for consistent logging
const (
	FieldUserID      = "user_id"
	FieldRequestID   = "request_id"
	FieldTraceID     = "trace_id"
	FieldOperationID = "operation_id"
	FieldDuration    = "duration"
	FieldStatusCode  = "status_code"
	FieldMethod      = "method"
	FieldPath        = "path"
	FieldErrorCode   = "error_code"
	FieldClientIP    = "client_ip"
)

// Context-aware logging helpers
func WithContext(ctx context.Context) *zap.Logger {
	fields := []zap.Field{}

	if userID := getUserIDFromContext(ctx); userID != "" {
		fields = append(fields, zap.String(FieldUserID, userID))
	}

	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, zap.String(FieldRequestID, requestID))
	}

	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		fields = append(fields, zap.String(FieldTraceID, traceID))
	}

	return Logger.With(fields...)
}

// Helper functions to extract values from context
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

func getTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

// Structured logging helpers for common patterns
func LogHTTPRequest(method, path, clientIP string, statusCode int, duration interface{}) {
	Logger.Info("HTTP request",
		zap.String(FieldMethod, method),
		zap.String(FieldPath, path),
		zap.String(FieldClientIP, clientIP),
		zap.Int(FieldStatusCode, statusCode),
		zap.Any(FieldDuration, duration),
	)
}

func LogDBOperation(operation, table string, duration interface{}, err error) {
	if err != nil {
		Logger.Error("Database operation failed",
			zap.String("operation", operation),
			zap.String("table", table),
			zap.Any(FieldDuration, duration),
			zap.Error(err),
		)
	} else {
		Logger.Debug("Database operation completed",
			zap.String("operation", operation),
			zap.String("table", table),
			zap.Any(FieldDuration, duration),
		)
	}
}

func LogUserAction(userID, action string, success bool, details map[string]interface{}) {
	if success {
		Logger.Info("User action completed",
			zap.String(FieldUserID, userID),
			zap.String("action", action),
			zap.Any("details", details),
		)
	} else {
		Logger.Warn("User action failed",
			zap.String(FieldUserID, userID),
			zap.String("action", action),
			zap.Any("details", details),
		)
	}
}
