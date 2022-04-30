package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
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
		{"testprog", []string{"-h"}, "Usage: testprog [OPTION]\n\nOptions:\n  -h, --help\tThis help message\n  -p, --port\tport to listen on for webserver\n", flag.ErrHelp},
		{"testprog", []string{"--help"}, "Usage: testprog [OPTION]\n\nOptions:\n  -h, --help\tThis help message\n  -p, --port\tport to listen on for webserver\n", flag.ErrHelp},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			config := &Config{
				progname: tt.progname,
				args: tt.args,
			}
			message, err := parseFlags(config)
			if err != tt.err {
				t.Errorf("err got %v, want %v", err, tt.err)
			}
			if message != tt.message {
				t.Errorf("message got %q, want %q", message, tt.message)
			}
		})
	}
}

func TestWebIndexPage(t *testing.T) {
	t.Run("Returns index page", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		s := new(service)
		s.index(response, request)

		got := response.Body.String()
		want := "<html><h1>Web Server</h1></html>\n"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

// TODO: Integration test for the service stack that asserts about:
//		* Configurable listen port (test this by assigning a random free port to make concurrent testing possible)
//		* /health endpoint behaviour including http status code and mime type
// 		* Status logging interceptor behaviour
//		* Clean shutdown behaviour
