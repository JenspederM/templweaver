package server

import (
	"fmt"
	"net/http"

	"github.com/ServiceWeaver/weaver"
	"github.com/jenspederm/templweaver/layout"
)

func instrument(label string, fn func(http.ResponseWriter, *http.Request), methods []string) http.Handler {
	allowed := map[string]struct{}{}
	for _, method := range methods {
		allowed[method] = struct{}{}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[r.Method]; len(allowed) > 0 && !ok {
			layout.Error(r.Response.StatusCode, fmt.Sprintf("method %q not allowed", r.Method)).Render(r.Context(), w)
			msg := fmt.Sprintf("method %q not allowed", r.Method)
			http.Error(w, msg, http.StatusMethodNotAllowed)
			return
		}
		fn(w, r)
	}
	return weaver.InstrumentHandlerFunc(label, handler)
}
