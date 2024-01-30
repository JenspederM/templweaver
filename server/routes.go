package server

import (
	"github.com/jenspederm/templweaver/server/internal/route"
)

var Routes = route.Routes{
	"/": {
		Titel: "Home",
	},
	"/board": {
		Titel: "Board",
	},
}
