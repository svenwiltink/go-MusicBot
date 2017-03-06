package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"log"
	"net/http"
)

type API struct {
	Router *mux.Router
	Player player.MusicPlayer
	Routes []Route
}

func NewAPI(player player.MusicPlayer) *API {

	return &API{
		Router: mux.NewRouter(),
		Player: player,
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
		},
	}
}

// registerRoute - Register a rout with the
func (api *API) registerRoute(route Route) bool {

	api.Router.HandleFunc(route.Pattern, route.handler).Methods(route.Method)

	return true
}

func (api *API) ListHandler(w http.ResponseWriter, r *http.Request) {
	items := api.Player.GetQueueItems()

	err := json.NewEncoder(w).Encode(items)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
