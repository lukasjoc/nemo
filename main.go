package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/assets"
)

func drawShape(sc tcell.Screen, shape []string, x int, y int, style tcell.Style) {
	sx := x
	for _, tile := range shape {
		if len(tile) == 0 {
			continue
		}
		for _, r := range tile {
			sc.SetContent(sx-1, y, ' ', nil, style)
			sc.SetContent(sx+len(tile), y, ' ', nil, style)
			sc.SetContent(x, y, r, nil, style)
			x++
		}
		x = sx
		y++
	}
}

type layer struct {
	x     int
	y     int
	shape []string
	velo  int
}

func main() {
	sc, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := sc.Init(); err != nil {
		panic(err)
	}

	sc.SetStyle(tcell.StyleDefault)
	sc.Clear()

	quit := func() {
		p := recover()
		sc.Fini()
		if p != nil {
			panic(p)
		}
	}
	defer quit()

	normie := assets.Load(assets.Other)
	layers := []*layer{}

	w, _ := sc.Size()
	// l := layer{x: 0, y: 0, shape: normie, velo: 1}
	for i := 0; i < 35; i += 7 {
		l := layer{x: 0, y: i, shape: normie, velo: 2}
		layers = append(layers, &l)
	}
	for i := 0; i < 35; i += 7 {
		l := layer{x: w - 5, y: i, shape: normie, velo: -1}
		layers = append(layers, &l)
	}

	go func() {
		for {
			for _, l := range layers {
				drawShape(sc, l.shape, l.x, l.y, tcell.StyleDefault.Foreground(tcell.ColorLightCyan))
				l.x += l.velo
			}
			sc.Show()
			time.Sleep(time.Millisecond * 120)
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
