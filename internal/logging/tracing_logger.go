package logging

import (
	"go.uber.org/zap"
)

// TracingLogger this is the wrapper around zap.SugaredLogger that implements interface used by jaeger library
type TracingLogger struct {
	Logger *zap.SugaredLogger
}

// Error logs a message as error priority
func (l *TracingLogger) Error(msg string) {
	l.Logger.Errorf("ERROR: %s", msg)
}

// Infof logs a message at info priority
func (l *TracingLogger) Infof(msg string, args ...interface{}) {
	l.Logger.Infof(msg, args...)
}

// Debugf logs a message at debug priority
func (l *TracingLogger) Debugf(msg string, args ...interface{}) {
	l.Logger.Debugf(msg, args...)
}

func (l *TracingLogger) Print(args ...interface{}) {
	l.Logger.Info(args)
}
