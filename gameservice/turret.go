package gameservice

import (
	"fmt"

	"github.com/ServiceWeaver/weaver"
)

type Turret struct {
	weaver.AutoMarshal
	Position Point
	Symbol   string
	Range    int
	Shots    int
	Ammo     int
}

func NewTurret(position Point, symbol string, range_, shots int) *Turret {
	return &Turret{Position: position, Symbol: symbol, Range: range_, Shots: shots, Ammo: shots}
}

func (t *Turret) String() string {
	return fmt.Sprintf("Turret{Symbol: %v, Range: %v, Shots: %v}", t.Symbol, t.Range, t.Shots)
}

func (t *Turret) Shoot(m *Monster) {
	if t.Ammo == 0 || !t.InRange(m) || m.Health == 0 {
		return
	}
	t.Ammo--
	m.Health--
}

func (t *Turret) InRange(m *Monster) bool {
	return t.Position.Distance(m.Position) <= t.Range
}

func (t *Turret) Reload() {
	t.Ammo = t.Shots
}
