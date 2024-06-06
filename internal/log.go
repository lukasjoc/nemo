//go:build !debug
// +build !debug

package internal

var logFileName = "DUMMY"

// TODO: convert to use slog
// NOTE: Remember.. Logging is not cheap!
func Logln(s string, args ...any) {}

func LogCleanup() {}
