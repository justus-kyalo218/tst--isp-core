package logger

import (
	"fmt"
	"log"
	"time"
)

type Level int

const (
	INFO Level = iota
	WARN
	ERROR
)

type Logger struct {
	level Level
}

var defaultLogger = &Logger{level: INFO}

func New(level Level) *Logger {
	return &Logger{level: level}
}

func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}
	levelStr := []string{"INFO", "WARN", "ERROR"}[level]
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(msg, args...)
	log.Printf("[%s] %s: %s", timestamp, levelStr, message)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

func SetLevel(level Level) {
	defaultLogger.level = level
}
