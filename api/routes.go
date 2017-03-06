package api

import "net/http"

type Route struct {
	Pattern string
	Method  string
	handler http.HandlerFunc
}
