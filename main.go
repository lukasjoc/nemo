package main

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/assets"
)

type layer struct {
	x     int
	y     int
	velo  int
	style tcell.Style
	shape []string
}

func drawShape(sc tcell.Screen, l layer) {
	sx := l.x
	for _, tile := range l.shape {
		if len(tile) == 0 {
			continue
		}
		for _, r := range tile {
			if l.velo > 0 {
				for i := sx - (l.velo) - 1; i < sx; i++ {
					sc.SetContent(i, l.y, ' ', nil, l.style)
				}
			}

			if l.velo < 0 {
				for i := (sx + len(tile)); i < (sx + len(tile) + -l.velo); i++ {
					sc.SetContent(i, l.y, ' ', nil, l.style)
				}
				// for i := sx + len(tile) - (l.velo); i < (sx + len(tile)); i++ {
				// }
			}

			// sc.SetContent(sx+len(tile), l.y, ' ', nil, l.style)
			sc.SetContent(l.x, l.y, r, nil, l.style)
			l.x++
		}
		l.x = sx
		l.y++
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

	normieR := assets.Load(assets.NormieR)
	normieL := assets.Load(assets.NormieL)
	layers := []*layer{}

	darkCyan := tcell.StyleDefault.Foreground(tcell.ColorLightCyan)
	darkPink := tcell.StyleDefault.Foreground(tcell.ColorLightPink)
	rand.Seed(time.Now().UnixNano())
	w, h := sc.Size()
	for i := 0; i < 10; i += 1 {
		l := layer{
			x:     rand.Intn(w/2-0) + 0,
			y:     rand.Intn(h - 1),
			velo:  []int{3, 1, 2}[rand.Intn(3)],
			style: []tcell.Style{darkCyan, darkPink}[rand.Intn(2)],
			shape: normieR,
		}
		layers = append(layers, &l)
	}

	for i := 0; i < 20; i += 1 {
		l := layer{
			x:     rand.Intn(w-w/2) + w/2,
			y:     int(rand.Intn(h - 1)),
			velo:  []int{-1, -2}[rand.Intn(2)],
			style: []tcell.Style{darkCyan, darkPink}[rand.Intn(2)],
			shape: normieL,
		}
		layers = append(layers, &l)
	}

	paused := make(chan bool, 1)

	render := func() {
		lastPaused := false
		for {
			select {
			case lastPaused = <-paused:
			default:
				if lastPaused {
					continue
				}
				for _, l := range layers {
					drawShape(sc, *l)
					l.x += l.velo
				}
				sc.Show()
				time.Sleep(time.Millisecond * 120)
			}
		}
	}

	go render()

	lastPaused := false

	for {
		ev := sc.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sc.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			}
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'p':
					lastPaused = !lastPaused
					select {
					case paused <- lastPaused:
					default:
					}
				}
			}
		}
	}
}
