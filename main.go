package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
)

func main() {
	// With thanks to the *awesome* Eli Bendersky: https://eli.thegreenplace.net/2020/testing-flag-parsing-in-go-programs/
	message, err := parseFlags(os.Args[0], os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Fprint(os.Stderr, message)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", message)
		os.Exit(2)
	}
	fmt.Println("This does nothing yet. But it did nothing successfully 👍")
}

// Yak-shaving: make cmd line argument parsing testable
func parseFlags(progname string, args []string) (output string, err error) {
	progname = path.Base(progname)
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	// Yak-shaving: Show -h option in usage help
	flags.Usage = func() { usageHelpMessage(progname, flags.Output) }

	var buf bytes.Buffer
	flags.SetOutput(&buf)

	err = flags.Parse(args)

	return buf.String(), err
}

// Yak-shaving: Avoid my flags.Usage override in parseFlags() defeating my flags.SetOutput()
type writerProvider func() io.Writer

func usageHelpMessage(progname string, wp writerProvider) {
	template := `Usage: %s [OPTION]

Options:
  -h, --help	This help message
`
	fmt.Fprintf(wp(), template, progname)
}
