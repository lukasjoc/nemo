package internal

import "math/rand"

func Choose[T any](selection ...T) T {
	return selection[rand.Intn(len(selection))]
}

func IntRand(n int) int { return rand.Intn(intMax(n, 1)) }
