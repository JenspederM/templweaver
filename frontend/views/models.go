package views

import "github.com/a-h/templ"

type ComponentProducerFunc func() templ.Component

type Route struct {
	Title                 string
	ComponentProducerFunc ComponentProducerFunc
}
