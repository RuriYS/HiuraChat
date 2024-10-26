package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	file      *os.File
	mu        sync.Mutex
	logLevel  LogLevel
	useColors bool
}

var instance *Logger
var once sync.Once

func GetLogger() *Logger {
    once.Do(func() {
        instance = NewLogger()
    })
    return instance
}

func NewLogger() *Logger {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		fmt.Printf("Error creating logs directory: %v\n", err)
		return nil
	}

	timestamp := time.Now().Format("2006-01-02_15-04")
	logPath := filepath.Join("logs", fmt.Sprintf("%s.log", timestamp))
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return nil
	}

	return &Logger{
		file:      file,
		logLevel:  INFO,
		useColors: true,
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.logLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("15:04:05.000")
	logMessage := strings.ReplaceAll(fmt.Sprintf(format, args...), "\n", "\\n")
	
	var colorCode string
	var levelStr string
	
	switch level {
	case DEBUG:
		colorCode = colorBlue
		levelStr = "DEBUG"
	case INFO:
		colorCode = colorGreen
		levelStr = "INFO"
	case WARN:
		colorCode = colorYellow
		levelStr = "WARN"
	case ERROR:
		colorCode = colorRed
		levelStr = "ERROR"
	}

	consoleLog := fmt.Sprintf("[%s] %s%s: %s%s\n", 
		timestamp, 
		colorCode,
		levelStr,
		logMessage,
		colorReset,
	)

	fileLog := fmt.Sprintf("[%s] %s: %s\n",
		timestamp,
		levelStr,
		logMessage,
	)

	if l.useColors {
		fmt.Print(consoleLog)
	} else {
		fmt.Print(fileLog)
	}

	if l.file != nil {
		_, err := l.file.WriteString(fileLog)
		if err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
		}
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) SetLogLevel(level LogLevel) {
	l.logLevel = level
}

func (l *Logger) SetUseColors(use bool) {
	l.useColors = use
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}