package internal

import (
	"fmt"
	"io"
	"os"
)

// Logger provides styled logging for CLI user feedback.
type Logger struct {
	Out io.Writer
	Err io.Writer
}

// DefaultLogger writes to os.Stdout and os.Stderr.
var DefaultLogger = &Logger{
	Out: os.Stdout,
	Err: os.Stderr,
}

// Info prints an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Fprintf(l.Out, "[INFO] "+format+"\n", args...)
}

// Success prints a success message.
func (l *Logger) Success(format string, args ...interface{}) {
	fmt.Fprintf(l.Out, "[OK] "+format+"\n", args...)
}

// Warn prints a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	fmt.Fprintf(l.Err, "[WARN] "+format+"\n", args...)
}

// Error prints an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(l.Err, "[ERROR] "+format+"\n", args...)
}

// Global convenience functions using DefaultLogger
func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

func Success(format string, args ...interface{}) {
	DefaultLogger.Success(format, args...)
}

func Warn(format string, args ...interface{}) {
	DefaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}
