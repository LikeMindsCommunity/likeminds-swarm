package logging

type CustomLogger struct{}

func (l *CustomLogger) Info(args ...interface{}) {
	Info(args...)
}

func (l *CustomLogger) Debug(args ...interface{}) {
	Debug(args...)
}

func (l *CustomLogger) Fatal(args ...interface{}) {
	Fatal(args...)
}

func (l *CustomLogger) Panic(args ...interface{}) {
	Panic(args...)
}

func (l *CustomLogger) Error(args ...interface{}) {
	Error(args...)
}

func (l *CustomLogger) Warn(args ...interface{}) {
	Warn(args...)
}

// NewCustomLogger | Returns a new instance of the CustomLogger
func NewCustomLogger() *CustomLogger {
	return &CustomLogger{}
}
