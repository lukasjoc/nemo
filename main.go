package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
	"github.com/lukasjoc/nemo/internal/assets"
)

var (
	modeMask = flag.Bool("chac", true, "enables character color mode")
)

type message uint

const (
	renderStart message = iota
	renderPause
)

const renderTickDelay = time.Millisecond * 120
const restartDelay = time.Millisecond * 50

var (
	fgBluePallete = []tcell.Style{
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightBlue),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSkyBlue),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSteelBlue),
	}
	fgPallete = append([]tcell.Style{
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorOrchid),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGoldenrod),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleTurquoise),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleVioletRed),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPapayaWhip),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPeachPuff),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCoral),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCyan),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGoldenrodYellow),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGray),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightPink),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSalmon),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSeaGreen),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSlateGray),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightYellow),
		tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLimeGreen),
	}, fgBluePallete...)

	bodypartColorMask = func(ch rune) tcell.Style {
		style := tcell.StyleDefault.Dim(true).Bold(true)
		switch ch {
		case '\\', '/', '#', '~', '-', '_', '<', '(', ')':
			return internal.Choose([]tcell.Style{
				style.Foreground(tcell.ColorLightYellow),
				style.Foreground(tcell.ColorLightGreen),
				style.Foreground(tcell.ColorLightBlue),
			}...)
		case 'C', '@', 'o':
			return style.Foreground(tcell.ColorPaleVioletRed)
		case ',', '"', '\'', ';', ':', '=':
			return style.Foreground(tcell.ColorLightCoral)
		}
		return style
	}
)

var initialSwarmSize = 32

type layer struct {
	x          int
	y          int
	velo       int
	hidden     bool
	style      tcell.Style
	asset      assets.Asset
	assetIndex int
	// NOTE: that the drawFunc doesnt actually update the screen
	// it just computes the next layer. Its up to the renderer to sync
	// the changes to the screen. This effectively allows for double buffering.
	drawFunc func(l *layer, sc tcell.Screen)
}

func (l layer) String() string {
	return fmt.Sprintf("x:%4d y:%4d velo:%4d hidden:%6t group:%6s",
		l.x, l.y, l.velo, l.hidden, l.asset.Group)
}

func (l *layer) setDrawFunc(f func(l *layer, sc tcell.Screen)) {
	l.drawFunc = f
}

func fishDrawFunc(l *layer, sc tcell.Screen) {
	drawW, _ := sc.Size()
	initialX := l.x
	initialY := l.y
	ty := initialY
	for _, tile := range l.asset.Sources[l.assetIndex] {
		tlen := len(tile)
		if tlen == 0 {
			continue
		}
		tx := initialX
		for _, r := range tile {
			// clear any garbage from the previous draw
			if l.velo > 0 {
				for i := initialX - (l.velo) - 1; i < initialX; i++ {
					sc.SetContent(i, ty, ' ', nil, tcell.StyleDefault)
				}
			}
			if l.velo < 0 {
				for i := (initialX + tlen); i < (initialX + tlen + -l.velo); i++ {
					sc.SetContent(i, ty, ' ', nil, tcell.StyleDefault)
				}
			}
			// draw space in default color to not leave any (invisible) trails
			if unicode.IsSpace(r) {
				sc.SetContent(tx, ty, r, nil, tcell.StyleDefault)
			} else {
				if *modeMask {
					sc.SetContent(tx, ty, r, nil, bodypartColorMask(r))
				} else {
					sc.SetContent(tx, ty, r, nil, l.style)
				}
			}
			tx++
		}
		ty++
	}
	if l.velo > 0 && l.x > drawW+l.asset.Width ||
		l.velo < 0 && l.x < -l.asset.Width {
		(*l).hidden = true
	}
	(*l).x += l.velo
}

func newRandomFish(w int, h int) *layer {
	asset := assets.Random("fish")
	l := layer{
		velo:       internal.Choose(2, 1, 3),
		style:      internal.Choose(fgPallete...),
		asset:      asset,
		assetIndex: internal.Choose(0, 1),
	}
	leftSide := l.assetIndex == 0
	if leftSide {
		l.x = -(internal.IntRand((asset.Width*8)-asset.Width) + asset.Width)
		l.y = internal.IntRand(h - asset.Height)
	} else {
		l.x = (internal.IntRand((w+asset.Width*8)-(w+asset.Width)) + w + asset.Width)
		l.y = internal.IntRand(h - asset.Height)
		l.velo *= -1
	}
	l.setDrawFunc(fishDrawFunc)
	return &l
}

func newSwarm(w int, h int, swarmSize int) []*layer {
	swarm := []*layer{}
	for i := 0; i < swarmSize; i++ {
		swarm = append(swarm, newRandomFish(w, h))
	}
	return swarm
}

func bubbleDrawFunc(l *layer, sc tcell.Screen) {
	_, drawH := sc.Size()
	(*l).asset = assets.Random("bubble")
	initialX := l.x
	initialY := l.y
	ty := initialY
	for _, tile := range l.asset.Sources[l.assetIndex] {
		tlen := len(tile)
		if tlen == 0 {
			continue
		}
		tx := initialX
		for _, r := range tile {
			// TODO: dont rely on asset size implicitly
			sc.SetContent(tx, ty-l.velo, ' ', nil, tcell.StyleDefault)
			if !unicode.IsSpace(r) {
				sc.SetContent(tx, ty, r, nil, l.style)
			} else {
				sc.SetContent(tx, ty, r, nil, tcell.StyleDefault)
			}
			tx++
		}
		ty++
	}
	if l.y > drawH+l.asset.Height {
		(*l).hidden = true
	}
	(*l).y += l.velo
}

