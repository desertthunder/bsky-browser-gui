package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// LogService provides logging functionality via Wails bindings
type LogService struct {
	ctx        context.Context
	mu         sync.RWMutex
	entries    []LogEntry
	maxEntries int
	file       *os.File
	writer     *LogWriter
	level      LogLevel
}

// LogWriter implements io.Writer and emits Wails events
type LogWriter struct {
	service *LogService
	mu      sync.Mutex
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the log level from the line
		level := w.parseLevel(line)

		entry := LogEntry{
			Level:     string(level),
			Message:   line,
			Timestamp: time.Now(),
		}

		w.service.addEntry(entry)
		w.service.emitLogLine(entry)
	}

	return len(p), nil
}

func (w *LogWriter) parseLevel(line string) LogLevel {
	upper := strings.ToUpper(line)
	if strings.Contains(upper, "ERROR") || strings.Contains(upper, "ERR") {
		return ErrorLevel
	}
	if strings.Contains(upper, "WARN") || strings.Contains(upper, "WARNING") {
		return WarnLevel
	}
	if strings.Contains(upper, "DEBUG") || strings.Contains(upper, "DBG") {
		return DebugLevel
	}
	return InfoLevel
}

// NewLogService creates a new LogService instance
func NewLogService() *LogService {
	return &LogService{
		entries:    make([]LogEntry, 0),
		maxEntries: 1000,
		level:      InfoLevel,
	}
}

// SetContext sets the Wails context for event emission
func (s *LogService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// Initialize sets up the log service with a file writer
func (s *LogService) Initialize() error {
	// Open log file
	logPath := s.getLogPath()
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	s.file = file

	// Create the log writer
	s.writer = &LogWriter{service: s}

	return nil
}

// Close closes the log file
func (s *LogService) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// GetWriter returns the io.Writer for logging
func (s *LogService) GetWriter() io.Writer {
	if s.writer == nil {
		return io.Discard
	}
	return s.writer
}

// GetMultiWriter returns a MultiWriter that writes to both the log file and the event emitter
func (s *LogService) GetMultiWriter() io.Writer {
	if s.file == nil || s.writer == nil {
		return io.Discard
	}
	return io.MultiWriter(s.file, s.writer)
}

// GetEntries returns all log entries
func (s *LogService) GetEntries() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]LogEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

// GetEntriesByLevel returns log entries filtered by level
func (s *LogService) GetEntriesByLevel(level string) []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []LogEntry
	for _, entry := range s.entries {
		if entry.Level == level {
			result = append(result, entry)
		}
	}
	return result
}

// Clear clears all log entries
func (s *LogService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make([]LogEntry, 0)
	s.emitLogCleared()
}

// SetLevel sets the minimum log level
func (s *LogService) SetLevel(level string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.level = LogLevel(level)
}

// GetLevel returns the current log level
func (s *LogService) GetLevel() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return string(s.level)
}

func (s *LogService) addEntry(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only add if level is >= current level
	if !s.shouldLog(entry.Level) {
		return
	}

	s.entries = append(s.entries, entry)

	// Trim if exceeding max entries
	if len(s.entries) > s.maxEntries {
		s.entries = s.entries[len(s.entries)-s.maxEntries:]
	}
}

func (s *LogService) shouldLog(entryLevel string) bool {
	levels := map[LogLevel]int{
		DebugLevel: 0,
		InfoLevel:  1,
		WarnLevel:  2,
		ErrorLevel: 3,
	}

	entryIdx := levels[LogLevel(entryLevel)]
	currentIdx := levels[s.level]

	return entryIdx >= currentIdx
}

func (s *LogService) emitLogLine(entry LogEntry) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "log:line", entry)
	}
}

func (s *LogService) emitLogCleared() {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "log:cleared", map[string]any{})
	}
}

func (s *LogService) getLogPath() string {
	if path := os.Getenv("BSKY_BROWSER_LOG"); path != "" {
		return path
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}

	appDir := filepath.Join(configDir, "bsky-browser", "logs")
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(appDir, fmt.Sprintf("bsky-browser_%s.log", timestamp))
}

// BufferedLogWriter wraps a writer with buffering for better performance
type BufferedLogWriter struct {
	writer  *bufio.Writer
	service *LogService
}

func NewBufferedLogWriter(service *LogService) *BufferedLogWriter {
	return &BufferedLogWriter{
		writer:  bufio.NewWriter(service.GetMultiWriter()),
		service: service,
	}
}

func (w *BufferedLogWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *BufferedLogWriter) Flush() error {
	return w.writer.Flush()
}
