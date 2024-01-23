package frontend

import (
	"net/http"

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

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		views.Error(404, "Not found").Render(r.Context(), w)
		return
	}
	views.Home(sessionID(r)).Render(r.Context(), w)
}
func (s *Server) page1Handler(w http.ResponseWriter, r *http.Request) {
	views.Page1().Render(r.Context(), w)
}
func (s *Server) boardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		reverse := r.FormValue("method") == "previous"
		state, drawables, err := s.gameservice.Get().Draw(r.Context(), true, reverse)
		s.Logger(r.Context()).Info("boardHandler", "survivors", state.Survivors, "round", state.Round, "reverse", reverse)
		if err != nil {
			views.Error(500, err.Error()).Render(r.Context(), w)
			return
		}
		views.HtmxBoard("board", state, drawables).Render(r.Context(), w)
		return
	}

	state, drawables, err := s.gameservice.Get().Draw(r.Context(), false, false)
	if err != nil {
		views.Error(500, err.Error()).Render(r.Context(), w)
		return
	}

	views.Board(state, drawables).Render(r.Context(), w)
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
