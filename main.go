package main

import (
	"math/rand"
	"time"
	"unicode"

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

func drawLayer(sc tcell.Screen, l layer) {
	sx := l.x
	for _, tile := range l.shape {
		if len(tile) == 0 {
			continue
		}
		for _, r := range tile {
			// clearing the garbage from the last transforms
			if l.velo > 0 {
				for i := sx - (l.velo) - 1; i < sx; i++ {
					sc.SetContent(i, l.y, ' ', nil, tcell.StyleDefault)
				}
			}
			if l.velo < 0 {
				for i := (sx + len(tile)); i < (sx + len(tile) + -l.velo); i++ {
					sc.SetContent(i, l.y, ' ', nil, tcell.StyleDefault)
				}
			}
			// draw space in default color to be a nice citizen of terminalland
			if unicode.IsSpace(r) {
				sc.SetContent(l.x, l.y, r, nil, tcell.StyleDefault)
			} else {
				sc.SetContent(l.x, l.y, r, nil, l.style)
			}
			l.x++
		}
		l.x = sx
		l.y++
	}
}

type message uint

const (
	renderStart message = iota
	renderPause
	renderHalt
)

const renderFPS = time.Millisecond * 120

var (
	fgCyan        = tcell.StyleDefault.Foreground(tcell.ColorRed)
	fgGreen       = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	fgYellow      = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	fgPallete     = []tcell.Style{fgCyan, fgGreen, fgYellow}
	fgPalleteSize = len(fgPallete)
)

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

	// normieR := assets.Load(assets.NormieR)
	other := assets.Load(assets.Other)
	normieL := assets.Load(assets.NormieL)
	layers := []*layer{}

	rand.Seed(time.Now().UnixNano())
	w, h := sc.Size()
	for i := 0; i < 20; i += 1 {
		l := layer{
			x:     rand.Intn(w/2-0) + 0,
			y:     rand.Intn(h - 1),
			velo:  []int{3, 1, 2}[rand.Intn(3)],
			style: fgPallete[rand.Intn(fgPalleteSize)],
			shape: other,
		}
		layers = append(layers, &l)
	}

	for i := 0; i < 20; i += 1 {
		l := layer{
			x:     rand.Intn(w-w/2) + w/2,
			y:     int(rand.Intn(h - 1)),
			velo:  []int{-1, -2}[rand.Intn(2)],
			style: fgPallete[rand.Intn(fgPalleteSize)],
			shape: normieL,
		}
		layers = append(layers, &l)
	}

	messages := make(chan message, 1)

	render := func() {
		lastMessage := renderStart
	rendering:
		for {
			select {
			// try to read from the inbox
			case lastMessage = <-messages:
			default:
				// halt the rendering
				if lastMessage == renderHalt {
					break rendering
				}
				// pause the rendering
				if lastMessage == renderPause {
					continue
				}
				// render each layer into the tcell buffer before calling
				// show to reduce flickering, especially when they collide.
				for _, l := range layers {
					drawLayer(sc, *l)
					l.x += l.velo
				}
				sc.Show()
				time.Sleep(renderFPS)
			}
		}
	}

	go render()

	lastMessage := renderStart
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
					nextMessage := renderPause
					if lastMessage == renderPause {
						nextMessage = renderStart
					}
					select {
					case messages <- nextMessage:
						lastMessage = nextMessage
					default:
					}
				case 'h':
					select {
					case messages <- renderHalt:
						lastMessage = renderHalt
					default:
					}
				}
			}
		}
	}
}
