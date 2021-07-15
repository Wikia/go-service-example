package logging

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type key int

const loggerIDKey key = 119

// EchoLogger is a middleware and zap to provide an "access log" like logging for each request.
func EchoLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()

			var traceID string

			if span := opentracing.SpanFromContext(req.Context()); span != nil {
				if sc, ok := span.Context().(jaeger.SpanContext); ok {
					traceID = sc.TraceID().String()
				}
			}

			var logger = log
			if traceID != "" {
				logger = log.With(zap.String("trace_id", traceID))
			}

			c.SetRequest(req.WithContext(addLoggerToContext(req.Context(), logger)))

			err := next(c)

			if err != nil {
				c.Error(err)
				logger = logger.With(zap.Error(err))
			}

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

			if traceID != "" {
				fields = append(fields, zap.String("trace_id", traceID))
			}

			n := res.Status

			switch {
			case n >= http.StatusInternalServerError:
				logger.Error("Server error", fields...)
			case n >= http.StatusBadRequest:
				logger.Warn("Client error", fields...)
			case n >= http.StatusMultipleChoices:
				logger.Info("Redirection", fields...)
			default:
				logger.Info("Success", fields...)
			}

			return nil
		}
	}
}

func FromEchoContext(ctx echo.Context) *zap.Logger {
	return FromContext(ctx.Request().Context())
}

// FromRequest will return current logger embedded in the given request object
func FromRequest(r *http.Request) *zap.Logger {
	return FromContext(r.Context())
}

// FromContext will return current logger from the given context.Context object
func FromContext(ctx context.Context) *zap.Logger {
	logger := ctx.Value(loggerIDKey)
	if logger == nil {
		return nil
	}

	return logger.(*zap.Logger)
}

func AddToContext(c echo.Context, logger *zap.Logger) {
	c.SetRequest(c.Request().WithContext(addLoggerToContext(c.Request().Context(), logger)))
}

// addLoggerToContext adds given logger to the context.Context and returns new context
func addLoggerToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerIDKey, logger)
}
