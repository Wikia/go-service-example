package logging

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// EchoLoggger is a middleware and zap to provide an "access log" like logging for each request.
func EchoLoggger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			var logger *zap.Logger
			if err != nil {
				c.Error(err)
				logger = log.With(zap.Error(err))
			} else {
				logger = log
			}

			req := c.Request()
			res := c.Response()

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("time", time.Since(start).String()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id != "" {
				id = res.Header().Get(echo.HeaderXRequestID)
				fields = append(fields, zap.String("request_id", id))
			}
			if span := opentracing.SpanFromContext(req.Context()); span != nil {
				if sc, ok := span.Context().(jaeger.SpanContext); ok {
					fields = append(fields, zap.String("trace_id", sc.TraceID().String()))
				}
			}

			n := res.Status
			switch {
			case n >= 500:
				logger.Error("Server error", fields...)
			case n >= 400:
				logger.Warn("Client error", fields...)
			case n >= 300:
				logger.Info("Redirection", fields...)
			default:
				logger.Info("Success", fields...)
			}

			return nil
		}
	}
}