func newRandomBubble(w int, h int) *layer {
	asset := assets.Random("bubble")
	l := layer{
		velo:       -internal.Choose(3, 2, 4, 5),
		style:      internal.Choose(fgBluePallete...),
		asset:      asset,
		x:          internal.IntRand(w),
		y:          internal.IntRand(h / 2),
		assetIndex: 0,
	}
	l.setDrawFunc(bubbleDrawFunc)
	return &l
}

func render(messages <-chan message, sc tcell.Screen, layers *[]*layer) {
	nameStyle := internal.Choose(fgPallete...)
	lastMessage := renderStart
	sc.Clear()
loop:
	for {
		select {
		case lastMessage = <-messages:
			if lastMessage == renderPause {
				break loop
			}
		default:
			renderW, renderH := sc.Size()

			name := `	
  ___  ___ __ _  ___
 / _ \/ -_)  ' \/ _ \
/_//_/\__/_/_/_/\___/ 1.0`
			nameTiles := strings.Split(name, "\n")
			nameX := renderW - len(nameTiles[len(nameTiles)-1]) - 1
			nameY := renderH - len(nameTiles) - 1
			for _, tile := range nameTiles {
				rx := nameX
				for _, r := range tile {
					sc.SetContent(rx, nameY, r, nil, nameStyle)
					rx++
				}
				nameY++
			}
			bubbles := []*layer{}

			hiddenFish := 0
			*layers = slices.DeleteFunc(*layers, func(l *layer) bool {
				if l.hidden {
					if l.asset.Group == "fish" {
						hiddenFish++
					}
					return true
				}
				return false
			})
			*layers = append(*layers, newSwarm(renderW, renderH, hiddenFish)...)

			for _, l := range *layers {
				bx := 0
				if l.velo < 0 {
					bx = l.x + 1
				} else {
					bx = l.x + l.asset.Width
				}
				if bx > 0 && (bx%(renderW/4) == 0) {
					b := newRandomBubble(renderW, renderH)
					b.x = bx
					b.y = l.y - 1
					bubbles = append(bubbles, b)
				}
			}
			*layers = append(*layers, bubbles...)

			for _, l := range *layers {
				internal.Logln("LAYER DRAW %v", l)
				l.drawFunc(l, sc)
			}
			sc.Show()
			time.Sleep(renderTickDelay)
		}
	}
}

func main() {
	flag.Parse()

	// NOTE: For dev only via -tags=debug
	internal.LogCleanup()

	sc, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldnt create screen: %v\n", err)
		os.Exit(1)
	}

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
		os.Exit(1)
	}()

	if err := sc.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Couldnt init tcell: %v\n", err)
		os.Exit(1)
	}
	sc.SetStyle(tcell.StyleDefault)
	sc.Clear()

	initW, initH := sc.Size()
	layers := newSwarm(initW, initH, initialSwarmSize)

	messages := make(chan message, 1)
	go render(messages, sc, &layers)

	lastMessage := renderStart
	for {
		ev := sc.PollEvent()
		evW, evH := sc.Size()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			nextW, nextH := ev.Size()
			if nextW == initW && nextH == initH {
				continue
			}
			internal.Logln("RESIZE t:%d, w:%d, h:%d", ev.When().Unix(), evW, evH)
			select {
			case messages <- renderPause:
				lastMessage = renderPause
				time.Sleep(restartDelay)
				layers = newSwarm(evW, evH, initialSwarmSize)
				go render(messages, sc, &layers)
				lastMessage = renderStart
			default:
			}
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Key() == tcell.KeyCtrlC {
				return
			}
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'p':
					internal.Logln("KEY EVENT %s t:%d, w:%d, h:%d", ev.Name(), ev.When().Unix(), evW, evH)
					if lastMessage == renderPause {
						go render(messages, sc, &layers)
						lastMessage = renderStart
						continue
					}
					select {
					case messages <- renderPause:
						lastMessage = renderPause
					default:
					}
				case 'r':
					if lastMessage == renderPause {
						continue
					}
					internal.Logln("KEY EVENT %s t:%d, w:%d, h:%d", ev.Name(), ev.When().Unix(), evW, evH)
					select {
					case messages <- renderPause:
						time.Sleep(restartDelay)
						layers = newSwarm(evW, evH, initialSwarmSize)
						go render(messages, sc, &layers)
						lastMessage = renderStart
					default:
					}
				}
			}
		}
	}
}

// TODO:
// ?? V2
// the render func should not have direct access to Show of the screen
// but still be able to set the content on the screen
// make it prettier with more assets in the background (flora)
// Use Quadtree for storing 2d position data better and to determine
// the amount of fishies fitting without overlapping automatically.
//	 never spawn fishies overlapping each other
// 	 never spawn fishies directly above or behind other fishies
// 	 automatically decide swarmSize (based on the current width and height)
// More perf. analisys and improvements
