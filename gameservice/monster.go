package gameservice

import "github.com/ServiceWeaver/weaver"

type Monster struct {
	weaver.AutoMarshal
	Position Point
	Health   int
}

func NewMonster(health int) Monster {
	return Monster{Position: NewPoint(-1, -1), Health: health}
}
