package gameservice

import (
	"fmt"

	"github.com/ServiceWeaver/weaver"
)

type Monster struct {
	weaver.AutoMarshal
	Position Point
	Health   int
}

func NewMonster(health int) *Monster {
	return &Monster{Position: NewPoint(-1, -1), Health: health}
}

func (m *Monster) String() string {
	return fmt.Sprintf("Monster{Position: %v, Health: %v}", m.Position, m.Health)
}

func (m *Monster) IsDead() bool {
	if m.Health <= 0 {
		return true
	}
	return m.Health <= 0 || m.Position.X == -1 || m.Position.Y == -1
}
