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
	x      int
	y      int
	velo   int
	hidden bool
	style  tcell.Style
	asset  assets.Asset
	// TODO: dont store tiles additionally to the asset
	// just store chosen index and change callsites where needed
	tiles []string
	// NOTE: that the drawFunc doesnt actually update the screen
	// it just computes the next layer. Its up to the renderer to sync
	// the changes to the screen. This effectively allows for double buffering.
	drawFunc func(l *layer, sc tcell.Screen)
}

func (l layer) String() string {
	return fmt.Sprintf("x:%3d y:%3d velo:%3d hidden:%t", l.x, l.y, l.velo, l.hidden)
}

func (l *layer) setDrawFunc(f func(l *layer, sc tcell.Screen)) {
	l.drawFunc = f
}

func drawFishLayer(sc tcell.Screen, l layer) {
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

var velocityRange = []int{5, 4, 3, 1, 2, 6}

func newRandomBatch(w, h int, batchSize int) []*layer {
	batch := []*layer{}
	for i := 0; i < batchSize; i++ {
		asset := assets.Random()
		side := internal.Choose(0, 1)
		tiles := asset.Sources[side]
		velo := internal.Choose(velocityRange...)
		style := internal.Choose(fgPallete...)
		var l *layer = nil
		if side == 0 {
			// setup swarm coming from the left side
			l = &layer{
				x:     (rand.Intn((asset.Width*8)-asset.Width) + asset.Width) * -1,
				y:     rand.Intn(h - asset.Height),
				velo:  velo,
				style: style,
				asset: asset,
				tiles: tiles,
			}
		} else {
			// setup swarm coming from the right side
			l = &layer{
				x: (rand.Intn((w+asset.Width*8)-(w+asset.Width)) + w + asset.Width),
				y: rand.Intn(h - asset.Height),
				// NOTE: make sure to invert the velo to get correct direction
				velo:  velo * -1,
				style: style,
				asset: asset,
				tiles: tiles,
			}
		}
		if l == nil {
			// unreachable: just for sanity reasons
			panic("random layer was expected but not generated")
		}
		l.setDrawFunc(func(l *layer, sc tcell.Screen) {
			drawW, _ := sc.Size()
			drawFishLayer(sc, *l)
			if l.velo > 0 && l.x >= drawW+l.asset.Width ||
				l.velo < 0 && l.x < -l.asset.Width {
				l.hidden = true
			}
			l.x += l.velo
		})
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
			for _, l := range *layers {
				internal.Logln("LAYER DRAW %v", l)
				l.drawFunc(l, sc)
			}
			sc.Show()

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

	// NOTE: For dev only via -tags=debug
	internal.LogCleanup()

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
			t := ev.When().Unix()
			internal.Logln("RESIZE t:%d, w:%d h:%d -> w:%d h:%d", t, initW, initH, nextW, nextH)
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
						t := ev.When().Unix()
						name := ev.Name()
						evW, evH := sc.Size()
						internal.Logln("RESIZE name:%s, t:%d, w:%d, h:%d", name, t, evW, evH)
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
// bubbles drawFunc and layer O o .
// make it prettier with more assets in the background
// simple cli for average,min,max velocity and refresh rate, monotone etc.
