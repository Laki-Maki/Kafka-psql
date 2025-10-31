package logger

import "log"

type Logger interface {
    Info(msg string)
    Error(msg string, err error)
    Debug(msg string)
}

type StdLogger struct{}

func (l *StdLogger) Info(msg string) {
    log.Println("[INFO]", msg)
}

func (l *StdLogger) Error(msg string, err error) {
    log.Println("[ERROR]", msg, err)
}

func (l *StdLogger) Debug(msg string) {
    log.Println("[DEBUG]", msg)
}
