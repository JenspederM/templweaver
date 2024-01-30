package route

import "github.com/a-h/templ"

type ModelInterface interface {
	SetUser(string)
	SetRoutes(Routes)
}

type Route struct {
	Titel     string
	Model     ModelInterface
	Component func(model ModelInterface) templ.Component
}

type Routes map[string]Route
