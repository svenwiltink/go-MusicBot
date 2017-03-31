package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
	"log"
	"net/http"
)

type API struct {
	Configuration *config.API
	Router        *mux.Router
	playlist      playlist.ListInterface
	Routes        []Route
}

type Status struct {
	Status  playlist.Status
	Current *player.ListItem
	List    []player.ListItem
}

func NewAPI(conf *config.API, playl playlist.ListInterface) *API {
	return &API{
		Configuration: conf,
		Router:        mux.NewRouter(),
		playlist:      playl,
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

func (api *API) authenticator(inner http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()

		if api.Configuration.Username != username || api.Configuration.Password != password {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"MusicBot\"")
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		inner.ServeHTTP(w, r)
	}
}

func (api *API) initializeRoutes() {
	api.Routes = []Route{
		{
			Pattern: "/status",
			Method:  http.MethodGet,
			handler: api.StatusHandler,
		}, {
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
			handler: api.authenticator(api.PlayHandler),
		}, {
			Pattern: "/pause",
			Method:  http.MethodGet,
			handler: api.authenticator(api.PauseHandler),
		}, {
			Pattern: "/stop",
			Method:  http.MethodGet,
			handler: api.authenticator(api.StopHandler),
		}, {
			Pattern: "/next",
			Method:  http.MethodGet,
			handler: api.authenticator(api.NextHandler),
		}, {
			Pattern: "/add",
			Method:  http.MethodGet,
			handler: api.authenticator(api.AddHandler),
		},
	}
}

// registerRoute - Register a rout with the
func (api *API) registerRoute(route Route) bool {
	api.Router.HandleFunc(route.Pattern, route.handler).Methods(route.Method)

	return true
}

func (api *API) StatusHandler(w http.ResponseWriter, r *http.Request) {
	s := Status{
		Status:  api.playlist.GetStatus(),
		Current: api.convertItem(api.playlist.GetCurrentItem()),
		List:    api.convertItems(api.playlist.GetItems()),
	}
	err := json.NewEncoder(w).Encode(s)
	if err != nil {
		fmt.Printf("API status encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) ListHandler(w http.ResponseWriter, r *http.Request) {
	items := api.playlist.GetItems()
	err := json.NewEncoder(w).Encode(api.convertItems(items))
	if err != nil {
		fmt.Printf("API list encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) CurrentHandler(w http.ResponseWriter, r *http.Request) {
	itm := api.playlist.GetCurrentItem()
	err := json.NewEncoder(w).Encode(api.convertItem(itm))
	if err != nil {
		fmt.Printf("API current encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) PlayHandler(w http.ResponseWriter, r *http.Request) {
	itm, err := api.playlist.Play()
	if err != nil {
		fmt.Printf("API play error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(api.convertItem(itm))
	if err != nil {
		fmt.Printf("API next encode error: %v\n", err)
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

	err = json.NewEncoder(w).Encode(api.convertItem(itm))
	if err != nil {
		fmt.Printf("API next encode error: %v\n", err)
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
		fmt.Printf("API add encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) convertItem(itm playlist.ItemInterface) (newItem *player.ListItem) {
	if itm != nil {
		newItem = player.NewListItem(itm.GetTitle(), itm.GetDuration(), itm.GetURL())
	}
	return
}

func (api *API) convertItems(itms []playlist.ItemInterface) (newItems []player.ListItem) {
	for _, itm := range itms {
		if itm == nil {
			continue
		}
		newItems = append(newItems, *api.convertItem(itm))
	}
	return
}
