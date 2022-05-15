package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
	"github.com/craigjperry2/mingo/internal/app/mingo/database"
)

// Handle CRUD requests to the Person resource
type CrudHandler struct {
	db *database.DbNothingBurger
}

func NewCrudHandler() CrudHandler {
	return CrudHandler{config.GetInstance().GetDatabase()}
}

func (h CrudHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		if req.URL.Path != "/crud" {
			http.NotFound(w, req)
			return
		}
		id, err := strconv.Atoi(req.URL.Query().Get("id"))
		if err != nil {
			http.NotFound(w, req)
			return
		}
		h.db.Delete(id)
	case http.MethodPut:
		if req.URL.Path != "/crud" {
			http.NotFound(w, req)
			return
		}
	default: // GET
		if req.URL.Path != "/crud" {
			http.NotFound(w, req)
			return
		}
		offset, err := strconv.Atoi(req.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}
		limit, err := strconv.Atoi(req.URL.Query().Get("limit"))
		if err != nil || limit < 1 {
			limit = 1
		}
		all, _ := h.db.GetAll(offset, limit)
		fmt.Println("all", all)
		for _, p := range all {
			fmt.Println("Writing", p)
			fmt.Fprintf(w, `<tr> <td>%d</td> <td>%s</td> <td>%s</td> <td><div class="buttons are-small"><button class="button is-info" hx-get="/edit?id=%d">Edit</button><button class="button is-danger" hx-delete="/crud?id=%d">Delete</button></div></td> </tr>`, p.Id, p.Name, p.Location, p.Id, p.Id)
		}
		if limit == len(all) {
			fmt.Fprintf(w, `<tr id="replaceMe"> <td colspan="4" class="has-text-centered"> <button class="button is-link" hx-get="/crud?limit=%d&offset=%d" hx-target="#replaceMe" hx-swap="outerHTML" hx-confirm="unset"> Load More... <span class="htmx-indicator is-transparent"> <span class="icon-text"> <span class="icon"> <i class="fas fa-spinner"></i> </span> </span> </span> </button> </td> </tr>`, limit, all[len(all)-1].Id)
		}
	}
}
