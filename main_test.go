package main

import (
	"flag"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	var tests = []struct {
		progname string
		args     []string
		message  string
		err      error
	}{
		{"testprog", []string{}, "", nil},
		{"testprog", []string{"-h"}, "Usage: testprog [OPTION]\n\nOptions:\n  -h, --help\tThis help message\n", flag.ErrHelp},
		{"testprog", []string{"--help"}, "Usage: testprog [OPTION]\n\nOptions:\n  -h, --help\tThis help message\n", flag.ErrHelp},
		{"/path/to/testprog", []string{"--help"}, "Usage: testprog [OPTION]\n\nOptions:\n  -h, --help\tThis help message\n", flag.ErrHelp},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			message, err := parseFlags(tt.progname, tt.args)
			if err != tt.err {
				t.Errorf("err got %v, want %v", err, tt.err)
			}
			if message != tt.message {
				t.Errorf("message got %q, want %q", message, tt.message)
			}
		})
	}
}
