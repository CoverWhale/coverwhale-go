package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Level int

var (
	std = NewLogger()
)

type Logger struct {
	Level Level
	*log.Logger
	contextMessage string
}

const (
	ErrorLevel = 0
	InfoLevel  = 1
	DebugLevel = 2
)

func getLevel(l string) Level {
	switch strings.ToLower(l) {
	case "error":
		return ErrorLevel
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	default:
		return InfoLevel
	}
}

func NewLogger() *Logger {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		level = os.Getenv("LOG_LEVEL")
	}

	l := getLevel(level)
	logger := log.Default()
	return &Logger{
		l,
		logger,
		"",
	}
}

func (l *Logger) log(lvl Level, s interface{}) {
	if lvl <= l.Level {
		if l.contextMessage != "" {
			l.Printf("%s [%s]\n", s, l.contextMessage)
		} else {
			l.Println(s)
		}
	}

}

func (l *Logger) clone() *Logger {
	copy := *l
	return &copy
}

func (l *Logger) WithContext(s string) *Logger {
	c := l.clone()
	c.contextMessage = s

	return c

}

func (l *Logger) Error(s interface{}) {
	m := fmt.Sprintf("[ERROR] %s", s)
	l.log(ErrorLevel, m)
}

func (l *Logger) Errorf(format string, s ...interface{}) {
	f := fmt.Sprintf("[ERROR] %s", format)
	m := fmt.Sprintf(f, s...)
	l.log(ErrorLevel, m)
}

func (l *Logger) Info(s interface{}) {
	m := fmt.Sprintf("[INFO] %s", s)
	l.log(InfoLevel, m)
}

func (l *Logger) Infof(format string, s ...interface{}) {
	f := fmt.Sprintf("[INFO] %s", format)
	m := fmt.Sprintf(f, s...)
	l.log(InfoLevel, m)
}

func (l *Logger) Debug(s interface{}) {
	m := fmt.Sprintf("[DEBUG] %s", s)
	l.log(DebugLevel, m)
}

func (l *Logger) Debugf(format string, s ...interface{}) {
	f := fmt.Sprintf("[DEBUG] %s", format)
	m := fmt.Sprintf(f, s...)
	l.log(DebugLevel, m)
}

func (l *Logger) Fatal(s interface{}) {
	m := fmt.Sprintf("[FATAL] %s", s)
	l.Logger.Fatal(m)
}

func (l *Logger) Fatalf(format string, s ...interface{}) {
	f := fmt.Sprintf("[FATAL] %s", format)
	m := fmt.Sprintf(f, s...)
	l.Logger.Fatalf(format, m)
}

func Error(s interface{}) {
	std.Error(s)
}

func Errorf(format string, s ...interface{}) {
	std.Errorf(format, s...)
}

func Info(s interface{}) {
	std.Info(s)
}

func Infof(format string, s ...interface{}) {
	std.Infof(format, s...)
}

func Debug(s interface{}) {
	std.Debug(s)
}

func Debugf(format string, s ...interface{}) {
	std.Debugf(format, s...)
}

func Fatal(s interface{}) {
	std.Fatal(s)
}

func Fatalf(format string, s ...interface{}) {
	std.Fatalf(format, s...)
}
