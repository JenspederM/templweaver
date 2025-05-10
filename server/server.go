package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/ServiceWeaver/weaver"
	"github.com/jenspederm/templweaver/models"
	"github.com/jenspederm/templweaver/services/authservice"
	"github.com/jenspederm/templweaver/services/towerdefenseservice"
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
	routes   models.Routes

	// Setup the services we need.
	authservice         weaver.Ref[authservice.AuthService]
	towerdefenseservice weaver.Ref[towerdefenseservice.GameService]

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
	const get = http.MethodGet
	const post = http.MethodPost
	const head = http.MethodHead

	s.routes = models.Routes{
		"/": {
			ApiOnly:        false,
			Titel:          "Home",
			Handler:        s.homeHandler,
			AllowedMethods: []string{get, head},
		},
		"/towerdefense": {
			ApiOnly:        false,
			Titel:          "Tower Defense",
			Handler:        s.boardHandler,
			AllowedMethods: []string{get, head, post},
		},
		"/ping": {
			ApiOnly:        true,
			Handler:        func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "pong") },
			AllowedMethods: []string{post},
		},
		"/login": {
			ApiOnly:        true,
			Handler:        s.loginHandler,
			AllowedMethods: []string{post},
		},
		"/robots.txt": {
			ApiOnly:        true,
			Handler:        func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "User-agent: *\nDisallow: /") },
			AllowedMethods: []string{get, head},
		},
		"/static/": {
			ApiOnly:        true,
			Handler:        http.StripPrefix("/static/", http.FileServer(http.FS(staticHTML))).ServeHTTP,
			AllowedMethods: []string{get, head},
		},
		weaver.HealthzURL: {
			ApiOnly:        true,
			Handler:        weaver.HealthzHandler,
			AllowedMethods: []string{get, head, post},
		},
	}

	s.routes.Bind(r)
	var handler http.Handler = r
	handler = ensureSessionID(handler)              // add session ID
	handler = newLogHandler(s.Logger(ctx), handler) // add logging
	s.handler = handler

	s.Logger(ctx).Debug("Frontend available", "addr", s.frontend)
	return http.Serve(s.frontend, s.handler)
}
