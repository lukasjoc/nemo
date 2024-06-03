package main

import "os"

// TODO: convert to use slog
func nemoLog(log string) {
	f, err := os.OpenFile("nemo.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(log); err != nil {
		panic(err)
	}
}
