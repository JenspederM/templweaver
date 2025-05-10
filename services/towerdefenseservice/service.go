package towerdefenseservice

import (
	"context"
	"fmt"

	"github.com/ServiceWeaver/weaver"
)

type DrawableType int

const (
	DrawableMonster DrawableType = iota
	DrawableTurret
	DrawablePath
	DrawableStart
	DrawableTree
)

type Drawable struct {
	weaver.AutoMarshal
	Position Point
	Type     DrawableType
	Symbol   string
	Tooltip  string
	Class    string
}

type Game struct {
	weaver.AutoMarshal
	board   []string
	turrets map[string][2]int
	wave    []int
}

type GameService interface {
	Draw(ctx context.Context, move bool, reverse bool) (*GameState, [][]Drawable, error)
}

type gameServiceImpl struct {
	weaver.Implements[GameService]
	state      *GameState
	states     []GameState
	original   []string
	maxRounds  int
	Path       []Point
	Directions []Direction
	Start      Point
	Final      Point
}

func (b *gameServiceImpl) Init(ctx context.Context) error {
	b.Logger(ctx).Info("Game board started")
	g1 := Game{
		board: []string{
			"0111111",
			"  A  B1",
			" 111111",
			" 1     ",
			" 1C1111",
			" 111 D1",
			"      1",
		},
		turrets: map[string][2]int{
			"A": {3, 2},
			"B": {1, 4},
			"C": {2, 2},
			"D": {1, 3},
		},
		wave: []int{30, 14, 27, 21, 13, 0, 15, 17, 0, 18, 26},
	}
	g2 := Game{
		board: []string{
			"011  1111",
			" A1  1BC1",
			" 11  1 11",
			" 1D  1 1E",
			" 111 1F11",
			"  G1 1  1",
			" 111 1 11",
			" 1H  1 1I",
			" 11111 11",
		},
		turrets: map[string][2]int{
			"A": {1, 4},
			"B": {2, 2},
			"C": {1, 3},
			"D": {1, 3},
			"E": {1, 2},
			"F": {3, 3},
			"G": {1, 2},
			"H": {2, 3},
			"I": {2, 3},
		},
		wave: []int{36, 33, 46, 35, 44, 27, 25, 48, 39, 0, 39, 36, 55, 22, 26},
	}

	games := []Game{g1, g2}
	game := games[0]
	monsters := []*Monster{}
	for _, health := range game.wave {
		monsters = append(monsters, NewMonster(health))
	}

	// original is used to draw the board
	b.original = game.board
	start, path, directions, turrets, err := b.findElements(ctx, game.turrets)
	if err != nil {
		return err
	}

	s := NewGameState(game.board, turrets, monsters)
	b.maxRounds = len(path) + len(monsters) - 1
	b.Start = start
	b.Final = path[len(path)-1]
	b.Path = copyAndReverse(path) // We want to start from the end
	b.Directions = copyAndReverse(directions)
	b.state = s
	b.states = []GameState{*s}

	return nil
}

func (b *gameServiceImpl) Draw(ctx context.Context, move bool, reverse bool) (*GameState, [][]Drawable, error) {
	if move {
		err := b.updateState(ctx, reverse)
		if err != nil {
			return nil, nil, err
		}
	}

	monsterData, turretData, err := b.getData(ctx)
	if err != nil {
		return nil, nil, err
	}

	b.Logger(ctx).Info("Drawing state", "cursor", b.state.Round)
	sizeClass := "w-8 h-8"
	drawables := make([][]Drawable, len(b.state.Current))
	for i, row := range b.state.Current {
		drawables[i] = make([]Drawable, len(row))
		for j := range row {
			d := Drawable{Position: NewPoint(j, i), Symbol: string(row[j]), Class: sizeClass}
			switch string(row[j]) {
			case " ":
				d.Type = DrawableTree
			case "0":
				d.Type = DrawableStart
			case "1":
				d.Type = DrawablePath
			case "@":
				if md, ok := monsterData[NewPoint(j, i)]; ok {
					if md[1] == "fill-error" {
						d.Type = DrawablePath
					} else {
						d.Type = DrawableMonster
						d.Tooltip = md[0]
						d.Class = fmt.Sprintf("%s %s", sizeClass, md[1])
					}
				}
			default:
				if td, ok := turretData[NewPoint(j, i)]; ok {
					d.Type = DrawableTurret
					d.Tooltip = td[0]
					d.Class = fmt.Sprintf("%s %s", sizeClass, td[1])
				}
			}

			drawables[i][j] = d
		}
	}

	return b.state, drawables, nil
}

