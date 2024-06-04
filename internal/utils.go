package internal

import (
	"math/rand"
	"os"
)

func Choose[T any](selection ...T) T {
	return selection[rand.Intn(len(selection))]
}

// TODO: convert to use slog
func Log(log string) {
	f, err := os.OpenFile("nemo.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(log); err != nil {
		panic(err)
	}
}
