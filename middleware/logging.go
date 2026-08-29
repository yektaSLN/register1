package middleware

import (
	"bytes"
	"context"
	"os"
	"time"

	"login/kafka"
	"login/logger"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware(
	kafkaProducer *kafka.Producer,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		duration := time.Since(start)

		var buffer bytes.Buffer

		log := logger.New(&buffer)

		event := log.Info()

		status := c.Writer.Status()

		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("type", kafka.EventHTTPRequest).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Int("status", status).
			Int64("latency_ms", duration.Milliseconds())

		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(uint); ok {
				event = event.Uint("user_id", id)
			}
		}

		if username, exists := c.Get("username"); exists {
			if name, ok := username.(string); ok {
				event = event.Str("username", name)
			}
		}

		event.Msg("http request")

		if err := kafkaProducer.PublishRaw(
			context.Background(),
			kafka.EventHTTPRequest,
			buffer.Bytes(),
		); err != nil {

			fileLogger := logger.New(os.Stderr)

			fileLogger.Error().
				Err(err).
				Msg("failed to queue http log")
		}
	}
}
