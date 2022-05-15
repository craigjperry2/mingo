package logger

import (
	"bytes"
	"testing"

	"github.com/craigjperry2/mingo/internal/app/mingo/system"
)

func TestLoggingFormat(t *testing.T) {
	var expected = "2022-04-30T23:59:59Z | testhost | testcomponent | Testing logging format\n"
	var buf bytes.Buffer

	l := NewComponentLogger(system.ClockForTesting("2022-04-30T23:59:59Z"), &buf, "testhost", "testcomponent")
	l.Println("Testing logging format")

	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}
