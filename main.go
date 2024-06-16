package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
)

var (
	chacMode   = flag.Bool("chac", true, "enables character color mode")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

const restartDelay = time.Millisecond * 50

var (
	fgBluePallete = []tcell.Style{
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightBlue),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSkyBlue),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSteelBlue),
	}
	fgPallete = append([]tcell.Style{
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorOrchid),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGoldenrod),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleTurquoise),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleVioletRed),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPapayaWhip),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPeachPuff),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCoral),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCyan),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGoldenrodYellow),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGray),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightPink),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSalmon),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSeaGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSlateGray),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightYellow),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLimeGreen),
	}, fgBluePallete...)

	bodypartColorMask = func(ch rune) tcell.Style {
		style := tcell.StyleDefault.Dim(true).Bold(true)
		switch ch {
		case '\\', '/', '#', '~', '-', '_', '<', '(', ')':
			return internal.Choose(
				style.Foreground(tcell.ColorLightYellow),
				style.Foreground(tcell.ColorLightGreen),
				style.Foreground(tcell.ColorLightBlue))
		case 'C', '@', 'o':
			return style.Foreground(tcell.ColorPaleVioletRed)
		case ',', '"', '\'', ';', ':', '=':
			return style.Foreground(tcell.ColorLightCoral)
		}
		return style
	}
)

func main() {
	flag.Parse()

	// NOTE: For dev only via -tags=debug
	internal.LogCleanup()

	// TODO: hide behind debug tag
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// TODO: should the renderer create the screen automatically?
	sc, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldnt create screen: %v\n", err)
		os.Exit(1)
	}

	if err := sc.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Couldnt init tcell: %v\n", err)
		os.Exit(1)
	}
	sc.SetStyle(tcell.StyleDefault)
	sc.Clear()

	r := newRenderer(&rendererConfig{sc, 12})

	quit := func() {
		p := recover()
		r.stop()
		r.destroy()
		sc.Fini()
		if p != nil {
			panic(p)
		}
		// TODO: hide behind debug tag
		if *memprofile != "" {
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			defer f.Close() // error handling omitted for example
			runtime.GC()    // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
		}
	}
	defer quit()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		quit()
		os.Exit(1)
	}()

	r.seed()
	r.start()

	initW, initH := sc.Size()

	// TODO: hide this in a r.poll()
	for {
		ev := sc.PollEvent()
		evW, evH := sc.Size()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			nextW, nextH := ev.Size()
			if nextW == initW && nextH == initH {
				continue
			}
			r.restart()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Key() == tcell.KeyCtrlC {
				fmt.Println("CTRL-C Received!!")
				return
			}
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'p':
					internal.Logln("KEY EVENT %s t:%d, w:%d, h:%d", ev.Name(), ev.When().Unix(), evW, evH)
					select {
					case <-r.stopped:
						r.start()
					default:
						r.stop()
					}
				case 'r':
					r.restart()
				}
			}
		}
	}
}

// TODO:
// More perf. analisys and improvements

// ?? V2
// make it prettier with more assets in the background (flora)
// Use Quadtree for storing 2d position data better and to determine
// the amount of fishies fitting without overlapping automatically.
//	 never spawn fishies overlapping each other
// 	 never spawn fishies directly above or behind other fishies
// 	 automatically decide swarmSize (based on the current width and height)
