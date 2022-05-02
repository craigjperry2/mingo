package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

// --- Main Entry Point -------------------------------------------------------

type clock func() time.Time

// TODO: Ideally this should be an immutable singleton
type Config struct {
	startUtc           time.Time
	progname           string
	args               []string
	username           string
	hostname           string
	listenPort         uint16
	loggingDestination io.Writer
	health             int64
	clock              clock
	staticDir          string
}

//go:embed static
var static embed.FS

const staticMount = "/static/"

// Bootstrap the environment, assemble the server, then hand over to the http main loop
func main() {
	config := &Config{
		startUtc:           time.Now().UTC(),
		progname:           path.Base(os.Args[0]),
		args:               os.Args[1:],
		username:           mustUsername(),
		hostname:           mustHostname(),
		listenPort:         8080,
		loggingDestination: os.Stderr,
		health:             0,
		clock:              time.Now,
		staticDir:          "",
	}
	mustHandleCliOverrides(config)

	setupLogging(config.loggingDestination, config.hostname)

	ctx, server := makeHttpServer(config)
	runServerMainLoop(config, ctx, server)
}

// --- Config Helper Utilities ------------------------------------------------

// Will exit the process on error rather than returning
func mustUsername() string {
	user, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: unable to determine username")
		os.Exit(3)
	}
	return user.Username
}

// Will exit the process on error rather than returning
func mustHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: unable to determine hostname")
		os.Exit(4)
	}
	return hostname
}

// --- CLI Parsing & Handling -------------------------------------------------

// Will exit the process on error rather than returning
func mustHandleCliOverrides(config *Config) {
	// with thanks to the *awesome* Eli Bendersky: https://eli.thegreenplace.net/2020/testing-flag-parsing-in-go-programs/
	err := parseFlags(config)
	if err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		os.Exit(2)
	}
}

// GNU-style CLI flags & make testing easy
func parseFlags(config *Config) error {
	flags := flag.NewFlagSet(config.progname, flag.ContinueOnError)

	flags.SetOutput(config.loggingDestination)

	flags.Usage = func() { usageHelpMessage(config.progname, flags.Output()) }

	// Duplicated flags to achieve GNU-like command line syntax
	flags.StringVar(&config.staticDir, "d", "", "override files embedded in binary and serve /static/* urls from disk")
	flags.StringVar(&config.staticDir, "dir", "", "override files embedded in binary and serve /static/* urls from disk")

	port := NewPortVar(&config.listenPort)
	flags.Var(port, "port", "port to listen on for webserver")
	flags.Var(port, "p", "port to listen on for webserver")

	return flags.Parse(config.args)
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

func NewPortVar(port *uint16) *portVar {
	return &portVar{port}
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

// --- Logging ----------------------------------------------------------------

// I want logging in the format of "<ISO8601 date/time> | <hostname> | <component> | <message...>"
type componentLogger struct {
	clock     clock
	w         io.Writer
	hostname  string
	component string
}

// TODO: suspect this is not idiomatic Go. This New* func is returning a *log.Logger not a *componentLogger
func NewComponentLogger(clock clock, loggingDestination io.Writer, hostname string, component string) *log.Logger {
	logger := log.New(componentLogger{clock, loggingDestination, hostname, component}, "", 0)
	return logger
}

func (logger componentLogger) Write(bytes []byte) (int, error) {
	return fmt.Fprint(logger.w, logger.clock().UTC().Format("2006-01-02T15:04:05.999Z"), " | ", logger.hostname, " | ", logger.component, " | ", string(bytes))
}

func setupLogging(loggingDestination io.Writer, hostname string) {
	log.SetFlags(0)
	log.SetOutput(componentLogger{time.Now, loggingDestination, hostname, "main"})
}

// I want to log the HTTP Status code of each request
// with thanks to https://gist.github.com/Boerworz/b683e46ae0761056a636
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// --- Assemble the HTTP Server -----------------------------------------------

// Configure an HTTP server with routes, handlers, middleware & graceful shutdown ability
// with thanks to https://gist.github.com/creack/4c00ee404f2d7bd5983382cc93af5147
func makeHttpServer(config *Config) (context.Context, *http.Server) {

	router := http.NewServeMux()
	router.HandleFunc("/", index)
	router.HandleFunc("/health", makeHealthHandler(&config.health, config.clock))
	router.Handle(staticMount, makeStaticHandler(config.staticDir))

	server := &http.Server{
		Addr: "0.0.0.0:" + strconv.Itoa(int(config.listenPort)), // TODO: IPv6 controls
		Handler: (middlewares{
			makeTracingMiddleware(makeIdFountain(config.clock)),
			makeLoggingMiddleware(config.clock, config.loggingDestination, config.hostname),
		}).apply(router),
		ErrorLog:     NewComponentLogger(config.clock, config.loggingDestination, config.hostname, "error"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	ctx := spawnGracefulShutdownReceiver(context.Background(), server, config)

	atomic.StoreInt64(&config.health, config.startUtc.UnixNano())

	return ctx, server
}

// When Ctrl+C, cause the server to stop accepting new requests, finish any existing requests then terminate
func spawnGracefulShutdownReceiver(ctx context.Context, server *http.Server, config *Config) context.Context {
	ctx, done := context.WithCancel(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer done()

		<-quit
		signal.Stop(quit)
		close(quit)

		atomic.StoreInt64(&config.health, 0)
		server.ErrorLog.Println("Server is shutting down")

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			server.ErrorLog.Printf("Could not gracefully shutdown the server: %s\n", err)
			os.Exit(6)
		}
	}()

	return ctx
}

// --- HTTP Request Handlers --------------------------------------------------

// Handle requests to the server's root
func index(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	fmt.Fprintf(w, "<html><h1>Web Server</h1></html>\n")
}

// Handle requests for files (.js, .css) from static dir
func makeStaticHandler(staticDir string) http.Handler {
	if staticDir == "" {
		fSys, err := fs.Sub(static, ".")
		if err != nil {
			panic(err)
		}
		return http.FileServer(http.FS(fSys))
	} else {
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			panic("dir doesn't exist: " + staticDir)
		}
		return http.StripPrefix(staticMount, http.FileServer(http.Dir(staticDir)))
	}
}

// Expose read-only server health status
func makeHealthHandler(health *int64, clock clock) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if h := atomic.LoadInt64(health); h == 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			// TODO: JSONify
			fmt.Fprintf(w, "uptime: %s\n", clock().UTC().Sub(time.Unix(0, h)))
		}
	}
}

