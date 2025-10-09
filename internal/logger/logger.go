package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"honey/internal/config"
)

// Logger interface pour le logging
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Close() error
}

// LogLevel représente le niveau de log
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String retourne la représentation string du niveau de log
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parse un string en LogLevel
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// SimpleLogger implémentation simple du logger
type SimpleLogger struct {
	level    LogLevel
	writer   io.Writer
	file     *os.File
}

// New crée un nouveau logger
func New(cfg config.LoggingConfig) (Logger, error) {
	level := ParseLogLevel(cfg.Level)
	
	// Créer le dossier de logs s'il n'existe pas
	if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Ouvrir le fichier de log
	file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Writer multi-sortie (console + fichier)
	writer := io.MultiWriter(os.Stdout, file)

	return &SimpleLogger{
		level:  level,
		writer: writer,
		file:   file,
	}, nil
}

// log écrit un message de log
func (l *SimpleLogger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := level.String()
	
	var formattedMsg string
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	} else {
		formattedMsg = msg
	}

	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, levelStr, formattedMsg)
	l.writer.Write([]byte(logEntry))
}

// Info log un message d'information
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warn log un message d'avertissement
func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

// Error log un message d'erreur
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

// Debug log un message de debug
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Fatal log un message fatal et termine le programme
func (l *SimpleLogger) Fatal(msg string, args ...interface{}) {
	l.log(FATAL, msg, args...)
	os.Exit(1)
}

// Infof log un message d'information formaté
func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warnf log un message d'avertissement formaté
func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Errorf log un message d'erreur formaté
func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Debugf log un message de debug formaté
func (l *SimpleLogger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Fatalf log un message fatal formaté et termine le programme
func (l *SimpleLogger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// Close ferme le logger
func (l *SimpleLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
