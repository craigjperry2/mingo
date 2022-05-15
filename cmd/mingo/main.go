package main

import (
	"flag"
	"github.com/craigjperry2/mingo/internal/app/mingo/errors"
	"github.com/craigjperry2/mingo/internal/app/mingo/orchestrator"
	"os"
)

const (
	EXIT_BAD_FLAG = iota + 1
	EXIT_HELP
	EXIT_UNKNOWN_USER
	EXIT_UNKNOWN_HOST
	EXIT_PORT_UNAVAILABLE
	EXIT_HTTP_GRACEFUL_SHUTDOWN_FAILED
)

// A thin adapter between the operating system and this app, responsible for:
//	* collecting raw cli arguments
//	* providing std(in|out|err) file handles
//  	* invoking the app's bootstrap function
//	* returning an exit code to the OS on app termination
func main() {
	if err := orchestrator.Orchestrate(os.Args[1:], os.Stderr); err == flag.ErrHelp {
		os.Exit(EXIT_HELP)
	} else if err == errors.ErrUnknownUser {
		os.Exit(EXIT_UNKNOWN_USER)
	} else if err == errors.ErrUnknownHost {
		os.Exit(EXIT_UNKNOWN_HOST)
	} else if err == errors.ErrPortUnavailable {
		os.Exit(EXIT_PORT_UNAVAILABLE)
	} else {
		os.Exit(EXIT_BAD_FLAG)
	}
}
