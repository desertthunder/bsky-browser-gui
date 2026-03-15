package main

import (
	"fmt"
	"log"
	"os"
)

// AppLogger provides application logging that writes to both file and Wails events
type AppLogger struct {
	stdLogger *log.Logger
	service   *LogService
}

var appLogger *AppLogger

// InitLogger initializes the application logger with the log service
func InitLogger(service *LogService) {
	multiWriter := service.GetMultiWriter()
	appLogger = &AppLogger{
		stdLogger: log.New(multiWriter, "", log.LstdFlags|log.Lshortfile),
		service:   service,
	}

	log.SetOutput(multiWriter)
}

// GetLogger returns the global app logger
func GetLogger() *AppLogger {
	if appLogger == nil {
		return &AppLogger{
			stdLogger: log.New(os.Stdout, "", log.LstdFlags),
		}
	}
	return appLogger
}

// Debug logs a debug message
func (l *AppLogger) Debug(msg string) {
	l.stdLogger.Printf("[DEBUG] %s", msg)
}

// Debugf logs a formatted debug message
func (l *AppLogger) Debugf(format string, args ...interface{}) {
	l.stdLogger.Printf("[DEBUG] "+format, args...)
}

// Info logs an info message
func (l *AppLogger) Info(msg string) {
	l.stdLogger.Printf("[INFO] %s", msg)
}

// Infof logs a formatted info message
func (l *AppLogger) Infof(format string, args ...interface{}) {
	l.stdLogger.Printf("[INFO] "+format, args...)
}

// Warn logs a warning message
func (l *AppLogger) Warn(msg string) {
	l.stdLogger.Printf("[WARN] %s", msg)
}

// Warnf logs a formatted warning message
func (l *AppLogger) Warnf(format string, args ...interface{}) {
	l.stdLogger.Printf("[WARN] "+format, args...)
}

// Error logs an error message
func (l *AppLogger) Error(msg string) {
	l.stdLogger.Printf("[ERROR] %s", msg)
}

// Errorf logs a formatted error message
func (l *AppLogger) Errorf(format string, args ...interface{}) {
	l.stdLogger.Printf("[ERROR] "+format, args...)
}

// Fatal logs a fatal message and exits
func (l *AppLogger) Fatal(msg string) {
	l.stdLogger.Fatalf("[FATAL] %s", msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *AppLogger) Fatalf(format string, args ...interface{}) {
	l.stdLogger.Fatalf("[FATAL] "+format, args...)
}

// Convenience functions that use the global logger
func LogDebug(msg string) {
	GetLogger().Debug(msg)
}

func LogDebugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func LogInfo(msg string) {
	GetLogger().Info(msg)
}

func LogInfof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func LogWarn(msg string) {
	GetLogger().Warn(msg)
}

func LogWarnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

func LogError(msg string) {
	GetLogger().Error(msg)
}

func LogErrorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func LogFatal(msg string) {
	GetLogger().Fatal(msg)
}

func LogFatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// String returns a string representation of the logger
func (l *AppLogger) String() string {
	return fmt.Sprintf("AppLogger{initialized: %v}", l.service != nil)
}
