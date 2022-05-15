package handlers

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
	"github.com/craigjperry2/mingo/web"
)

// Handle requests for files (.js, .css) from static dir
type StaticHandler struct {
	handler http.Handler
}

func NewStaticHandler(staticMount string) StaticHandler {
	staticDir := config.GetInstance().GetStaticDir()
	if staticDir == "" {
		fSys, err := fs.Sub(web.StaticDir, ".")
		if err != nil {
			panic(err)
		}
		return StaticHandler{http.FileServer(http.FS(fSys))}
	} else {
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			panic("dir doesn't exist: " + staticDir)
		}
		return StaticHandler{http.StripPrefix(staticMount, http.FileServer(http.Dir(staticDir)))}
	}
}

func (h StaticHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.handler.ServeHTTP(w, req)
}
