package logger

import (
	"bytes"
	"log"
	"testing"

	"github.com/craigjperry2/mingo/internal/app/mingo/system"
)

func TestMainLoggingFormat(t *testing.T) {
	var expected = "2022-04-30T23:59:59Z | testhost.logger | main | Testing logging format\n"
	var buf bytes.Buffer

	Setup(&buf, system.ClockForTesting("2022-04-30T23:59:59Z"), "testhost.logger")

	log.Println("Testing logging format")

	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}
