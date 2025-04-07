package logger

import (
    "fmt"
    "log"
    "os"
)

// Log levels
const (
	DebugLevel = "DEBUG"
    InfoLevel  = "INFO"
    WarnLevel  = "WARN"
    ErrorLevel = "ERROR"
)

// Custom logger struct
type Logger struct {
    logger *log.Logger
}

// NewLogger creates a new instance of Logger
func NewLogger() *Logger {
    return &Logger{
        logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
    }
}

// Info logs an informational message
func (l *Logger) Info(message string) {
    l.logger.Printf("[%s] %s", InfoLevel, message)
}

// Infof logs an informational formatted message
func (l *Logger) Infof(format string, v ...interface{}) {
    l.logger.Printf("[%s] %s", InfoLevel, fmt.Sprintf(format, v...))
}

// Infoln logs an informational message with a line break
func (l *Logger) Infoln(v ...interface{}) {
    l.logger.Printf("[%s] %s", InfoLevel, fmt.Sprintln(v...))
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
    l.logger.Printf("[%s] %s", WarnLevel, message)
}

// Warnf logs a warning formatted message
func (l *Logger) Warnf(format string, v ...interface{}) {
    l.logger.Printf("[%s] %s", WarnLevel, fmt.Sprintf(format, v...))
}

// Warnln logs a warning message with a line break
func (l *Logger) Warnln(v ...interface{}) {
    l.logger.Printf("[%s] %s", WarnLevel, fmt.Sprintln(v...))
}

// Error logs an error message
func (l *Logger) Error(message string) {
    l.logger.Printf("[%s] %s", ErrorLevel, message)
}

// Errorf logs an error formatted message
func (l *Logger) Errorf(format string, v ...interface{}) {
    l.logger.Printf("[%s] %s", ErrorLevel, fmt.Sprintf(format, v...))
}

// Errorln logs an error message with a line break
func (l *Logger) Errorln(v ...interface{}) {
    l.logger.Printf("[%s] %s", ErrorLevel, fmt.Sprintln(v...))
}

// Debugf logs a debug formatted message
func (l *Logger) Debugf(format string, v ...interface{}) {
    l.logger.Printf("[%s] %s", DebugLevel, fmt.Sprintf(format, v...))
}

// Debugln logs a debug message with a line break
func (l *Logger) Debugln(v ...interface{}) {
    l.logger.Printf("[%s] %s", DebugLevel, fmt.Sprintln(v...))
}