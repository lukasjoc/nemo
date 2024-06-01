package main

import (
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/assets"
)

func drawShape(sc tcell.Screen, shape []string, x int, y int, style tcell.Style) {
	start := x
	for _, tile := range shape {
		if len(tile) == 0 {
			continue
		}
		for _, r := range tile {
			sc.SetContent(start-1, y, ' ', nil, style)
			sc.SetContent(x, y, r, nil, style)
			x++
		}
		x = start
		y++
	}
}

func main() {
	sc, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := sc.Init(); err != nil {
		panic(err)
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	sc.SetStyle(defStyle)
	sc.Clear()

	normie := strings.Split(assets.Normie, "\n")

	quit := func() {
		p := recover()
		sc.Fini()
		if p != nil {
			panic(p)
		}
	}
	defer quit()

	go func() {
		dx := 10
		dy := 20
		for {
			drawShape(sc, normie, dx, dy, defStyle)
			sc.Show()
			w, _ := sc.Size()
			if dx+1 > w {
				break
			}
			dx += 1
			time.Sleep(time.Millisecond * 100)
		}
	}()


	go func() {
		dx := 0
		dy := 0
		for {
			drawShape(sc, normie, dx, dy, defStyle)
			sc.Show()
			w, _ := sc.Size()
			if dx+1 > w {
				break
			}
			dx++
			time.Sleep(time.Millisecond * 300)
		}
	}()

	for {
		ev := sc.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sc.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			}
		}
	}
}
