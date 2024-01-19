package gameservice

import "github.com/ServiceWeaver/weaver"

type Point struct {
	weaver.AutoMarshal
	X int
	Y int
}

func NewPoint(x, y int) Point {
	return Point{X: x, Y: y}
}

func (p *Point) Distance(p2 Point) int {
	return Abs(p.X-p2.X) + Abs(p.Y-p2.Y)
}
