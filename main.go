package main

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"time"
	"unicode"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/assets"
)

type message uint

const (
	renderStart message = iota
	renderPause
	renderHalt
)

const renderTickDelay = time.Millisecond * 120

var (
	fgCyan    = tcell.StyleDefault.Foreground(tcell.ColorLightCyan)
	fgRed     = tcell.StyleDefault.Foreground(tcell.ColorRed)
	fgGreen   = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	fgYellow  = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	fgOrange  = tcell.StyleDefault.Foreground(tcell.ColorOrange)
	fgPurple  = tcell.StyleDefault.Foreground(tcell.ColorPurple)
	fgPallete = []tcell.Style{fgCyan, fgRed, fgGreen, fgYellow, fgOrange,
		fgPurple}
	fgPalleteSize = len(fgPallete)
)

var screenMargin = 15
var initialSwarmSize = 45

type layer struct {
	id      int64
	x       int
	y       int
	velo    int
	visible bool
	style   tcell.Style
	tiles   assets.Tiles
}

func drawLayer(sc tcell.Screen, l layer) {
	sx := l.x
	for _, tile := range l.tiles {
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

func randomFish() ([]assets.Tiles, error) {
	// TODO: shouldnt load them everytime (cache them somehow in a package var)
	a := []assets.Tiles{
		assets.Nemo,
		assets.NemoJr,
		assets.Runner,
		assets.AQ0,
		assets.AQ1,
		// TODO: more tiles/fishies
	}[rand.Intn(5)]
	return assets.LoadTiles(a)
}

func choose[T comparable](selection ...T) T {
	return selection[rand.Intn(len(selection))]
}

func newRandomWithTiles(tiles assets.Tiles, x int, y int) *layer {
	return &layer{
		id: time.Now().Unix(),
		x:  x,
		y:  y,
		// TODO: better way to handle velocity (colission based/dynamic)
		velo:    choose(3, 4, 2, 10, 6, 2, 3, 1, 7, 6, 3, 3),
		style:   choose(fgPallete...),
		tiles:   tiles,
		visible: true,
	}
}

func newRandomBatch(w, h int, batchSize int) []*layer {
	batch := []*layer{}
	for i := 0; i < batchSize; i++ {
		tiles, _ := randomFish()
		side := choose(0, 1)
		// TODO: clean this up (magic variables, ugly AF)
		var l *layer = nil
		if side == 0 {
			// setup swarm coming from the left side
			lx := (rand.Intn(screenMargin*8-screenMargin) + screenMargin) * -1
			ly := rand.Intn(h - 1)
			l = newRandomWithTiles(tiles[0], lx, ly)
		} else {
			// setup swarm coming from the right side
			rx := (rand.Intn(w+screenMargin*8-w+screenMargin) + w + screenMargin)
			ry := int(rand.Intn(h - 1))
			l = newRandomWithTiles(tiles[1], rx, ry)
			// NOTE: make sure to invert the velo to get correct direction
			// for tiles
			l.velo *= -1
		}
		if l == nil {
			// unreachable: just for sanity reasons
			panic("random layer was expected but not generated")
		}
		batch = append(batch, l)
	}
	return batch
}

func main() {
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

	quit := func() {
		p := recover()
		sc.Fini()
		if p != nil {
			panic(p)
		}
	}
	defer quit()

	w, h := sc.Size()
	layers := []*layer{}
	layers = append(layers, newRandomBatch(w-1, h-1, initialSwarmSize)...)

	messages := make(chan message, 1)

	render := func(w int, h int) {
		lastMessage := renderStart
	rendering:
		for {
			select {
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
					if l.velo > 0 && l.x >= w+screenMargin ||
						l.velo < 0 && l.x < -screenMargin {
						l.visible = false
					}
					l.x += l.velo
				}
				sc.Show()

				// delete the hidden layers before next tick
				deleted := 0
				layers = slices.DeleteFunc(layers, func(l *layer) bool {
					if l.visible == false {
						deleted++
						return true
					}
					return false
				})
				layers = append(layers, newRandomBatch(w-1, h-1, deleted)...)
				time.Sleep(renderTickDelay)
			}
		}
	}

	go render(w-1, h-1)

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

// TODO:
// restart rendering on 'r' and when screen resizes
// add color mask on ascii fishies to make them colorful
// simple cli for average,min,max velocity and refresh rate, monotone etc.