// --- HTTP Request Middleware ------------------------------------------------

type middleware func(http.Handler) http.Handler
type middlewares []middleware

// Wrap handler with each listed middleware, in order
func (mws middlewares) apply(hdlr http.Handler) http.Handler {
	if len(mws) == 0 {
		return hdlr
	}
	return mws[1:].apply(mws[0](hdlr))
}

// A middleware than can log accesses
func makeLoggingMiddleware(clock clock, loggingDestination io.Writer, hostname string) middleware {
	accessLogger := NewComponentLogger(clock, loggingDestination, hostname, "access")
	return func(hdlr http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := clock().UTC()
			lrw := NewLoggingResponseWriter(w)
			hdlr.ServeHTTP(lrw, req)
			requestID := w.Header().Get("X-Request-Id")
			if requestID == "" {
				requestID = "unknown"
			}
			accessLogger.Println(requestID, req.Method, lrw.statusCode, req.URL.Path, req.RemoteAddr, req.UserAgent(), clock().UTC().Sub(start))
		})
	}
}

// Facilitate tracing requests by assigning unique id's where they don't exist already
type idFountain func() string

func makeIdFountain(clock clock) idFountain {
	return func() string {
		return strconv.FormatInt(clock().UTC().UnixNano(), 36)
	}
}

func makeTracingMiddleware(nextRequestID idFountain) middleware {
	return func(hdlr http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			w.Header().Set("X-Request-Id", requestID)
			hdlr.ServeHTTP(w, req)
		})
	}
}

// --- Run the HTTP Server ----------------------------------------------------

// Invoke the HTTP server main loop then await graceful shutdown
func runServerMainLoop(config *Config, ctx context.Context, httpServer *http.Server) {
	serviceLogger := NewComponentLogger(config.clock, config.loggingDestination, config.hostname, "service")
	serviceLogger.Println(config.progname, "is starting as user", config.username, "on", config.listenPort)

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		serviceLogger.Printf("Could not listen on %d: %v\n", config.listenPort, err)
		os.Exit(5)
	}
	<-ctx.Done()
	serviceLogger.Printf("Server has shutdown\n")
}
