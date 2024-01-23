package gameservice

import "slices"

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Min(x, y int) int {
	if y > x {
		return x
	}
	return y
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func copyAndReverse[V any](input []V) []V {
	reversed := make([]V, len(input))
	copy(reversed, input)
	slices.Reverse(reversed)
	return reversed
}
