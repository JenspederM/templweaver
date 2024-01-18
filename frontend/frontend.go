package frontend

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/ServiceWeaver/weaver"
	"github.com/jenspederm/templweaver/authservice"
	"github.com/jenspederm/templweaver/frontend/views"
	"github.com/joho/godotenv"
)

const (
	cookieMaxAge = 60 * 60 * 48

	cookiePrefix    = "templweaver_"
	cookieSessionID = cookiePrefix + "session-id"
	cookieCurrency  = cookiePrefix + "currency"
)

var (
	//go:embed static/*
	staticFS embed.FS
)

// Server is the application frontend.
type Server struct {
	weaver.Implements[weaver.Main]

	handler  http.Handler
	hostname string

	// Setup the services we need.
	authservice weaver.Ref[authservice.AuthService]

	// Setup the listeners we need.
	frontend weaver.Listener
}

func Serve(ctx context.Context, s *Server) error {
	godotenv.Load()

	env := os.Getenv("ENV_PLATFORM")
	// Find out where we're running.
	// Set ENV_PLATFORM (default to local if not set; use env var if set;
	// otherwise detect GCP, which overrides env).
	s.Logger(ctx).Debug("ENV_PLATFORM", "platform", env)
	hn, err := os.Hostname()
	if err != nil {
		s.Logger(ctx).Debug(`cannot get hostname for frontend: using "unknown"`)
		hn = "unknown"
	}
	s.hostname = hn
	// Setup the handler.
	staticHTML, err := fs.Sub(fs.FS(staticFS), "static")
	if err != nil {
		return err
	}
	r := http.NewServeMux()

	// Helper that adds a handler with HTTP metric instrumentation.
	instrument := func(label string, fn func(http.ResponseWriter, *http.Request), methods []string) http.Handler {
		allowed := map[string]struct{}{}
		for _, method := range methods {
			allowed[method] = struct{}{}
		}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowed[r.Method]; len(allowed) > 0 && !ok {
				views.Error(r.Response.StatusCode, fmt.Sprintf("method %q not allowed", r.Method)).Render(r.Context(), w)
				msg := fmt.Sprintf("method %q not allowed", r.Method)
				http.Error(w, msg, http.StatusMethodNotAllowed)
				return
			}
			fn(w, r)
		}
		return weaver.InstrumentHandlerFunc(label, handler)
	}

	const get = http.MethodGet
	const post = http.MethodPost
	const head = http.MethodHead
	r.Handle("/", instrument("home", s.indexHandler, []string{get, head}))
	r.Handle("/login", instrument("login", s.loginHandler, []string{post}))
	r.Handle("/page1", instrument("page1", s.page1Handler, []string{get, head}))
	r.Handle("/ping", instrument("ping", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "pong") }, []string{post}))
	r.Handle("/static/", weaver.InstrumentHandler("static", http.StripPrefix("/static/", http.FileServer(http.FS(staticHTML)))))
	r.Handle("/robots.txt", instrument("robots", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "User-agent: *\nDisallow: /") }, nil))
	r.HandleFunc(weaver.HealthzURL, weaver.HealthzHandler)

	var handler http.Handler = r
	handler = ensureSessionID(handler)              // add session ID
	handler = newLogHandler(s.Logger(ctx), handler) // add logging
	s.handler = handler

	s.Logger(ctx).Debug("Frontend available", "addr", s.frontend)
	return http.Serve(s.frontend, s.handler)
}
