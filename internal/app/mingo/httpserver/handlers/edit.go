package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
	"github.com/craigjperry2/mingo/internal/app/mingo/database"
)

// Handle CRUD requests to the Person resource
type EditHandler struct {
	db *database.Db
}

func NewEditHandler() EditHandler {
	return EditHandler{config.GetInstance().GetDatabase()}
}

func (h EditHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPut:
		if req.URL.Path != "/edit" {
			http.NotFound(w, req)
			return
		}
		id, err := strconv.Atoi(req.URL.Query().Get("id"))
		if err != nil {
			http.NotFound(w, req)
			return
		}
		req.ParseForm()
		p, _ := h.db.Update(id, req.FormValue("name"), req.FormValue("location"))
		fmt.Fprintf(w, `<tr> <td>%d</td> <td>%s</td> <td>%s</td> <td><div class="buttons are-small"><button class="button is-info" hx-get="/edit?id=%d">Edit</button><button class="button is-danger" hx-delete="/crud?id=%d">Delete</button></div></td> </tr>`, p.Id, p.Name, p.Location, p.Id, p.Id)
	case http.MethodPost:
		if req.URL.Path != "/edit" {
			http.NotFound(w, req)
			return
		}
		req.ParseForm()
		p, _ := h.db.Insert(req.FormValue("name"), req.FormValue("location"))
		fmt.Fprintf(w, `<tr hx-swap-oob="afterbegin:.tablebody" hx-swap="outerHTML"><td>%d</td> <td>%s</td> <td>%s</td> <td><div class="buttons are-small"><button class="button is-info" hx-get="/edit?id=%d">Edit</button><button class="button is-danger" hx-delete="/crud?id=%d">Delete</button></div></td> </tr> <tr> <td></td> <td><input name="name" placeholder="name"></td> <td><input name="location" placeholder="location"></td> <td><div class="buttons are-small"><button class="button is-info" hx-post="/edit" hx-include="closest tr" hx-target="closest tr" hx-swap="outerHTML">Add</button></div></td> </tr>`, p.Id, p.Name, p.Location, p.Id, p.Id)
	default: // GET
		if req.URL.Path != "/edit" {
			http.NotFound(w, req)
			return
		}
		id, err := strconv.Atoi(req.URL.Query().Get("id"))
		if err != nil {
			http.NotFound(w, req)
			return
		}
		row, _ := h.db.Get(id)
		fmt.Fprintf(w, `<tr> <td>%d</td> <td><input name='name' value='%s'></td> <td><input name='location' value='%s'></td> <td><div class="buttons are-small"><button class="button is-info">Cancel</button><button class="button is-danger" hx-put="/edit?id=%d" hx-include="closest tr">Save</button></div></td> </tr>`, row.Id, row.Name, row.Location, row.Id)
	}
}
