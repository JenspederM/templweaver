package frontend

import (
	"github.com/a-h/templ"
	"github.com/jenspederm/templweaver/frontend/views"
	"github.com/jenspederm/templweaver/gameservice"
)

var Routes = map[string]views.Route{
	"/": {
		Title: "Home",
		ComponentProducerFunc: func() templ.Component {
			return views.Home("test")
		},
	},
	"/board": {
		Title: "Board",
		ComponentProducerFunc: func() templ.Component {
			return views.Board(&gameservice.GameState{}, [][]gameservice.Drawable{})
		},
	},
}
