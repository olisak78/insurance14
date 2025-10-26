package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus for structured logging with context support
type Logger struct {
	*logrus.Entry
}

// New creates a new logger
func New() *Logger {
	return &Logger{
		Entry: logrus.NewEntry(logrus.StandardLogger()),
	}
}

// WithContext creates a logger with user context information
func WithContext(ctx context.Context) *Logger {
	logger := New()

	// Extract user information from context
	if email, ok := ctx.Value("email").(string); ok && email != "" {
		logger.Entry = logger.Entry.WithField("user", email)
	} else if username, ok := ctx.Value("username").(string); ok && username != "" {
		logger.Entry = logger.Entry.WithField("user", username)
	} else if user, ok := ctx.Value("user").(string); ok && user != "" {
		logger.Entry = logger.Entry.WithField("user", user)
	} else {
		logger.Entry = logger.Entry.WithField("user", "unknown")
	}

	return logger
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Entry: l.Entry.WithField(key, value),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{
		Entry: l.Entry.WithFields(fields),
	}
}

// Debug logs a debug message (only shown when LOG_LEVEL=debug)
func (l *Logger) Debug(args ...interface{}) {
	l.Entry.Debug(args...)
}

// Debugf logs a formatted debug message (only shown when LOG_LEVEL=debug)
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Entry.Debugf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.Entry.Info(args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Entry.Infof(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.Entry.Warn(args...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Entry.Warnf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.Entry.Error(args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Entry.Errorf(format, args...)
}
