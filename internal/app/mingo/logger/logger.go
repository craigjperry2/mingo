package logger

import (
	"io"
	"log"

	"github.com/craigjperry2/mingo/internal/app/mingo/system"
)

func Setup(loggingDestination io.Writer, clock system.Clock, hostname string) {
	log.SetFlags(0)
	log.SetOutput(componentLogger{clock, loggingDestination, hostname, "main"})
}
