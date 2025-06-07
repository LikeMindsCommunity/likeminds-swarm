package logging

import (
	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger
)

func init() {
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.JSONFormatter{})

}

// Trace
func Trace(v ...interface{}) {
	log.Trace(v...)
}

// Debug
func Debug(v ...interface{}) {
	log.Debug(v...)
}

// Info
func Info(v ...interface{}) {
	log.Info(v...)
}
func InfoWithFields(data map[string]interface{}) {
	log.WithFields(logrus.Fields(data)).Info()
}

// Warn
func Warn(v ...interface{}) {
	log.Warn(v...)
}

// Error
func Error(v ...interface{}) {
	log.Error(v...)
}
func ErrorWithFields(data map[string]interface{}) {
	log.WithFields(logrus.Fields(data)).Error()
}

// Fatal
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Panic
func Panic(v ...interface{}) {
	log.Panic(v...)
}
