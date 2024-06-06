//go:build debug
// +build debug

package internal

import (
	"fmt"
	"os"
)

var logFileName = "nemo.log"

// TODO: convert to use slog
// NOTE: Remember.. Logging is not cheap!
func Logln(s string, args ...any) {
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(
		fmt.Sprintf("%s\n", fmt.Sprintf(s, args...))); err != nil {
		panic(err)
	}
}

func LogCleanup() {
	os.Remove(logFileName)
}
