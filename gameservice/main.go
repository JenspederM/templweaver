package gameservice

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/ServiceWeaver/weaver"
)

var ()

type GameService interface {
	GetBoard(ctx context.Context) ([]string, error)
	DrawBoard(ctx context.Context) error
	Move(ctx context.Context, reverse bool) error
}

type Point struct {
	X int
	Y int
}

type impl struct {
	weaver.Implements[GameService]
	Board       []string
	board       []string
	Turrets     map[string][]int
	Wave        []int
	Path        []Point
	current_pos int
}

func (s *impl) Init(ctx context.Context) error {
	s.Logger(ctx).Info("Game service started")
	s.Board = []string{
		"0111111",
		"  A  B1",
		" 111111",
		" 1     ",
		" 1C1111",
		" 111 D1",
		"      1",
	}
	s.board = []string{
		"0111111",
		"  A  B1",
		" 111111",
		" 1     ",
		" 1C1111",
		" 111 D1",
		"      1",
	}
	s.Turrets = map[string][]int{
		"A": {3, 2},
		"B": {1, 4},
		"C": {2, 2},
		"D": {1, 3},
	}
	s.Wave = []int{30, 14, 27, 21, 13, 0, 15, 17, 0, 18, 26}
	s.current_pos = 0
	s.initPath()

	return nil
}

func (s *impl) GetBoard(ctx context.Context) ([]string, error) {
	return s.Board, nil
}

type Monster struct {
	Position Point
	Health   int
}

type Monsters []Monster

func NewMonsters(position []Point, health []int) []Monster {
	if len(position) != len(health) {
		panic("position and health must be the same length")
	}

	slog.Info("creating new monster from", "position", position, "health", health)

	posReverse := make([]Point, len(position))
	copy(posReverse, position)

	slices.Reverse(posReverse)
	z := make(Monsters, len(position))
	for i := 0; i < len(position); i++ {
		z[i] = Monster{posReverse[i], health[i]}
	}

	return z
}

func (s *impl) Move(ctx context.Context, reverse bool) error {
	maxSteps := len(s.Path) + len(s.Wave) - 1
	if reverse {
		s.Logger(ctx).Info("reverse")
		s.current_pos--
		if s.current_pos < 0 {
			s.current_pos = 0
		}
	} else {
		s.current_pos++
		if s.current_pos > maxSteps {
			s.current_pos = maxSteps
		}
	}

	pathOffset := 0
	if s.current_pos-len(s.Path) > 0 {
		pathOffset = s.current_pos - len(s.Path)
	}

	waveMax := len(s.Wave) - 1
	pathMax := s.current_pos - pathOffset
	pathMin := Max(0, pathMax-waveMax) + pathOffset
	wave := s.Wave[pathOffset:Min(pathMax, waveMax)]
	monsters := NewMonsters(s.Path[pathMin:pathMax], wave)
	newBoard := make([]string, len(s.board))
	copy(newBoard, s.board)

	for _, monster := range monsters {
		fmt.Printf("monster: %v\n", monster)
		p := monster.Position
		newBoard[p.Y] = newBoard[p.Y][:p.X] + "M" + newBoard[p.Y][p.X+1:]
	}

	s.Board = newBoard
	return nil
}

func (s *impl) initPath() {
	start := Point{}
	unsorted_path := []Point{}
	for y := 0; y < len(s.Board); y++ {
		for x := 0; x < len(s.Board[0]); x++ {
			if string(s.Board[y][x]) == "1" {
				unsorted_path = append(unsorted_path, Point{x, y})
			} else if string(s.Board[y][x]) == "0" {
				start = Point{x, y}
			}
		}
	}

	pos := start
	dir := ""
	path := []Point{start}
	for len(path) <= len(unsorted_path) {
		left, err := s.peek(pos.X-1, pos.Y)
		if err != nil {
			fmt.Printf("there was a problem going %s", left)
		}
		right, err := s.peek(pos.X+1, pos.Y)
		if err != nil {
			fmt.Printf("there was a problem going %s", right)
		}
		up, err := s.peek(pos.X, pos.Y-1)
		if err != nil {
			fmt.Printf("there was a problem going %s", up)
		}
		down, err := s.peek(pos.X, pos.Y+1)
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
		path = append(path, Point{pos.X, pos.Y})
	}

	s.Path = path
}

func (s *impl) peek(x int, y int) (string, error) {
	if x < 0 || x >= len(s.Board[0]) || y < 0 || y >= len(s.Board) {
		return "", nil
	}
	return string(s.Board[y][x]), nil
}
func (s *impl) DrawBoard(ctx context.Context) error {
	if s.Board[0] == "X111111" {
		s.Board[0] = "0111111"
		return nil
	}

	return nil
}
