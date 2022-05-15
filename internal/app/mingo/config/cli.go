package config

import (
	"flag"
	"fmt"
	"io"
	"strconv"
)

func parseFlags(config *Config) (*Config, error) {
	flags := flag.NewFlagSet(config.progname, flag.ContinueOnError)

	flags.SetOutput(config.loggingDestination)

	flags.Usage = func() { usageHelpMessage(config.progname, flags.Output()) }

	// Duplicated flags to achieve GNU-like command line syntax
	flags.StringVar(&config.staticDir, "d", "", "override files embedded in binary and serve /static/* urls from disk")
	flags.StringVar(&config.staticDir, "dir", "", "override files embedded in binary and serve /static/* urls from disk")

	port := &portVar{&config.listenPort}
	flags.Var(port, "port", "port to listen on for webserver")
	flags.Var(port, "p", "port to listen on for webserver")

	err := flags.Parse(config.args)
	return config, err
}

// Default usage neglects help flag and uses -flag rather than --flag or -f
func usageHelpMessage(progname string, w io.Writer) {
	// TODO: append options based on defined flags in order
	template := `Usage: %s [OPTION]

Options:
 -d, --dir <dir>	override files embedded in binary and serve /static/*
 			urls from disk
 -h, --help		this help message
 -p, --port <port>	port to listen on for webserver
`
	fmt.Fprintf(w, template, progname)
}

// Capture port flag's constraint in the type system, with thanks to https://blog.gopheracademy.com/advent-2019/flags/
type portVar struct {
	port *uint16
}

func (p *portVar) String() string {
	if p.port == nil {
		return ""
	}

	return fmt.Sprintf("%d", *p.port)
}

func (p *portVar) Set(s string) error {
	val, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	// NB: Port 0 could be valid in testing since it means use a random free port
	const minPort, maxPort = 1, 65535
	if val < minPort || val > maxPort {
		return fmt.Errorf("port %d out of range [%d:%d]", val, minPort, maxPort)
	}

	*p.port = uint16(val)
	return nil
}
