package api

import "net/http"

type Route struct {
	Pattern string
	Method  string
	handler http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		Pattern: "list",
		Method:  http.MethodGet,
		handler: (*API).ListHandler,
	},
}
