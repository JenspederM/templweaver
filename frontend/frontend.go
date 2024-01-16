package frontend

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/ServiceWeaver/weaver"
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

type platformDetails struct {
	css      string
	provider string
}

func (plat *platformDetails) setPlatformDetails(env string) {
	if env == "gcp" {
		plat.provider = "Google Cloud"
		plat.css = "gcp-platform"
	} else {
		plat.provider = "local"
		plat.css = "local"
	}
}

// Server is the application frontend.
type Server struct {
	weaver.Implements[weaver.Main]

	handler  http.Handler
	platform platformDetails
	hostname string

	frontend weaver.Listener
}

func Serve(ctx context.Context, s *Server) error {
	godotenv.Load()

	env := os.Getenv("ENV_PLATFORM")
	// Find out where we're running.
	// Set ENV_PLATFORM (default to local if not set; use env var if set;
	// otherwise detect GCP, which overrides env).
	s.Logger(ctx).Debug("ENV_PLATFORM", "platform", env)
	s.platform = platformDetails{}
	s.platform.setPlatformDetails(strings.ToLower(env))
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
			fmt.Printf("allowed: %v\n", allowed)
			if _, ok := allowed[r.Method]; len(allowed) > 0 && !ok {
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
	r.Handle("/increment", instrument("increment", s.incrementHandler, []string{post}))
	r.Handle("/reset", instrument("reset", s.resetHandler, []string{post}))
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
