package orchestrator

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
	"github.com/craigjperry2/mingo/internal/app/mingo/errors"
	"github.com/craigjperry2/mingo/internal/app/mingo/httpserver"
	"github.com/craigjperry2/mingo/internal/app/mingo/logger"
)

func Orchestrate(args []string, stderr io.Writer) error {
	ctx, server, err := bootstrap(args, stderr)
	if err != nil {
		return err
	}
	attemptTransitionToRunning() // transition STARTING -> RUNNING
	return run(ctx, server)
}

// Bootstrap the app, triggers the following side-effects:
//	* Signal handler setup for SIGINT & SIGTERM to cause a graceful app shutdown
//	* App config will be initialised and made accessible as an immutable singleton via the config pkg
//	* Logging will be enabled
func bootstrap(args []string, stderr io.Writer) (context.Context, *http.Server, error) {
	err := config.Build(args, stderr)

	c := config.GetInstance()
	logger.Setup(c.GetLoggingDestination(), c.GetClock(), c.GetHostname())

	server := httpserver.MakeHttpServer()
	ctx := setupSignalHandler(context.Background(), server)

	return ctx, server, err
}

// Invoke the HTTP server main loop then await graceful shutdown
func run(ctx context.Context, server *http.Server) error {

	// TODO: Try listening on the port then open the browser to the location unless --no-browser, if already bound, just open browser

	serviceLogger := logger.NewComponentLogger(config.GetInstance().GetClock(), config.GetInstance().GetLoggingDestination(), config.GetInstance().GetHostname(), "service")
	serviceLogger.Println(config.GetInstance().GetProgname(), "is starting as user", config.GetInstance().GetUsername(), "on host", config.GetInstance().GetHostname(), "port", config.GetInstance().GetListenPortStr())

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		serviceLogger.Printf("Could not listen on port %d: %v\n", config.GetInstance().GetListenPort(), err)
		return errors.ErrPortUnavailable
	}
	<-ctx.Done()
	serviceLogger.Printf("Server has shutdown\n")
	return nil
}

func setupSignalHandler(ctx context.Context, server *http.Server) context.Context {
	ctx, done := context.WithCancel(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer done()

		<-quit
		signal.Stop(quit)
		close(quit)

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			server.ErrorLog.Printf("Could not gracefully shutdown the server: %s\n", err)
		}

		transitionToStopping() // transition (STARTING|RUNNING) -> STOPPING
	}()
	return ctx
}
