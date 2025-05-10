package models

import (
	"fmt"
	"net/http"

	"github.com/ServiceWeaver/weaver"
)

type Routes map[string]Route

type Route struct {
	ApiOnly        bool
	Titel          string
	Handler        func(w http.ResponseWriter, r *http.Request)
	AllowedMethods []string
}

func instrument(label string, fn func(http.ResponseWriter, *http.Request), methods []string) http.Handler {
	allowed := map[string]struct{}{}
	for _, method := range methods {
		allowed[method] = struct{}{}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[r.Method]; len(allowed) > 0 && !ok {
			msg := fmt.Sprintf("method %q not allowed", r.Method)
			http.Error(w, msg, http.StatusMethodNotAllowed)
			return
		}
		fn(w, r)
	}
	return weaver.InstrumentHandlerFunc(label, handler)
}

func (r Routes) Bind(mu *http.ServeMux) {
	for path, route := range r {
		mu.Handle(path, instrument(route.Titel, route.Handler, route.AllowedMethods))
	}
}
