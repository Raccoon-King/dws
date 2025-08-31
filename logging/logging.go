package logging

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var currentLevel Level = LevelInfo
var jsonFormat bool

// Init configures the logger according to environment variables.
// LOGGING: stdout | stderr | file
// LOG_LEVEL: debug | info | warn | error
// LOG_FORMAT: json | text (default text)
func Init() {
	switch os.Getenv("LOGGING") {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "file":
		f, err := os.OpenFile("dws.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		log.SetOutput(f)
	default:
		log.SetOutput(os.Stderr)
	}

	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		currentLevel = LevelDebug
	case "info":
		currentLevel = LevelInfo
	case "warn":
		currentLevel = LevelWarn
	case "error":
		currentLevel = LevelError
	default:
		currentLevel = LevelInfo
	}

	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "json" {
		jsonFormat = true
	}
}

func logf(level Level, levelStr, msg string, fields map[string]any) {
	if level < currentLevel {
		return
	}
	if jsonFormat {
		m := map[string]any{"level": levelStr, "msg": msg}
		for k, v := range fields {
			m[k] = v
		}
		b, _ := json.Marshal(m)
		log.Print(string(b))
		return
	}
	if len(fields) > 0 {
		b, _ := json.Marshal(fields)
		log.Printf("[%s] %s %s", strings.ToUpper(levelStr), msg, b)
	} else {
		log.Printf("[%s] %s", strings.ToUpper(levelStr), msg)
	}
}

// Debug logs a message at debug level.
func Debug(msg string, fields map[string]any) { logf(LevelDebug, "debug", msg, fields) }

// Info logs a message at info level.
func Info(msg string, fields map[string]any) { logf(LevelInfo, "info", msg, fields) }

// Warn logs a message at warn level.
func Warn(msg string, fields map[string]any) { logf(LevelWarn, "warn", msg, fields) }

// Error logs a message at error level.
func Error(msg string, fields map[string]any) { logf(LevelError, "error", msg, fields) }

// LevelEnabled returns true if the given level would be logged.
func LevelEnabled(level Level) bool { return level >= currentLevel }
