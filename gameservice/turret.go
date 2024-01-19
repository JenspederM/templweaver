package gameservice

import "github.com/ServiceWeaver/weaver"

type Turret struct {
	weaver.AutoMarshal
	Position Point
	Symbol   string
	Range    int
	ammo     int
	Shots    int
}

func NewTurret(position Point, symbol string, range_, shots int) Turret {
	return Turret{Position: position, Symbol: symbol, Range: range_, Shots: shots, ammo: shots}
}

func (t *Turret) Shoot(m *Monster) {
	if t.Shots == 0 || !t.InRange(*m) || m.Health == 0 {
		return
	}
	for t.Shots > 0 && m.Health > 0 {
		t.Shots--
		m.Health--
	}
}

func (t *Turret) InRange(m Monster) bool {
	return t.Position.Distance(m.Position) <= t.Range
}

func (t *Turret) Reload() {
	t.Shots = t.ammo
}
