package frontend

import (
	"fmt"
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
		"/":      {Titel: "Home"},
		"/page1": {Titel: "Page 1"},
		"/board": {Titel: "Board"},
	}
	if err := s.authservice.Get().CheckIsLoggedIn(r.Context()); err == nil {
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
func (s *Server) boardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		method := r.FormValue("method")
		fmt.Printf("method: %s\n", method)
		s.gameservice.Get().Move(r.Context(), method == "previous")
		err := s.gameservice.Get().DrawBoard(r.Context())
		if err != nil {
			s.renderView("Board", views.Error(500, err.Error()), w, r)
			return
		}
		board, err := s.gameservice.Get().GetBoard(r.Context())
		if err != nil {
			s.renderView("Board", views.Error(500, err.Error()), w, r)
			return
		}
		views.InnerBoard(board).Render(r.Context(), w)
		return
	}
	board, err := s.gameservice.Get().GetBoard(r.Context())
	if err != nil {
		s.renderView("Board", views.Error(500, err.Error()), w, r)
		return
	}
	s.renderView("Board", views.Board(board), w, r)
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
