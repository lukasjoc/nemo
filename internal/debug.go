package internal

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	DebugEnabled = os.Getenv("NEMO_DEBUG") == "1"
	ProfEnabled  = os.Getenv("NEMO_PROF") == "1"
)

var (
	// the log debug log path if you enable it with `NEMO_DEBUG=1`
	logFileName = "nemo.log"
	// a cpu profile if you enable it with `NEMO_PROF=1`
	cpuprof *os.File
	// a mem/heap profile if you enbable it with `NEMO_PROF=1`
	memprof *os.File
)

func DebugStart() {
	if DebugEnabled {
		os.Remove(logFileName)
	}
	if ProfEnabled {
		cpuprof, _ = os.Create("cpu.prof")
		memprof, _ = os.Create("mem.prof")
		if err := pprof.StartCPUProfile(cpuprof); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
}

func DebugEnd() {
	if ProfEnabled {
		defer cpuprof.Close()
		pprof.StopCPUProfile()
		defer memprof.Close()
		runtime.GC() // get up-to-date statistics
		pprof.WriteHeapProfile(memprof)
	}
}

// NOTE: Remember.. Logging is not cheap!
// TODO: use slog instead
func Logln(s string, args ...any) {
	if !DebugEnabled || ProfEnabled {
		return
	}
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(
		fmt.Sprintf("%s\n", fmt.Sprintf(s, args...))); err != nil {
		panic(err)
	}
}
