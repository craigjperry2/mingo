package config

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/craigjperry2/mingo/internal/app/mingo/system"
)

func TestFlags(t *testing.T) {
	const expectedHelpText = "Usage: testprog [OPTION]\n\nOptions:\n -d, --dir <dir>\toverride files embedded in binary and serve /static/*\n \t\t\turls from disk\n -h, --help\t\tthis help message\n -p, --port <port>\tport to listen on for webserver\n"
	const expectedFlagError = "flag: help requested"
	const missingPortArg = "flag needs an argument: -port"
	const missingDirArg = "flag needs an argument: -d"
	const portErrorTemplate = "invalid value \"%d\" for flag -port: port %d out of range [1:65535]"
	var unexpectedPort0Error = fmt.Sprintf(portErrorTemplate, 0, 0)
	var unexpectedPort65536Error = fmt.Sprintf(portErrorTemplate, 65536, 65536)
	var loggingBuf bytes.Buffer

	var tests = []struct {
		config          *Config
		expectedLogging string
		expectedErrStr  string
	}{
		{makeConfig([]string{}, 0, &loggingBuf, ""), "", ""},
		{makeConfig([]string{"-h"}, 0, &loggingBuf, ""), expectedHelpText, expectedFlagError},
		{makeConfig([]string{"--help"}, 0, &loggingBuf, ""), expectedHelpText, expectedFlagError},
		{makeConfig([]string{"-p", "1234"}, 1234, &loggingBuf, ""), "", ""},
		{makeConfig([]string{"--port", "1234"}, 1234, &loggingBuf, ""), "", ""},
		{makeConfig([]string{"--port", "0"}, 0, &loggingBuf, ""), unexpectedPort0Error + "\n" + expectedHelpText, unexpectedPort0Error},
		{makeConfig([]string{"--port", "1"}, 1, &loggingBuf, ""), "", ""},
		{makeConfig([]string{"--port", "65535"}, 65535, &loggingBuf, ""), "", ""},
		{makeConfig([]string{"--port", "65536"}, 0, &loggingBuf, ""), unexpectedPort65536Error + "\n" + expectedHelpText, unexpectedPort65536Error},
		{makeConfig([]string{"--port"}, 0, &loggingBuf, ""), missingPortArg + "\n" + expectedHelpText, missingPortArg},
		{makeConfig([]string{"-d"}, 0, &loggingBuf, ""), missingDirArg + "\n" + expectedHelpText, missingDirArg},
		{makeConfig([]string{"--dir", "does-not-exist"}, 0, &loggingBuf, "does-not-exist"), "", ""},
	}

	for _, tt := range tests {
		buf := tt.config.loggingDestination.(*bytes.Buffer)
		buf.Reset()

		t.Run(strings.Join(tt.config.args, " "), func(t *testing.T) {
			config, err := parseFlags(tt.config)
			if err != nil && err.Error() != tt.expectedErrStr {
				t.Errorf("err got %v, want %v", err, tt.expectedErrStr)
			}
			if buf.String() != tt.expectedLogging {
				t.Errorf("message got %q, want %q", buf.String(), tt.expectedLogging)
			}
			fmt.Println("conf", config)
			if !reflect.DeepEqual(config, tt.config) {
				t.Errorf("conf got %+v, want %+v", config, tt.config)
			}
		})
	}
}

func makeConfig(cli []string, port uint16, logDest *bytes.Buffer, staticDir string) *Config {
	return &Config{"testprog", time.Time{}, cli, "", "", port, logDest, staticDir, system.ClockForTesting("2022-04-30T23:59:59Z"), nil}
}
