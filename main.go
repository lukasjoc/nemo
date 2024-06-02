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

func newRandomWithTiles(tiles assets.Tiles, x int, y int) *layer {
	return &layer{
		id:      time.Now().Unix(),
		x:       x,
		y:       y,
		velo:    []int{3, 4, 2, 10, 6, 2, 3, 1, 7, 6, 3, 3}[rand.Intn(3)],
		style:   fgPallete[rand.Intn(fgPalleteSize)],
		tiles:   tiles,
		visible: true,
	}
}

// func randomFishN(liv int, riv int, w int, h int) []*layer {
// 	batch := []*layer{}
// 	for i := 0; i < liv; i++ {
// 		tiles, _ := randomFish()
// 		x := rand.Intn(w/2-0) + 0
// 		y := rand.Intn(h - 1)
// 		l := newRandomWithTiles(tiles[0], x, y, w, h)
// 		batch = append(batch, l)
// 	}
// 	for i := 0; i < riv; i++ {
// 		tiles, _ := randomFish()
// 		x := rand.Intn(w-w/2) + w/2
// 		y := int(rand.Intn(h - 1))
// 		l := newRandomWithTiles(tiles[1], x, y, w, h)
// 		batch = append(batch, l)
// 	}
// 	return batch
// }

func setupRandomSwarm(w int, h int) []*layer {
	layers := []*layer{}
	swarmCount := 50
	for i := 0; i < swarmCount; i += 1 {
		tiles, _ := randomFish()
		// setup swarm coming from the left side
		//lx := -screenMargin
		lx := (rand.Intn(screenMargin*8-screenMargin) + screenMargin) * -1
		ly := rand.Intn(h - 1)
		left := newRandomWithTiles(tiles[0], lx, ly)
		// setup swarm coming from the right side
		// rx := w + screenMargin
		rx := (rand.Intn(w+screenMargin*8-w+screenMargin) + w + screenMargin)
		ry := int(rand.Intn(h - 1))
		right := newRandomWithTiles(tiles[1], rx, ry)
		right.velo *= -1
		layers = append(layers, left, right)
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

	// TODO: remove the offscreen layers
	// garbage := []struct{ layerId int64 }{}

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
				// simulate next x values to tell if we need to respawn
				// for _, l := range layers {
				// 	nextX := l.x + l.velo
				// 	if l.velo < 0 && nextX <= 0-offscreenPadding ||
				// 		l.velo > 0 && nextX >= w+offscreenPadding {
				// 		// enqueue the layer as garbage
				// 		garbage = append(garbage, struct {
				// 			layerId int64
				// 		}{l.id})
				// 	}
				// }

				//for _, g := range garbage {
				//	layers = slices.DeleteFunc(layers, func(l *layer) bool {
				//		return l.id == g.layerId
				//	})
				//	// TODO: spawn new random one
				//}
				// render each layer into the tcell buffer before calling
				// show to reduce flickering, especially when they collide.
				for _, l := range layers {
					drawLayer(sc, *l)
					// layer is not visible anymore and can be cleaned up
					// if l.velo > 0 && l.x <= w+screenMargin {
					// 	l.visible = false
					// 	liv++
					// }
					// if l.velo < 0 && l.x >= -screenMargin {
					// 	l.visible = false
					// 	riv++
					// }
					l.x += l.velo
				}
				// delete the invisible ones
				// layers = slices.DeleteFunc(layers, func(l *layer) bool {
				// 	f, err := os.OpenFile("nemo.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				// 	if err != nil {
				// 		panic(err)
				// 	}
				// 	line := fmt.Sprintf("DEL: %d %d\n", l.id, l.visible)
				// 	if _, err = f.WriteString(line); err != nil {
				// 		panic(err)
				// 	}
				// 	f.Close()
				// 	return l.visible == false
				// })
				sc.Show()

				// layers = append(layers, randomFishN(1, 1, w, h)...)

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
