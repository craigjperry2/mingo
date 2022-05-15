package handlers

import (
	"fmt"
	"net/http"
)

type IndexHandler struct{}

func NewIndexHandler() IndexHandler {
	return IndexHandler{}
}

func (h IndexHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	fmt.Fprintf(w, "<html><h1>Web Server</h1><a href=\"static/\">HTMX Playground</a></html>\n")
}
