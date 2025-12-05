package middleware

import (
	"bytes"
	"fmt"
	"io"
	"li-acc/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerConfig lets you control what to log
type LoggerConfig struct {
	LogRequestBody  bool
	LogResponseBody bool
	MaxBodySize     int // optional limit to avoid huge file logs
}

// bodyLogWriter captures response content
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write writes given bytes firstly into buffer body, and then into response
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware logs request and response details
func LoggingMiddleware(cfg LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body safely
		var reqBody []byte
		if cfg.LogRequestBody && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			if len(bodyBytes) > cfg.MaxBodySize && cfg.MaxBodySize > 0 {
				reqBody = []byte(fmt.Sprintf("[TRUNCATED %d bytes]", len(bodyBytes)))
			} else {
				reqBody = bodyBytes
			}
			// Restore the body for handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Wrap response writer to capture response output
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		path := c.Request.URL.Path

		c.Next()

		if !strings.Contains(path, "/api") {
			return // do not log non-API requests
		}

		// After handler runs
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		errorsString := c.Errors.String()

		// Prepare fields for zap Logger
		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency_ms", latency),
			zap.String("user_agent", userAgent),
			zap.Bool("is_error", len(c.Errors) > 0),
		}

		// Add optional bodies
		if cfg.LogRequestBody && len(reqBody) > 0 {
			fields = append(fields, zap.String("request_body", sanitizeBody(reqBody)))
		}
		if cfg.LogResponseBody && blw.body.Len() > 0 {
			respBodyStr := blw.body.String()
			if cfg.MaxBodySize > 0 && len(respBodyStr) > cfg.MaxBodySize {
				respBodyStr = "[TRUNCATED]"
			}
			fields = append(fields, zap.String("response_body", sanitizeBody([]byte(respBodyStr))))
		}
		if len(errorsString) > 0 {
			fields = append(fields, zap.String("errors", errorsString))
		}

		if statusCode >= 400 {
			logger.Error("HTTP request completed with error", fields...)
		} else {
			logger.Info("HTTP request completed", fields...)
		}
	}
}

// sanitizeBody shortens binary or unreadable data for safe logging
func sanitizeBody(b []byte) string {
	s := string(b)
	if strings.ContainsAny(s, "\x00\x01\x02\x03\x04\x05\x06") {
		return "[binary data omitted]"
	}
	return s
}
