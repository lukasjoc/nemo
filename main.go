package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
	"unicode"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/assets"
)

type layer struct {
	x       int
	y       int
	velo    int
	visible bool
	style   tcell.Style
	shape   []string
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

const renderTickDelay = time.Millisecond * 120

var (
	fgCyan        = tcell.StyleDefault.Foreground(tcell.ColorLightCyan)
	fgRed         = tcell.StyleDefault.Foreground(tcell.ColorRed)
	fgGreen       = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	fgYellow      = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	fgPallete     = []tcell.Style{fgCyan, fgRed, fgGreen, fgYellow}
	fgPalleteSize = len(fgPallete)
)

func randomFish() ([]assets.Tiles, error) {
	a := []assets.Tiles{
		assets.Nemo,
		assets.NemoJr,
		assets.Runner,
		assets.AQ0,
		assets.AQ1,
	}[rand.Intn(5)]
	return assets.LoadTiles(a)
}

func setupRandomSwarm(w int, h int) []*layer {
	layers := []*layer{}
	rand.Seed(time.Now().UnixNano())
	swarmCount := 25
	for i := 0; i < swarmCount; i += 1 {
		tiles, _ := randomFish()
		// setup swarm coming from the left side
		left := layer{
			x:       rand.Intn(w/2-0) + 0,
			y:       rand.Intn(h - 1),
			velo:    []int{3, 1, 2}[rand.Intn(3)],
			style:   fgPallete[rand.Intn(fgPalleteSize)],
			shape:   tiles[0],
			visible: true,
		}
		// setup swarm coming from the right side
		right := layer{
			x:       rand.Intn(w-w/2) + w/2,
			y:       int(rand.Intn(h - 1)),
			velo:    []int{-1, -2}[rand.Intn(2)],
			style:   fgPallete[rand.Intn(fgPalleteSize)],
			shape:   tiles[1],
			visible: true,
		}
		layers = append(layers, &left, &right)
	}
	return layers
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
	layers := setupRandomSwarm(w-1, h-1)

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
					if l.velo < 0 && l.x <= 0 ||
						l.velo > 0 && l.x >= w {
						l.visible = false
					}
					l.x += l.velo
				}
				sc.Show()
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
