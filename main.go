package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
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

type Config struct {
	startUtc   time.Time
	progname   string
	args       []string
	username   string
	hostname   string
	listenPort uint16
}

func main() {

	// --- 1. Bootstrap App Config ----------------------------------------

	config := &Config{
		startUtc:   time.Now().UTC(),
		progname:   path.Base(os.Args[0]),
		args:       os.Args[1:],
		username:   mustUsername(),
		hostname:   mustHostname(),
		listenPort: 8080,
	}
	mustHandleCliOverrides(config)
	setupLogging(config)
	serverLogger := newComponentLogger(config, "server")

	// --- 2. Setup Web Server --------------------------------------------

	serverLogger.Println(config.progname, "is starting as user", config.username, "on", config.listenPort)
	ctx, server := NewService(config, serverLogger)

	// --- 3. Enter Main Loop and Serve Clients Until SIGINT or SIGTERM Received

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		serverLogger.Fatalf("Could not listen on %q: %s\n", config.listenPort, err)
	}
	<-ctx.Done()
	serverLogger.Printf("Server has shutdown\n")
}

func mustUsername() string {
	user, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: unable to determine username")
		os.Exit(3)
	}
	return user.Username
}

func mustHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: unable to determine hostname")
		os.Exit(4)
	}
	return hostname
}

func mustHandleCliOverrides(config *Config) {
	// with thanks to the *awesome* Eli Bendersky: https://eli.thegreenplace.net/2020/testing-flag-parsing-in-go-programs/
	message, err := parseFlags(config)
	if err == flag.ErrHelp {
		fmt.Fprint(os.Stderr, message)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", message)
		os.Exit(2)
	}
}

func parseFlags(config *Config) (string, error) {
	flags := flag.NewFlagSet(config.progname, flag.ContinueOnError)

	var buf bytes.Buffer
	flags.SetOutput(&buf)

	flags.Usage = func() { usageHelpMessage(config.progname, flags.Output()) }

	port := NewPortVar(&config.listenPort)
	flags.Var(port, "port", "port to listen on for webserver")
	flags.Var(port, "p", "port to listen on for webserver")

	err := flags.Parse(config.args)
	return buf.String(), err
}

func usageHelpMessage(progname string, w io.Writer) {
	template := `Usage: %s [OPTION]

Options:
  -h, --help	This help message
  -p, --port	port to listen on for webserver
`
	fmt.Fprintf(w, template, progname)
}

// I wanted to capture a port's constraints in the type system, with thanks to https://blog.gopheracademy.com/advent-2019/flags/
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

	const minPort, maxPort = 1, 65535
	if val < minPort || val > maxPort {
		return fmt.Errorf("port %d out of range [%d:%d]", val, minPort, maxPort)
	}

	*p.port = uint16(val)
	return nil
}

type componentLogger struct {
	hostname  string
	component string
}

func newComponentLogger(conf *Config, component string) *log.Logger {
	logger := log.New(componentLogger{conf.hostname, component}, "", 0)
	return logger
}

func (writer componentLogger) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05.999Z"), " | ", writer.hostname, " | ", writer.component, " | ", string(bytes))
}

func setupLogging(conf *Config) {
	log.SetFlags(0)
	log.SetOutput(componentLogger{conf.hostname, "main"})
}

// with thanks to https://gist.github.com/creack/4c00ee404f2d7bd5983382cc93af5147
type service struct {
	logger        *log.Logger
	nextRequestID func() string
	healthy       int64 // TODO: DRY violation, we hsve 2 sources of truth for "start time"
}

func NewService(config *Config, serverLogger *log.Logger) (context.Context, *http.Server) {
	s := &service{
		logger: newComponentLogger(config, "access"),
		nextRequestID: func() string {
			return strconv.FormatInt(time.Now().UnixNano(), 36)
		},
	}

	router := http.NewServeMux()
	router.HandleFunc("/", s.index)
	router.HandleFunc("/healthz", s.healthz)

	server := &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(int(config.listenPort)), // TODO: IPv6
		Handler:      (middlewares{s.tracing, s.logging}).apply(router),
		ErrorLog:     serverLogger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	ctx := s.shutdown(context.Background(), server)

	atomic.StoreInt64(&s.healthy, time.Now().UnixNano())
	return ctx, server
}

type middleware func(http.Handler) http.Handler
type middlewares []middleware

func (mws middlewares) apply(hdlr http.Handler) http.Handler {
	if len(mws) == 0 {
		return hdlr
	}
	return mws[1:].apply(mws[0](hdlr))
}

func (s *service) shutdown(ctx context.Context, server *http.Server) context.Context {
	ctx, done := context.WithCancel(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer done()

		<-quit
		signal.Stop(quit)
		close(quit)

		atomic.StoreInt64(&s.healthy, 0)
		server.ErrorLog.Println("Server is shutting down")

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			server.ErrorLog.Fatalf("Could not gracefully shutdown the server: %s\n", err)
		}
	}()

	return ctx
}

func (s *service) index(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	fmt.Fprintf(w, "<html><h1>Web Server</h1></html>\n")
}

func (s *service) healthz(w http.ResponseWriter, req *http.Request) {
	if h := atomic.LoadInt64(&s.healthy); h == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		fmt.Fprintf(w, "uptime: %s\n", time.Since(time.Unix(0, h)))
	}
}

func (s *service) logging(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		lrw := NewLoggingResponseWriter(w)
		hdlr.ServeHTTP(lrw, req)
		func(start time.Time, statusCode int) {
			requestID := w.Header().Get("X-Request-Id")
			if requestID == "" {
				requestID = "unknown"
			}
			s.logger.Println(requestID, req.Method, statusCode, req.URL.Path, req.RemoteAddr, req.UserAgent(), time.Since(start))
		}(time.Now(), lrw.statusCode)
	})
}

func (s *service) tracing(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = s.nextRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		hdlr.ServeHTTP(w, req)
	})
}

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
