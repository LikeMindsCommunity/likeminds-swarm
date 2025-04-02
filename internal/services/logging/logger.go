package logging

import (
	coralogix "github.com/coralogix/go-coralogix-sdk"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/sirupsen/logrus"
)

var (
	log                 		*logrus.Logger
	coralogixHookExists 		bool
	appInsightsInitialized		bool
	appInsightsClient			appinsights.TelemetryClient
)

func initializeApplicationInsights() {
	instrumentationKey := environment.GoDotEnvVariable("AZURE_APPINSIGHTS_INSTRUMENTATION_KEY")

	if len(instrumentationKey) == 0 {
		log.Error("Invalid Application Insights Instrumentation Key, Cannot start Application Insights Logger")
		return
	}

	appInsightsClient = appinsights.NewTelemetryClient(instrumentationKey)
	appInsightsInitialized = true
}

func GetAppInsightsClient() appinsights.TelemetryClient {
	return appInsightsClient
}

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

	if !appInsightsInitialized {
		initializeApplicationInsights()
	}
	
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
