package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
)

const renderTickDelay = time.Millisecond * 120

type rendererConfig struct {
	sc        tcell.Screen
	swarmSize int
}

type renderer struct {
	// TODO: probably need to `recover()` all the errors from the render
	// workers into a errs channel and report better
	//mu        sync.RWMutex
	t         *time.Ticker
	done      chan bool
	stopped   chan bool
	sc        tcell.Screen
	w         int
	h         int
	swarmSize int
	nameStyle tcell.Style
	swarm     []*layer
	bubbles   []*layer
}

func (r *renderer) stop() {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	r.t.Stop()
	go func() { r.done <- true }()
	go func() { r.stopped <- true }()
	r.renderStats(time.Now())
	r.sc.Show()
}

func (r *renderer) start() {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	r.t.Reset(renderTickDelay)
	go r.render()
}
func (r *renderer) restart() {
	r.stop()
	<-r.stopped
	r.refresh()
	r.seed()
	r.start()
}

func (r *renderer) destroy() {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	r.swarm = nil
	r.bubbles = nil
}

func (r *renderer) refresh() {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	w, h := r.sc.Size()
	r.w = w
	r.h = h
	r.sc.Clear()
}

//func (r *renderer) clean() {
//	hiddenFish := 0
//	*layers = slices.DeleteFunc(*layers, func(l *layer) bool {
//		if l.hidden {
//			if l.asset.Group == "fish" {
//				hiddenFish++
//			}
//			return true
//		}
//		return false
//	})
//	*layers = append(*layers, newSwarm(renderW, renderH, hiddenFish)...)
//}

func (r *renderer) seed() {
	r.destroy()
	r.refresh()
	//r.mu.Lock()
	//defer r.mu.Unlock()
	r.nameStyle = internal.Choose(fgPallete...)
	r.swarm = make([]*layer, r.swarmSize)
	// TODO: clean out swarm
	for i := 0; i < r.swarmSize; i++ {
		r.swarm[i] = newRandomFish(r.w, r.h)
	}
	r.bubbles = make([]*layer, r.swarmSize)
	// NOTE: the bubbles will be created and rendered as the fish moves
	// and the x,y of the fish is known..
}

var nameRaw = `	
  ___  ___ __ _  ___
 / _ \/ -_)  ' \/ _ \
/_//_/\__/_/_/_/\___/ 1.0`

var nameTiles = strings.Split(nameRaw, "\n")

// TODO: i should have a more generic function that can render a bunch of
// bytes to a x,y,w,h
// func (r *renderer) renderText(x int, y int, w int, h int, text string) { }

func (r *renderer) renderName() {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	nameX := r.w - len(nameTiles[len(nameTiles)-1]) - 4
	nameY := r.h - len(nameTiles) - 1
	for _, tile := range nameTiles {
		rx := nameX
		for _, ch := range tile {
			r.sc.SetContent(rx, nameY, ch, nil, r.nameStyle)
			rx++
		}
		nameY++
	}
}

func (r *renderer) renderStats(ts time.Time) {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	// TODO: render stats
	fishCount := 0
	bubbleCount := 0
	for _, l := range r.swarm {
		if l != nil {
			fishCount++
		}
	}
	for _, l := range r.bubbles {
		if l != nil {
			bubbleCount++
		}
	}
	stats := fmt.Sprintf("TS: %d\nFish: %2d\nBubbles: %2d", ts.Unix(), fishCount, bubbleCount)
	statsTiles := strings.Split(stats, "\n")
	nameX := r.w - len(statsTiles[len(statsTiles)-1]) - (1)
	nameY := 0 + len(statsTiles) - 1
	for _, tile := range statsTiles {
		rx := nameX
		for _, ch := range tile {
			r.sc.SetContent(rx, nameY, ch, nil, tcell.StyleDefault)
			rx++
		}
		nameY++
	}
}

func (r *renderer) renderBubbles() {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	if r.bubbles == nil {
		// TODO: what do do here.. should not happen
		return
	}
	for i, l := range r.swarm {
		bx := 0
		if l.velo < 0 {
			bx = l.x + 1
		} else {
			bx = l.x + l.asset.Width
		}
		if bx > 0 && (bx%(r.w/4) == 0) {
			// TODO: move the layer stuff into here
			b := newRandomBubble(r.w, r.h)
			b.x = bx
			b.y = l.y - 1
			r.bubbles[i] = b
		}
	}
	for _, l := range r.bubbles {
		if l == nil {
			continue
		}
		l.drawFunc(l, r.sc)
	}
}

func (r *renderer) renderSwarm() {
	// r.mu.Lock()
	// defer r.mu.Unlock()
	// TODO: clean out hidden layers
	//for i := 0; i < r.swarmSize; i++ {
	//	if r.swarm[i] == nil {
	//		continue
	//	}
	//	if r.swarm[i].hidden {
	//		r.swarm[i] = nil
	//	}
	//}
	for _, l := range r.swarm {
		if l == nil {
			continue
		}
		internal.Logln("LAYER DRAW %v", l)
		l.drawFunc(l, r.sc)
	}
}

// TODO: i should have a draw loop and a update loop with different
// tick delays. I think that would make it even smoother.
func (r *renderer) render() {
	// defer r.wg.Done()
	for {
		select {
		case <-r.done:
			return
		case ts := <-r.t.C:
			r.renderName()
			r.renderSwarm()
			r.renderBubbles()
			r.renderStats(ts)
			r.sc.Show()
		}
	}
}

func newRenderer(config *rendererConfig) *renderer {
	ticker := time.NewTicker(renderTickDelay)
	r := renderer{sc: config.sc, t: ticker, done: make(chan bool), stopped: make(chan bool),
		swarmSize: config.swarmSize, swarm: nil, bubbles: nil}
	return &r
}
