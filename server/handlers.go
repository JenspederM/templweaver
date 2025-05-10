package server

import (
	"net/http"

	"github.com/jenspederm/templweaver/layout"
	"github.com/jenspederm/templweaver/views"
)

func sessionID(r *http.Request) string {
	v := r.Context().Value(ctxKeySessionID{})
	if v != nil {
		return v.(string)
	}
	return ""
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		layout.Error(404, "Not found").Render(r.Context(), w)
		return
	}
	views.Home(sessionID(r), s.routes).Render(r.Context(), w)
}

func (s *Server) boardHandler(w http.ResponseWriter, r *http.Request) {
	service := s.towerdefenseservice.Get()

	if r.Method == http.MethodPost {
		reverse := r.FormValue("method") == "previous"
		state, drawables, err := service.Draw(r.Context(), true, reverse)
		s.Logger(r.Context()).Info("boardHandler", "survivors", state.Survivors, "round", state.Round, "reverse", reverse)
		if err != nil {
			layout.Error(500, err.Error()).Render(r.Context(), w)
			return
		}
		views.HtmxBoard("board", state, drawables).Render(r.Context(), w)
		return
	}

	state, drawables, err := service.Draw(r.Context(), false, false)
	if err != nil {
		layout.Error(500, err.Error()).Render(r.Context(), w)
		return
	}

	views.Board(state, drawables, s.routes).Render(r.Context(), w)
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
