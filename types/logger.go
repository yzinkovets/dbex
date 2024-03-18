package types

type Logger interface {
	Trace(args ...interface{})
	Error(args ...interface{})
}

// defaultLogger is a logger that does nothing
type defaultLogger struct{}

func (this *defaultLogger) Trace(args ...interface{}) {}
func (this *defaultLogger) Error(args ...interface{}) {}

func NewDefaultLogger() Logger {
	return &defaultLogger{}
}
