package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/svenwiltink/go-musicbot/config"
	"github.com/svenwiltink/go-musicbot/player"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
)

type API struct {
	config     *config.API
	router     *mux.Router
	player     player.MusicPlayer
	routes     []Route
	wsUpgrader *websocket.Upgrader
}

type Context string

const (
	CONTEXT_AUTHENTICATED Context = "IS_AUTHENTICATED"
	CONTEXT_USERNAME      Context = "USERNAME"
)

func NewAPI(conf *config.API, player player.MusicPlayer) *API {
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
		player: player,
	}
}

func (api *API) Start() (err error) {
	api.initializeRoutes()

	// Register all routes
	for _, r := range api.routes {
		api.registerRoute(r)
	}

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", api.config.Host, api.config.Port), api.router)
	logrus.Errorf("API.Start: Error serving API http server on %s:%d: %v", api.config.Host, api.config.Port, err)
	return
}

func (api *API) authenticator(inner http.HandlerFunc, optional bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated := false

		username, password, _ := r.BasicAuth()
		if api.config.Username == username && api.config.Password == password {
			authenticated = true
		}

		if !optional && !authenticated {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"MusicBot\"")
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, CONTEXT_AUTHENTICATED, authenticated)
		ctx = context.WithValue(ctx, CONTEXT_USERNAME, api.config.Username)
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
			Pattern: "/statistics",
			Method:  http.MethodGet,
			handler: api.StatisticsHandler,
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
			Pattern: "/previous",
			Method:  http.MethodGet,
			handler: api.authenticator(api.PreviousHandler, false),
		}, {
			Pattern: "/jump",
			Method:  http.MethodGet,
			handler: api.authenticator(api.JumpHandler, false),
		}, {
			Pattern: "/add",
			Method:  http.MethodGet,
			handler: api.authenticator(api.AddHandler, false),
		}, {
			Pattern: "/open",
			Method:  http.MethodGet,
			handler: api.authenticator(api.OpenHandler, false),
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
	song, remaining := api.player.GetCurrent()

	s := Status{
		Status:  api.player.GetStatus(),
		Current: getAPISong(song, remaining),
		List:    getAPISongs(api.player.GetQueue()),
	}
	err := json.NewEncoder(w).Encode(s)
	if err != nil {
		logrus.Errorf("API.StatusHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) ListHandler(w http.ResponseWriter, r *http.Request) {
	songs := api.player.GetQueue()
	err := json.NewEncoder(w).Encode(getAPISongs(songs))
	if err != nil {
		logrus.Errorf("API.ListHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) CurrentHandler(w http.ResponseWriter, r *http.Request) {
	song, remaining := api.player.GetCurrent()
	err := json.NewEncoder(w).Encode(getAPISong(song, remaining))
	if err != nil {
		logrus.Errorf("API.CurrentHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	stats := api.player.GetStatistics()
	err := json.NewEncoder(w).Encode(stats)
	if err != nil {
		logrus.Errorf("API.StatisticsHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) PlayHandler(w http.ResponseWriter, r *http.Request) {
	song, err := api.player.Play()
	if err != nil {
		logrus.Errorf("API.PlayHandler: Error playing: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(getAPISong(song, song.GetDuration()))
	if err != nil {
		logrus.Errorf("API.PlayHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) PauseHandler(w http.ResponseWriter, r *http.Request) {
	err := api.player.Pause()
	if err != nil {
		logrus.Errorf("API.PauseHandler: Error pausing: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) StopHandler(w http.ResponseWriter, r *http.Request) {
	err := api.player.Stop()
	if err != nil {
		logrus.Errorf("API.StopHandler: Error stopping: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) NextHandler(w http.ResponseWriter, r *http.Request) {
	song, err := api.player.Next()
	if err != nil {
		logrus.Errorf("API.NextHandler: Error next-ing: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(getAPISong(song, song.GetDuration()))
	if err != nil {
		logrus.Errorf("API.NextHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) PreviousHandler(w http.ResponseWriter, r *http.Request) {
	song, err := api.player.Previous()
	if err != nil {
		logrus.Errorf("API.PreviousHandler: Error previous-ing: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(getAPISong(song, song.GetDuration()))
	if err != nil {
		logrus.Errorf("API.PreviousHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) JumpHandler(w http.ResponseWriter, r *http.Request) {
	index, err := strconv.ParseInt(r.URL.Query().Get("index"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	song, err := api.player.Jump(int(index))
	if err != nil {
		logrus.Errorf("API.JumpHandler: Error jumping: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(getAPISong(song, song.GetDuration()))
	if err != nil {
		logrus.Errorf("API.JumpHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) AddHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	user := r.Context().Value(CONTEXT_USERNAME).(string)
	songs, err := api.player.Add(url, fmt.Sprintf("WebAPI_%s", user))
	if err != nil {
		logrus.Errorf("API.AddHandler: Error adding [%s] %v", url, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(getAPISongs(songs))
	if err != nil {
		logrus.Errorf("API.AddHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) OpenHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	user := r.Context().Value(CONTEXT_USERNAME).(string)
	songs, err := api.player.Insert(url, 0, fmt.Sprintf("WebAPI_%s", user))
	if err != nil {
		logrus.Errorf("API.OpenHandler: Error inserting [%s] %v", url, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(getAPISongs(songs))
	if err != nil {
		logrus.Errorf("API.OpenHandler: Json encode error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) SocketHandler(w http.ResponseWriter, r *http.Request) {
	readOnly := !r.Context().Value(CONTEXT_AUTHENTICATED).(bool)

	ws, err := api.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("API.SocketHandler: Error upgrading to socket: %v", err)
		return
	}

	user, _, _ := r.BasicAuth()
	logrus.Infof("API.SocketHandler: Opening new socket [ReadOnly: %v | User: %s]", readOnly, user)
	cws := NewControlWebsocket(ws, readOnly, api.player, fmt.Sprintf("WebAPI_%s", user))
	cws.Start()
}
