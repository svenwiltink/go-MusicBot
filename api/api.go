package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
	"log"
	"net/http"
	"time"
)

type API struct {
	config     *config.API
	router     *mux.Router
	playlist   playlist.ListInterface
	routes     []Route
	wsUpgrader *websocket.Upgrader
}

type Item struct {
	Title            string
	Seconds          int
	SecondsRemaining int
	FormattedTime    string
	URL              string
}

type Status struct {
	Status  playlist.Status
	Current *Item
	List    []Item
}

func NewAPI(conf *config.API, playl playlist.ListInterface) *API {
	return &API{
		config: conf,
		router: mux.NewRouter(),
		wsUpgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// For the musicbot we are not gonna care
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		playlist: playl,
	}
}

func (api *API) Start() {
	api.initializeRoutes()

	// Register all routes
	for _, r := range api.routes {
		api.registerRoute(r)
	}

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", api.config.Host, api.config.Port), api.router)
	log.Fatal(err)
}

func (api *API) authenticator(inner http.HandlerFunc, optional bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated := true

		username, password, _ := r.BasicAuth()
		if api.config.Username != username || api.config.Password != password {
			authenticated = false
		}

		if !optional && !authenticated {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"MusicBot\"")
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "isAuthenticated", authenticated)
		ctx = context.WithValue(ctx, "user", api.config.Username)
		inner.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (api *API) initializeRoutes() {
	api.routes = []Route{
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
			handler: api.authenticator(api.PlayHandler, false),
		}, {
			Pattern: "/pause",
			Method:  http.MethodGet,
			handler: api.authenticator(api.PauseHandler, false),
		}, {
			Pattern: "/stop",
			Method:  http.MethodGet,
			handler: api.authenticator(api.StopHandler, false),
		}, {
			Pattern: "/next",
			Method:  http.MethodGet,
			handler: api.authenticator(api.NextHandler, false),
		}, {
			Pattern: "/add",
			Method:  http.MethodGet,
			handler: api.authenticator(api.AddHandler, false),
		}, {
			Pattern: "/socket",
			Method:  http.MethodGet,
			handler: api.authenticator(api.SocketHandler, true),
		},
	}
}

// registerRoute - Register a rout with the
func (api *API) registerRoute(route Route) bool {
	api.router.HandleFunc(route.Pattern, route.handler).Methods(route.Method)

	return true
}

func (api *API) StatusHandler(w http.ResponseWriter, r *http.Request) {
	itm, remaining := api.playlist.GetCurrentItem()

	s := Status{
		Status:  api.playlist.GetStatus(),
		Current: api.convertItem(itm, remaining),
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
	itm, remaining := api.playlist.GetCurrentItem()
	err := json.NewEncoder(w).Encode(api.convertItem(itm, remaining))
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
	err = json.NewEncoder(w).Encode(api.convertItem(itm, itm.GetDuration()))
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

	err = json.NewEncoder(w).Encode(api.convertItem(itm, itm.GetDuration()))
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

func (api *API) SocketHandler(w http.ResponseWriter, r *http.Request) {
	readOnly := r.Context().Value("isAuthenticated").(bool)

	ws, err := api.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Sprintf("API socket error: %v", err)
		return
	}

	cws := NewControlWebsocket(ws, readOnly, api.playlist)
	cws.Start()
}

func (api *API) convertItem(itm playlist.ItemInterface, remaining time.Duration) (newItem *Item) {
	if itm != nil {
		duration := itm.GetDuration()
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) - (minutes * 60)

		newItem = &Item{
			Title:            itm.GetTitle(),
			URL:              itm.GetURL(),
			Seconds:          int(duration.Seconds()),
			SecondsRemaining: int(remaining.Seconds()),
			FormattedTime:    fmt.Sprintf("%d:%02d", minutes, seconds),
		}
	}
	return
}

func (api *API) convertItems(itms []playlist.ItemInterface) (newItems []Item) {
	for _, itm := range itms {
		if itm == nil {
			continue
		}
		newItems = append(newItems, *api.convertItem(itm, itm.GetDuration()))
	}
	return
}
