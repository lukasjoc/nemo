package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
	"unicode"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
	"github.com/lukasjoc/nemo/internal/assets"
)

type message uint

const (
	renderStart message = iota
	renderPause
	renderHalt
)

const renderTickDelay = time.Millisecond * 120

var (
	fgPallete = []tcell.Style{
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorOrchid),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGoldenrod),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleTurquoise),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleVioletRed),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPapayaWhip),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPeachPuff),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightBlue),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCoral),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCyan),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGoldenrodYellow),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGray),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightPink),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSalmon),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSeaGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSkyBlue),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSlateGray),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSteelBlue),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightYellow),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLimeGreen),
	}
)

var initialSwarmSize = 32

type layer struct {
	id     string
	x      int
	y      int
	velo   int
	hidden bool
	style  tcell.Style
	asset  assets.Asset
	tiles  []string
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

func newRandomWithTiles(asset assets.Asset, assetIndex int, x int, y int) *layer {
	return &layer{
		id:    fmt.Sprintf(`layer-%d`, time.Now().Unix()),
		x:     x,
		y:     y,
		velo:  internal.Choose(5, 4, 3, 1, 2, 6),
		style: internal.Choose(fgPallete...),
		asset: asset,
		tiles: asset.Sources[assetIndex],
	}
}

func newRandomBatch(w, h int, batchSize int) []*layer {
	batch := []*layer{}
	for i := 0; i < batchSize; i++ {
		asset := assets.Random()
		side := internal.Choose(0, 1)
		var l *layer = nil
		if side == 0 {
			// setup swarm coming from the left side
			assetIndex := 0
			lx := (rand.Intn((asset.Width*8)-asset.Width) + asset.Width) * -1
			ly := rand.Intn(h - asset.Height)
			l = newRandomWithTiles(asset, assetIndex, lx, ly)
		} else {
			// setup swarm coming from the right side
			assetIndex := 1
			rx := (rand.Intn((w+asset.Width*8)-(w+asset.Width)) + w + asset.Width)
			ry := rand.Intn(h - asset.Height)
			l = newRandomWithTiles(asset, assetIndex, rx, ry)
			// NOTE: make sure to invert the velo to get correct direction for tiles
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

func render(messages <-chan message, sc tcell.Screen, layers *[]*layer) {
	lastMessage := renderStart
	sc.Clear()
loop:
	for {
		select {
		case lastMessage = <-messages:
			if lastMessage == renderPause ||
				lastMessage == renderHalt {
				break loop
			}
		default:
			renderW, renderH := sc.Size()
			// render each layer into the tcell buffer before calling
			// show to reduce flickering, especially when they collide.
			for _, l := range *layers {
				drawLayer(sc, *l)
				if l.velo > 0 && l.x >= renderW+l.asset.Width ||
					l.velo < 0 && l.x < -l.asset.Width {
					l.hidden = true
				}
				l.x += l.velo
			}
			// show the rendered results
			sc.Show()

			// do some cleanup before continuing to next frame
			hidden := 0
			*layers = slices.DeleteFunc(*layers, func(l *layer) bool {
				if l.hidden {
					hidden++
					return true
				}
				return false
			})
			*layers = append(*layers, newRandomBatch(renderW, renderH, hidden)...)
			time.Sleep(renderTickDelay)
		}
	}
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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		quit()
		os.Exit(0)
	}()

	initW, initH := sc.Size()
	layers := []*layer{}
	layers = append(layers, newRandomBatch(initW, initH, initialSwarmSize)...)

	messages := make(chan message, 1)
	go render(messages, sc, &layers)

	lastMessage := renderStart
	for {
		ev := sc.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			nextW, nextH := ev.Size()
			// t := ev.When().Unix()
			// internal.Log(fmt.Sprintf("RESIZE(%d): REV: %d %d => %d %d\n", t, initW, initH, nextW, nextH))
			if nextW != initW || nextH != initH {
				if lastMessage != renderStart {
					layers = append([]*layer{}, newRandomBatch(nextW, nextH, initialSwarmSize)...)
					go render(messages, sc, &layers)
					lastMessage = renderStart
					continue
				}
				select {
				case messages <- renderHalt:
					layers = append([]*layer{}, newRandomBatch(nextW, nextH, initialSwarmSize)...)
					go render(messages, sc, &layers)
					lastMessage = renderStart
				default:
				}
			}
			sc.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Key() == tcell.KeyCtrlC {
				return
			}
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'p':
					// continue where left off
					if lastMessage == renderPause {
						go render(messages, sc, &layers)
						lastMessage = renderStart
						continue
					}
					// stop the renderer but keep screen and layers intact
					if lastMessage == renderStart {
						select {
						case messages <- renderPause:
							lastMessage = renderPause
						default:
						}
					}
				case 'r':
					if lastMessage != renderStart {
						continue
					}
					select {
					case messages <- renderHalt:
						// t := ev.When().Unix()
						// name := ev.Name()
						evW, evH := sc.Size()
						// internal.Log(fmt.Sprintf("KEY(%s, %d): %d %d\n", name, t, evW, evH))
						layers = append([]*layer{}, newRandomBatch(evW, evH, initialSwarmSize)...)
						go render(messages, sc, &layers)
						lastMessage = renderStart
						sc.Sync()
					default:
					}
				}
			}
		}
	}
}

// TODO:
// bubbles
// make it prettier with more assets in the background
// simple cli for average,min,max velocity and refresh rate, monotone etc.