func (b *gameServiceImpl) updateState(ctx context.Context, reverse bool) error {
	currentRound := b.state.Round

	b.Logger(ctx).Debug("Updating state", "cursor", currentRound, "monsters", b.state.Monsters, "reverse", reverse)

	if reverse {
		if b.state.Round == 0 {
			return nil
		}

		for _, state := range b.states {
			if state.Round == b.state.Round-1 {
				b.state = &state
				b.Logger(ctx).Debug("Updated state", "cursor", b.state.Round, "monsters", b.state.Monsters)
				return nil
			}
		}

		if b.state.Round == currentRound {
			return fmt.Errorf("cannot find state for round %d", b.state.Round-1)
		}
	}

	newState, err := b.state.Update(ctx, b.original, b.Path, b.Directions)
	if err != nil {
		return err
	}
	b.states = append(b.states, *newState)
	b.state = newState
	b.Logger(ctx).Debug("Updated state", "cursor", b.state.Round, "monsters", b.state.Monsters)
	return nil
}

func (b *gameServiceImpl) getData(ctx context.Context) (map[Point][2]string, map[Point][2]string, error) {
	// list of tooltip, class
	monsterData := map[Point][2]string{}
	for i, monster := range b.state.Monsters {
		p := monster.Position
		_class := "fill-success"
		if p.X != -1 && p.Y != -1 && monster.IsDead() {

			_class = "fill-error"
			fmt.Printf("monster at %v is dead %d %s\n", p, monster.Health, _class)
		}
		monsterData[p] = [2]string{
			fmt.Sprintf("%d, [%d, %d]\nHealth: %d", i, monster.Position.Y, monster.Position.X, monster.Health),
			_class,
		}
	}

	turretData := map[Point][2]string{}
	for _, turret := range b.state.Turrets {
		p := turret.Position
		_class := "fill-success"
		if turret.Shots == 0 {
			_class = "fill-error"
			fmt.Printf("turret at %v is empty %s\n", p, _class)
		}
		turretData[p] = [2]string{
			fmt.Sprintf("%s\n[%d, %d]\nRange: %d\nShots: %d", turret.Symbol, turret.Position.Y, turret.Position.X, turret.Range, turret.Shots),
			_class,
		}
	}

	return monsterData, turretData, nil
}

func (b *gameServiceImpl) getFromOriginal(_ context.Context, x int, y int) (string, error) {
	if x < 0 || x >= len(b.original[0]) || y < 0 || y >= len(b.original) {
		return "", nil
	}
	return string(b.original[y][x]), nil
}

func (b *gameServiceImpl) findElements(ctx context.Context, turretMap map[string][2]int) (Point, []Point, []Direction, []*Turret, error) {
	start := Point{}
	turrets := []*Turret{}
	unsorted_path := []Point{}
	rows := len(b.original)
	cols := len(b.original[0])
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			val, err := b.getFromOriginal(ctx, x, y)
			if err != nil {
				return Point{}, nil, nil, nil, err
			}
			if val == " " {
				continue
			}

			switch val {
			case "1":
				unsorted_path = append(unsorted_path, NewPoint(x, y))
			case "0":
				start = NewPoint(x, y)
			default: // turret
				stats, ok := turretMap[val]
				if !ok {
					return Point{}, nil, nil, nil, fmt.Errorf("turret %s not found", val)
				}
				if len(stats) != 2 {
					return Point{}, nil, nil, nil, fmt.Errorf("turret %s has invalid stats", val)
				}
				turrets = append(turrets, NewTurret(NewPoint(x, y), val, stats[0], stats[1]))
			}

		}
	}

	path, directions, err := b.sortPath(ctx, unsorted_path, start)
	if err != nil {
		return Point{}, nil, nil, nil, err
	}

	return start, path, directions, turrets, nil
}

func (b *gameServiceImpl) sortPath(ctx context.Context, unsorted_path []Point, start Point) ([]Point, []Direction, error) {
	pos := start
	dir := ""
	path := []Point{start}
	directions := []Direction{}
	for len(path) <= len(unsorted_path) {
		left, err := b.getFromOriginal(ctx, pos.X-1, pos.Y)
		if err != nil {
			return nil, nil, err
		}
		right, err := b.getFromOriginal(ctx, pos.X+1, pos.Y)
		if err != nil {
			return nil, nil, err
		}
		up, err := b.getFromOriginal(ctx, pos.X, pos.Y-1)
		if err != nil {
			return nil, nil, err
		}
		down, err := b.getFromOriginal(ctx, pos.X, pos.Y+1)
		if err != nil {
			return nil, nil, err
		}
		if left == "1" && dir != "right" {
			dir = "left"
			directions = append(directions, Left)
			pos.X--
		} else if right == "1" && dir != "left" {
			dir = "right"
			directions = append(directions, Right)
			pos.X++
		} else if up == "1" && dir != "down" {
			dir = "up"
			directions = append(directions, Up)
			pos.Y--
		} else if down == "1" && dir != "up" {
			dir = "down"
			directions = append(directions, Down)
			pos.Y++
		}
		path = append(path, NewPoint(pos.X, pos.Y))
	}
	return path, directions, nil
}
