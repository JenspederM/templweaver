package frontend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jenspederm/templweaver/frontend/views"
)

var state = views.State{Count: 0}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	views.Index(state).Render(context.Background(), w)
}
func (s *Server) incrementHandler(w http.ResponseWriter, r *http.Request) {
	state.Count++
	count := strconv.Itoa(state.Count)
	w.Write([]byte(count))
}
func (s *Server) resetHandler(w http.ResponseWriter, r *http.Request) {
	state.Count = 0
	count := strconv.Itoa(state.Count)
	w.Write([]byte(count))
}
