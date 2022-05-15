package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
)

// Expose read-only server health status
type HealthHandler struct{}

func NewHealthHandler() HealthHandler {
	return HealthHandler{}
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// TODO: JSONify
	fmt.Fprintf(w, "uptime: %s\n", config.GetInstance().GetClock()().UTC().Sub(time.Unix(0, config.GetInstance().GetStartUtc().UnixNano())))
}
