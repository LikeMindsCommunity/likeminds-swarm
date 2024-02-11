package logging

import (
	coralogix "github.com/coralogix/go-coralogix-sdk"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/sirupsen/logrus"
)

var (
	log                 *logrus.Logger
	coralogixHookExists bool
)

func addCoralogixHook(log *logrus.Logger) {
	coralogixPrivateKey := environment.GoDotEnvVariable("CORALOGIX_PRIVATE_KEY")

	if len(coralogixPrivateKey) == 0 {
		log.Error("Invalid Coralogix Private Key, Cannot start Coralogix Logger")
		return
	}

	coralogixApplicationName := environment.GoDotEnvVariable("CORALOGIX_APPLICATION_NAME")

	if len(coralogixApplicationName) == 0 {
		log.Error("Invalid Coralogix Application Name, Cannot start Coralogix Logger")
		return
	}

	coralogixSystemName := environment.GoDotEnvVariable("CORALOGIX_SYSTEM_NAME")

	if len(coralogixSystemName) == 0 {
		log.Error("Invalid Coralogix System Name, Cannot start Coralogix Logger")
		return
	}

	CoralogixHook := coralogix.NewCoralogixHook(coralogixPrivateKey, coralogixApplicationName, coralogixSystemName)

	log.AddHook(CoralogixHook)
	coralogixHookExists = true
}

func init() {
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.JSONFormatter{})

	if !coralogixHookExists {
		addCoralogixHook(log)
	}
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

// Warn
func Warn(v ...interface{}) {
	log.Warn(v...)
}

// Error
func Error(v ...interface{}) {
	log.Error(v...)
}

// Fatal
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Panic
func Panic(v ...interface{}) {
	log.Panic(v...)
}
