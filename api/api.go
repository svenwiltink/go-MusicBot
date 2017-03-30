package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
	"log"
	"net/http"
)

type API struct {
	Router   *mux.Router
	playlist playlist.ListInterface
	Routes   []Route
}

func NewAPI(playlist playlist.ListInterface) *API {
	return &API{
		Router:   mux.NewRouter(),
		playlist: playlist,
	}
}

func (api *API) Start() {
	api.initializeRoutes()

	// Register all routes
	for _, r := range api.Routes {
		api.registerRoute(r)
	}

	log.Fatal(http.ListenAndServe(":7070", api.Router))
}

func (api *API) initializeRoutes() {
	api.Routes = []Route{
		{
			Pattern: "/list",
			Method:  http.MethodGet,
			handler: api.ListHandler,
		}, {
			Pattern: "/current",
			Method:  http.MethodGet,
			handler: api.CurrentHandler,
		}, {
			Pattern: "/play",
			Method:  http.MethodGet,
			handler: api.PlayHandler,
		}, {
			Pattern: "/pause",
			Method:  http.MethodGet,
			handler: api.PauseHandler,
		}, {
			Pattern: "/stop",
			Method:  http.MethodGet,
			handler: api.StopHandler,
		}, {
			Pattern: "/next",
			Method:  http.MethodGet,
			handler: api.NextHandler,
		}, {
			Pattern: "/add",
			Method:  http.MethodGet,
			handler: api.AddHandler,
		},
	}
}

// registerRoute - Register a rout with the
func (api *API) registerRoute(route Route) bool {
	api.Router.HandleFunc(route.Pattern, route.handler).Methods(route.Method)

	return true
}

func (api *API) ListHandler(w http.ResponseWriter, r *http.Request) {
	items := api.playlist.GetItems()

	err := json.NewEncoder(w).Encode(api.convertItems(items))
	if err != nil {
		fmt.Printf("API list error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) CurrentHandler(w http.ResponseWriter, r *http.Request) {
	itm := api.playlist.GetCurrentItem()
	var item *player.ListItem
	if itm != nil {
		item = itm.(*player.ListItem)
	}

	err := json.NewEncoder(w).Encode(item)
	if err != nil {
		fmt.Printf("API current error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) PlayHandler(w http.ResponseWriter, r *http.Request) {
	err := api.playlist.Play()
	if err != nil {
		fmt.Printf("API play error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) PauseHandler(w http.ResponseWriter, r *http.Request) {
	err := api.playlist.Pause()
	if err != nil {
		fmt.Printf("API pause error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) StopHandler(w http.ResponseWriter, r *http.Request) {
	err := api.playlist.Stop()
	if err != nil {
		fmt.Printf("API stop error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) NextHandler(w http.ResponseWriter, r *http.Request) {
	itm, err := api.playlist.Next()
	if err != nil {
		fmt.Printf("API next error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var item *player.ListItem
	if itm != nil {
		item = itm.(*player.ListItem)
	}

	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		fmt.Printf("API next error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) AddHandler(w http.ResponseWriter, r *http.Request) {
	items, err := api.playlist.AddItems(r.URL.Query().Get("url"))
	if err != nil {
		fmt.Printf("API add error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(api.convertItems(items))
	if err != nil {
		fmt.Printf("API add error (2): %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) convertItems(itms []playlist.ItemInterface) (newItems []player.ListItem) {
	for _, itm := range itms {
		newItems = append(newItems, *itm.(*player.ListItem))
	}
	return
}
