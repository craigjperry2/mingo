package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// --- Unit Tests -------------------------------------------------------------

func TestFlags(t *testing.T) {
	const expectedHelpText = "Usage: testprog [OPTION]\n\nOptions:\n  -h, --help\tThis help message\n  -p, --port\tport to listen on for webserver\n"
	const expectedFlagError = "flag: help requested"
	const portErrorTemplate = "invalid value \"%d\" for flag -port: port %d out of range [1:65535]"
	var unexpectedPort0Error = fmt.Sprintf(portErrorTemplate, 0, 0)
	var unexpectedPort65536Error = fmt.Sprintf(portErrorTemplate, 65536, 65536)
	var loggingBuf bytes.Buffer

	var tests = []struct {
		config          Config
		expectedLogging string
		expectedErrStr  string
	}{
		{makeConfig([]string{}, 0, &loggingBuf), "", ""},
		{makeConfig([]string{"-h"}, 0, &loggingBuf), expectedHelpText, expectedFlagError},
		{makeConfig([]string{"--help"}, 0, &loggingBuf), expectedHelpText, expectedFlagError},
		{makeConfig([]string{"-p", "1234"}, 1234, &loggingBuf), "", ""},
		{makeConfig([]string{"--port", "1234"}, 1234, &loggingBuf), "", ""},
		{makeConfig([]string{"--port", "0"}, 0, &loggingBuf), unexpectedPort0Error + "\n" + expectedHelpText, unexpectedPort0Error},
		{makeConfig([]string{"--port", "1"}, 1, &loggingBuf), "", ""},
		{makeConfig([]string{"--port", "65535"}, 65535, &loggingBuf), "", ""},
		{makeConfig([]string{"--port", "65536"}, 0, &loggingBuf), unexpectedPort65536Error + "\n" + expectedHelpText, unexpectedPort65536Error},
	}

	for _, tt := range tests {
		buf := tt.config.loggingDestination.(*bytes.Buffer)
		buf.Reset()

		t.Run(strings.Join(tt.config.args, " "), func(t *testing.T) {
			config := &Config{
				progname:           tt.config.progname,
				args:               tt.config.args,
				loggingDestination: buf,
			}
			err := parseFlags(config)
			if err != nil && err.Error() != tt.expectedErrStr {
				t.Errorf("err got %v, want %v", err, tt.expectedErrStr)
			}
			if buf.String() != tt.expectedLogging {
				t.Errorf("message got %q, want %q", buf.String(), tt.expectedLogging)
			}
			if !reflect.DeepEqual(*config, tt.config) {
				t.Errorf("conf got %+v, want %+v", *config, tt.config)
			}
		})
	}
}

