package logger

import (
	"fmt"
	"github.com/craigjperry2/mingo/internal/app/mingo/system"
	"io"
	"log"
)

// I want logging in the format of "<ISO8601 date/time> | <hostname> | <component> | <message...>"
type componentLogger struct {
	clock     system.Clock
	w         io.Writer
	hostname  string
	component string
}

// TODO: suspect this is not idiomatic Go. This New* func is returning a *log.Logger not a *componentLogger
func NewComponentLogger(clock system.Clock, loggingDestination io.Writer, hostname string, component string) *log.Logger {
	logger := log.New(componentLogger{clock, loggingDestination, hostname, component}, "", 0)
	return logger
}

func (logger componentLogger) Write(bytes []byte) (int, error) {
	return fmt.Fprint(logger.w, logger.clock().UTC().Format("2006-01-02T15:04:05.999Z"), " | ", logger.hostname, " | ", logger.component, " | ", string(bytes))
}
