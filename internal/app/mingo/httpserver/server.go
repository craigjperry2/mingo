package httpserver

import (
	"net/http"
	"time"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
	"github.com/craigjperry2/mingo/internal/app/mingo/httpserver/handlers"
	"github.com/craigjperry2/mingo/internal/app/mingo/httpserver/middleware"
	"github.com/craigjperry2/mingo/internal/app/mingo/logger"
)

// Configure an HTTP server with routes, handlers, middleware & graceful shutdown ability
// with thanks to https://gist.github.com/creack/4c00ee404f2d7bd5983382cc93af5147
func MakeHttpServer() (*http.Server) {

	router := http.NewServeMux()
	router.Handle("/", handlers.NewIndexHandler())
	router.Handle("/health", handlers.NewHealthHandler())
	router.Handle("/static/", handlers.NewStaticHandler("/static/"))
	router.Handle("/crud", handlers.NewCrudHandler())
	router.Handle("/edit", handlers.NewEditHandler())

	server := &http.Server{
		Addr: "0.0.0.0:" + config.GetInstance().GetListenPortStr(), // TODO: IPv6 controls
		Handler: (middleware.Middlewares{
			middleware.NewTracingMiddleware(middleware.NewIdFountain()),
			middleware.NewLoggingMiddleware(),
		}).Apply(router),
		ErrorLog:     logger.NewComponentLogger(config.GetInstance().GetClock(), config.GetInstance().GetLoggingDestination(), config.GetInstance().GetHostname(), "error"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return server
}
