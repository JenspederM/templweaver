package gameservice

import (
	"fmt"
)

type Board struct {
	original []string
	Current  []string
	Path     []Point
}

func NewBoard(board []string) *Board {
	return &Board{
		original: board,
		Current:  board,
	}
}

func (b *Board) Rows() int {
	return len(b.Current)
}

func (b *Board) Cols() int {
	return len(b.Current[0])
}

func (b *Board) Reset() {
	b.Current = b.original
}

func (b *Board) Set(x, y int, symbol string) {
	b.Current[y] = b.Current[y][x:] + symbol + b.Current[y][:x+1]
}

func (b *Board) Get(x int, y int) (string, error) {
	if x < 0 || x >= len(b.Current[0]) || y < 0 || y >= len(b.Current) {
		return "", nil
	}
	return string(b.Current[y][x]), nil
}

func (b *Board) Draw(monsters []Monster) error {
	newBoard := make([]string, len(b.Current))
	copy(newBoard, b.original)

	for _, monster := range monsters {
		// slog.Info("Board.draw", "monster", monster)
		p := monster.Position
		newBoard[p.Y] = newBoard[p.Y][:p.X] + "M" + newBoard[p.Y][p.X+1:]
	}

	b.Current = newBoard
	return nil
}

func (b *Board) Parse(turretMap map[string][]int) (Point, Point, []Point, []Turret) {
	start := Point{}
	final := Point{}
	path := []Point{start}
	unsorted_path := []Point{}
	turrets := []Turret{}
	for y := 0; y < b.Rows(); y++ {
		for x := 0; x < b.Cols(); x++ {
			val, err := b.Get(x, y)
			if err != nil {
				fmt.Printf("there was a problem getting %d, %d", x, y)
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
					fmt.Printf("turret %s not found\n", val)
				}
				if len(stats) != 2 {
					fmt.Printf("turret %s has wrong stats\n", val)
				}
				turrets = append(turrets, NewTurret(NewPoint(x, y), val, stats[0], stats[1]))
			}

		}
	}

	pos := start
	dir := ""
	for len(path) <= len(unsorted_path) {
		left, err := b.Get(pos.X-1, pos.Y)
		if err != nil {
			fmt.Printf("there was a problem going %s", left)
		}
		right, err := b.Get(pos.X+1, pos.Y)
		if err != nil {
			fmt.Printf("there was a problem going %s", right)
		}
		up, err := b.Get(pos.X, pos.Y-1)
		if err != nil {
			fmt.Printf("there was a problem going %s", up)
		}
		down, err := b.Get(pos.X, pos.Y+1)
		if err != nil {
			fmt.Printf("there was a problem going %s", down)
		}
		if left == "1" && dir != "right" {
			dir = "left"
			pos.X--
		} else if right == "1" && dir != "left" {
			dir = "right"
			pos.X++
		} else if up == "1" && dir != "down" {
			dir = "up"
			pos.Y--
		} else if down == "1" && dir != "up" {
			dir = "down"
			pos.Y++
		}
		path = append(path, NewPoint(pos.X, pos.Y))
	}

	final = path[len(path)-1]

	return start, final, path, turrets
}
