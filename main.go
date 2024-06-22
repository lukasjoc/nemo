package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
	"github.com/lukasjoc/nemo/internal/renderer"
)

func main() {
	internal.DebugStart()

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

	r := renderer.New(sc, 18, renderer.DefaultTickDelay)
	quit := func() {
		p := recover()
		r.Stop()
		r.Destroy()
		sc.Fini()
		if p != nil {
			panic(p)
		}
	}
	defer quit()
	defer internal.DebugEnd()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		quit()
		os.Exit(1)
	}()

	r.Reset()
	r.Start()

	initW, initH := sc.Size()
	for {
		ev := sc.PollEvent()
		evW, evH := sc.Size()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			nextW, nextH := ev.Size()
			if nextW == initW && nextH == initH {
				continue
			}
			r.Restart()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Key() == tcell.KeyCtrlC {
				return
			}
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'p':
					internal.Logln("KEY EVENT %s t:%d, w:%d, h:%d", ev.Name(), ev.When().Unix(), evW, evH)
					select {
					case <-r.Stopped:
						r.Start()
					default:
						r.Stop()
					}
				case 'r':
					r.Restart()
				}
			}
		}
	}
}
