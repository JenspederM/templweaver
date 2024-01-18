package frontend

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/jenspederm/templweaver/frontend/views"
)

type GlobalState struct {
}

func sessionID(r *http.Request) string {
	v := r.Context().Value(ctxKeySessionID{})
	if v != nil {
		return v.(string)
	}
	return ""
}

func (s *Server) renderView(title string, view templ.Component, w http.ResponseWriter, r *http.Request) {
	var routes = map[string]views.Route{
		"/":      {Titel: "Home", Component: views.Home(sessionID(r))},
		"/page1": {Titel: "Page 1", Component: views.Page1()},
		"/page2": {Titel: "Page 2", Component: views.Home(sessionID(r))},
	}
	if err := s.authservice.Get().CheckIsLoggedIn(r.Context()); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		views.Index(routes, title, view, false).Render(r.Context(), w)
		return
	}
	if r.Header.Get("Hx-Current-Url") == "" || r.Header.Get("Hx-Target") == "root" {
		views.Index(routes, title, view, true).Render(r.Context(), w)
	}
	view.Render(r.Context(), w)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.renderView("Not Found", views.Error(404, "Not found"), w, r)
		return
	}
	s.renderView("Home", views.Home(sessionID(r)), w, r)
}
func (s *Server) page1Handler(w http.ResponseWriter, r *http.Request) {
	s.renderView("Page 1", views.Page1(), w, r)
}
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		err := s.authservice.Get().Login(r.Context(), email, password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
