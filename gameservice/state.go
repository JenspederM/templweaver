package gameservice

import (
	"context"
	"log/slog"
	"slices"

	"github.com/ServiceWeaver/weaver"
)

type GameState struct {
	weaver.AutoMarshal
	Round                int
	Current              []string
	Turrets              []*Turret
	Monsters             []*Monster
	Survivors            [][2]int
	Score                int
	original_blank_count int
}

func NewGameState(board []string, turrets []*Turret, monsters []*Monster) *GameState {
	blank_count := 0
	for _, m := range monsters {
		if m.Health == 0 {
			blank_count++
		}
	}
	return &GameState{
		Round:                0,
		Current:              board,
		Turrets:              turrets,
		Monsters:             monsters,
		Survivors:            [][2]int{},
		Score:                0,
		original_blank_count: blank_count,
	}
}

func (b *GameState) Update(_ context.Context, original []string, path []Point, directions []Direction) (*GameState, error) {
	slog.Debug("Updating state", "cursor", b.Round)
	newMonsters := []*Monster{}
	for _, monster := range b.Monsters {
		m := NewMonster(monster.Health)
		m.Position = monster.Position
		newMonsters = append(newMonsters, m)
	}

	newTurrets := []*Turret{}
	for _, turret := range b.Turrets {
		newTurrets = append(newTurrets, NewTurret(turret.Position, turret.Symbol, turret.Range, turret.Shots))
	}

	newBoard := make([]string, len(b.Current))
	copy(newBoard, original)

	// Should not use NewGameState here, as it will reset the blank count
	s := GameState{
		Round:                b.Round,
		Current:              newBoard,
		Turrets:              newTurrets,
		Monsters:             newMonsters,
		Survivors:            b.Survivors,
		Score:                b.Score,
		original_blank_count: b.original_blank_count,
	}

	for i, m := range s.Monsters {
		if m.Position == path[0] {
			slog.Debug("Monster reached end", "monster", m)
			m.Position = NewPoint(-1, -1)
			if m.Health > 0 {
				s.Survivors = append(s.Survivors, [2]int{Max(0, i-b.original_blank_count), m.Health})
				s.Score += m.Health
			}
		}
	}

	s.moveCursor(len(path) + len(b.Monsters))
	n := len(b.Monsters)
	quo, rem := s.Round/n, s.Round%n
	waveTail := Max(0, (quo-1))*n + rem
	if quo == 0 {
		waveTail = -1
	}

	slog.Info("Calculating offsets", "quotient", quo, "remainder", rem, "waveTail", waveTail, "len(path)", len(path), "len(b.Monsters)", len(b.Monsters))

	moved, err := s.moveMonster(context.Background(), s.Round, path)
	if err != nil {
		return nil, err
	}

	for _, turret := range b.Turrets {
		turret.Reload()
	}

	err = s.shootTurrets(context.Background())
	if err != nil {
		return nil, err
	}

	for _, monster := range moved {
		p := monster.Position
		s.Current[p.Y] = s.Current[p.Y][:p.X] + "@" + s.Current[p.Y][p.X+1:]
	}

	return &s, nil
}

func (b *GameState) moveMonster(_ context.Context, cursor int, path []Point) ([]Monster, error) {
	waveHead := Max(0, len(path)-cursor)
	monsterOffset := 0
	if len(path)-cursor < 0 {
		monsterOffset = Max(0, cursor-len(path))
	}
	slog.Debug("Moving monsters", "waveHead", waveHead, "monsterOffset", monsterOffset, "cursor", cursor)
	moved := []Monster{}
	for i, p := range path[waveHead:] {
		if i >= len(b.Monsters)-monsterOffset {
			break
		}
		m := b.Monsters[i+monsterOffset]
		m.Position = p
		moved = append(moved, *m)
	}

	return moved, nil
}

func (b *GameState) shootTurrets(_ context.Context) error {
	turretsOutOfAmmo := []string{}
	for len(turretsOutOfAmmo) != len(b.Turrets) {
		for _, turret := range b.Turrets {
			if slices.Contains(turretsOutOfAmmo, turret.Symbol) {
				continue
			}
			if turret.Ammo == 0 {
				slog.Debug("Turret out of ammo", "turret", turret.Symbol, "ammo", turret.Ammo, "shots", turret.Shots)
				turretsOutOfAmmo = append(turretsOutOfAmmo, turret.Symbol)
				continue
			}

			monstersInRange := 0
			for i, monster := range b.Monsters {
				if monster.Health == 0 {
					continue
				}
				if turret.InRange(monster) {
					monstersInRange++
					turret.Shoot(monster)
					slog.Debug("Shooting monster", "turret", turret.Symbol, "monster", monster, "range", turret.Range, "distance", turret.Position.Distance(monster.Position), "ammo", turret.Ammo, "shots", turret.Shots)
					b.Monsters[i] = monster
					break
				}
			}

			// If we didn't shoot anything, we're either out of ammo, there are no monsters left, or we're out of range
			if monstersInRange == 0 || turret.Ammo == turret.Shots {
				slog.Debug("Turret did nothing", "turret", turret.Symbol, "ammo", turret.Ammo, "shots", turret.Shots)
				turretsOutOfAmmo = append(turretsOutOfAmmo, turret.Symbol)
			}
		}
		slog.Debug("Fired turrets", "turretOutOfAmmo", turretsOutOfAmmo, "len(b.Turrets)", len(b.Turrets))
	}
	return nil
}

func (b *GameState) moveCursor(maxRounds int) {
	before := b.Round
	b.Round++
	if b.Round >= maxRounds {
		b.Round = maxRounds
	}
	slog.Debug("Moving cursor", "before", before, "after", b.Round, "maxRounds", maxRounds)
}
