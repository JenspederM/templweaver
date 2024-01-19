package gameservice

import (
	"context"
	"slices"

	"github.com/ServiceWeaver/weaver"
)

type GameService interface {
	GetBoard(ctx context.Context) ([]string, error)
	Draw(ctx context.Context, reverse bool) error
}

type impl struct {
	weaver.Implements[GameService]
	cursor   int
	Board    Board
	Start    Point
	Final    Point
	Path     []Point
	Turrets  []Turret
	Monsters []Monster
	Score    int
}

func (s *impl) Init(ctx context.Context) error {
	s.Logger(ctx).Info("Game service started")
	board := []string{
		"0111111",
		"  A  B1",
		" 111111",
		" 1     ",
		" 1C1111",
		" 111 D1",
		"      1",
	}
	turrentMap := map[string][]int{
		"A": {3, 2},
		"B": {1, 4},
		"C": {2, 2},
		"D": {1, 3},
	}
	wave := []int{30, 14, 27, 21, 13, 0, 15, 17, 0, 18, 26}

	monsters := []Monster{}
	for _, health := range wave {
		monsters = append(monsters, NewMonster(health))
	}

	s.cursor = 0
	s.Board = *NewBoard(board)
	s.Monsters = monsters

	start, final, path, turrets := s.Board.Parse(turrentMap)
	s.Start = start
	s.Final = final
	s.Path = path
	s.Turrets = turrets
	return nil
}

func (s *impl) GetBoard(ctx context.Context) ([]string, error) {
	return s.Board.Current, nil
}

func (s *impl) moveCursor(ctx context.Context, maxSteps int, reverse bool) error {
	if reverse {
		s.Logger(ctx).Info("reverse")
		s.cursor--
		if s.cursor < 0 {
			s.cursor = 0
		}
	} else {
		s.cursor++
		if s.cursor > maxSteps {
			s.cursor = maxSteps
		}
	}
	return nil
}

func (s *impl) Draw(ctx context.Context, reverse bool) error {
	maxSteps := len(s.Path) + len(s.Monsters) - 1
	s.moveCursor(ctx, maxSteps, reverse)
	index := Max(0, len(s.Path)-s.cursor)
	cursorOffset := 0
	if s.cursor-len(s.Path) > 0 {
		cursorOffset = s.cursor - len(s.Path)
	}
	reversedPath := make([]Point, len(s.Path))
	copy(reversedPath, s.Path)
	slices.Reverse(reversedPath)

	moved := []Monster{}
	for i, p := range reversedPath[index:] {
		if i >= len(s.Monsters)-1-cursorOffset {
			break
		}
		s.Monsters[i+cursorOffset].Position = p
		mpos := s.Monsters[i+cursorOffset].Position
		spos := s.Final

		if mpos.X == spos.X && mpos.Y == spos.Y {
			s.Logger(ctx).Info("final", "monster", s.Monsters[i+cursorOffset], "final", s.Final)
			s.Score += s.Monsters[i+cursorOffset].Health
		}
		moved = append(moved, s.Monsters[i+cursorOffset])
	}

	for _, turret := range s.Turrets {
		turret.Reload()
		for i, monster := range s.Monsters {
			if turret.InRange(monster) {
				turret.Shoot(&monster)
				s.Monsters[i] = monster

			}

		}
	}
	s.Logger(ctx).Info("Score", "Score", s.Score)
	s.Board.Draw(moved)
	return nil
}