func TestLoggingFormat(t *testing.T) {
	var expected = "2022-04-30T23:59:59Z | testhost | testcomponent | Testing logging format\n"
	var buf bytes.Buffer
	l := NewComponentLogger(clockForTesting("2022-04-30T23:59:59Z"), &buf, "testhost", "testcomponent")
	l.Println("Testing logging format")

	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestLoggingHttpStatusCodes(t *testing.T) {
	response := httptest.NewRecorder()
	lrw := NewLoggingResponseWriter(response)
	expected := 123

	lrw.WriteHeader(expected)

	if lrw.statusCode != expected {
		t.Errorf("got %d, want %d", lrw.statusCode, expected)
	}
}

func TestIndexHandler(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	index(response, request)

	got := response.Body.String()
	want := "<html><h1>Web Server</h1></html>\n"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHealthHandler(t *testing.T) {
	var tests = []struct {
		health           int64
		expectedResponse string
		expectedStatus   int
	}{
		{0, "", 503},
		{1651359658000000000, "uptime: 59m1s\n", 200},
		{1651350000000000000, "uptime: 3h39m59s\n", 200},
	}

	for _, tt := range tests {
		t.Run("Health: "+strconv.FormatInt(tt.health, 10), func(t *testing.T) {
			healthHandler := makeHealthHandler(&tt.health, clockForTesting("2022-04-30T23:59:59Z"))
			response := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/", nil)

			healthHandler(response, request)

			if response.Body.String() != tt.expectedResponse {
				t.Errorf("body got %q, want %q", response.Body.String(), tt.expectedResponse)
			}

			if response.Result().StatusCode != tt.expectedStatus {
				t.Errorf("status got %d, want %d", response.Result().StatusCode, tt.expectedStatus)
			}
		})
	}
}

func TestMiddlewareOrdering(t *testing.T) {
	var m middlewares
	m = append(m, makeMiddleware("inner"))
	m = append(m, makeMiddleware("outer"))
	r := m.apply(hFunc())
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	r.ServeHTTP(response, request)

	got := response.Body.String()
	want := "middleware-outer\nmiddleware-inner\nhandler-innermost\n"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var tests = []struct {
		requestId        string
		expectedResponse string
		expectedStatus   int
	}{
		{"", "2022-04-30T23:59:59Z | testhost | access | unknown GET 200 /   0s\n", 200},
		{"TEST-ID", "2022-04-30T23:59:59Z | testhost | access | TEST-ID GET 200 /   0s\n", 200},
	}

	for _, tt := range tests {
		t.Run(tt.requestId, func(t *testing.T) {
			var buf bytes.Buffer
			loggingMiddleware := makeLoggingMiddleware(clockForTesting("2022-04-30T23:59:59Z"), &buf, "testhost")(hFunc())
			response := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/", nil)
			response.Header().Set("X-Request-Id", tt.requestId)

			loggingMiddleware.ServeHTTP(response, request)

			if buf.String() != tt.expectedResponse {
				t.Errorf("got %q, want %q", buf.String(), tt.expectedResponse)
			}

			if response.Result().StatusCode != tt.expectedStatus {
				t.Errorf("status got %d, want %d", response.Result().StatusCode, tt.expectedStatus)
			}
		})
	}
}

func TestTracingMiddleware(t *testing.T) {
	var tests = []struct {
		clock            clock
		requestId        string
		expectedResponse string
		expectedStatus   int
	}{
		{clockForTesting("2022-04-30T23:59:58Z"), "", "cjnzec7mxi4g", 200},
		{clockForTesting("2022-04-30T23:59:59Z"), "", "cjnzeco6az28", 200},
		{clockForTesting("2022-04-30T23:59:58Z"), "TEST-ID", "TEST-ID", 200},
		{clockForTesting("2022-04-30T23:59:59Z"), "TEST-ID", "TEST-ID", 200},
	}

	for _, tt := range tests {
		t.Run(tt.requestId, func(t *testing.T) {
			tracingMiddleware := makeTracingMiddleware(makeIdFountain(tt.clock))(hFunc())
			response := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("X-Request-Id", tt.requestId)

			tracingMiddleware.ServeHTTP(response, request)

			if response.Result().Header.Get("X-Request-Id") != tt.expectedResponse {
				t.Errorf("got %q, want %q", response.Result().Header.Get("X-Request-Id"), tt.expectedResponse)
			}

			if response.Result().StatusCode != tt.expectedStatus {
				t.Errorf("status got %d, want %d", response.Result().StatusCode, tt.expectedStatus)
			}
		})
	}
}

// --- Integration Tests ------------------------------------------------------

// with thanks to https://medium.com/nerd-for-tech/setup-and-teardown-unit-test-in-go-bd6fa1b785cd
func setupIntegrationTestSuite(t *testing.T) func(t *testing.T) {
	// TODO: We don't have any logic for port clashes (common in CI environments), should retry on incrementing ports until it finds a free one but for now we just delegate that externally to the CI pipeline and it can pass us a known good free port (which is a solution with a race condition...)
	testServerPort := os.Getenv("TEST_SERVER_PORT")

	oldArgs := os.Args
	os.Args = []string{"mingo", "--port", testServerPort}

	fmt.Println("Spawning main() with cli:", os.Args)
	// TODO: Ideally spawn a new isolated process, not just a go routine for main() - but calling "go run ." via os Exec in a test actually runs something else not this project
	go main()
	pollHealthUntilReady(testServerPort)

	return func(t *testing.T) {
		fmt.Println("ITS: Cleaning up after test, shutting down port", testServerPort)
		os.Args = oldArgs
		// TODO: We don't cleanup & shutdown the server or reap the goroutine - the port is taken until testing process exits
	}
}

func TestWebServerIntegration(t *testing.T) {
	// with thanks to https://peter.bourgon.org/blog/2021/04/02/dont-use-build-tags-for-integration-tests.html
	testServerPort := os.Getenv("TEST_SERVER_PORT")
	if testServerPort == "" {
		t.Skip("set TEST_SERVER_PORT to run this test")
	}

	teardownSuite := setupIntegrationTestSuite(t)
	defer teardownSuite(t)

	var tests = []struct {
		path           string
		expectedError  error
		expectedStatus int
		expectedBody   string
	}{
		// TODO: There's no scenario to validate handling of SIGINT
		{"/", nil, http.StatusOK, "<html><h1>Web Server</h1></html>\n"},
		{"/non-existant", nil, http.StatusNotFound, "404 page not found\n"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {

			res, err := http.Get("http://127.0.0.1:" + testServerPort + tt.path)

			if err != nil {
				t.Errorf("err got %v, want %v", err, tt.expectedError)
			}

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("status got %d, want %d", res.StatusCode, tt.expectedStatus)
			}

			bodyBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected error when reading body %v", err)
			}
			body := string(bodyBytes)
			if body != tt.expectedBody {
				t.Errorf("body got %s, want %s", body, tt.expectedBody)
			}
		})
	}
}

// --- Testing Helpers --------------------------------------------------------

func makeConfig(cli []string, port uint16, logDest *bytes.Buffer) Config {
	return Config{time.Time{}, "testprog", cli, "", "", port, logDest, 0, nil}
}

func clockForTesting(timespec string) clock {
	return func() time.Time {
		instant, _ := time.Parse(time.RFC3339, timespec)
		return instant
	}
}

func makeMiddleware(order string) middleware {
	return func(hdlr http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("middleware-" + order + "\n"))
			hdlr.ServeHTTP(w, req)
		})
	}
}

func dummyHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "handler-innermost\n")
}

func hFunc() http.Handler {
	return http.HandlerFunc(dummyHandler)
}

// poll /health for up to 500ms waiting for http 200
func pollHealthUntilReady(testServerPort string) {
	for i := 50; i > 0; i -= 1 {
		fmt.Println("Testing for server readiness on port", testServerPort)
		res, err := http.Get("http://127.0.0.1:" + testServerPort + "/health")
		if err == nil { // NB: we're expecting an err to begin with
			if res.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	panic("HTTP Server Never Came Up")
}
