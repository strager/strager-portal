package main

import (
	"net/http"
	"net/url"
)

type bangHandler func(response http.ResponseWriter, request *http.Request, query string)

var bangHandlers map[string]bangHandler = map[string]bangHandler{
	"": func(response http.ResponseWriter, request *http.Request, query string) {
		http.Redirect(response, request, "https://kagi.com/search?q="+url.QueryEscape(query), http.StatusFound)
	},
	"!go": func(response http.ResponseWriter, request *http.Request, query string) {
		http.Redirect(response, request, "https://pkg.go.dev/search?utm_source=godoc&q="+url.QueryEscape(query), http.StatusFound)
	},
}
