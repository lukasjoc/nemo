package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
)

const renderTickDelay = time.Millisecond * 120

type renderer struct {
	// TODO: probably need to `recover()` all the errors from the render
	// workers into a errs channel and report better
	mu        sync.RWMutex
	sc        tcell.Screen
	w         int
	h         int
	running   bool
	swarmSize int
	nameStyle tcell.Style
	swarm     []*layer
	bubbles   []*layer
}

func (r *renderer) stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running = false
}

func (r *renderer) refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, h := r.sc.Size()
	r.w = w
	r.h = h
}

func (r *renderer) reseed() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refresh()

	r.nameStyle = internal.Choose(fgPallete...)

	r.swarm = nil //TODO: does this really trigger the gc?

	r.swarm = make([]*layer, r.swarmSize)
	// TODO: clean out swarm
	for i := 0; i < r.swarmSize; i++ {
		r.swarm[i] = newRandomFish(r.w, r.h)
	}

	r.bubbles = nil //TODO: does this really trigger gc?
	r.bubbles = make([]*layer, r.swarmSize)
	// NOTE: the bubbles will be created and rendered as the fish moves
	// and the x,y of the fish is known..
}

func (r *renderer) start() {
	r.reseed()
	go r.render()
}

func (r *renderer) restart() {
	r.stop()
	r.start()
}

var nameRaw = `	
  ___  ___ __ _  ___
 / _ \/ -_)  ' \/ _ \
/_//_/\__/_/_/_/\___/ 1.0`

var nameTiles = strings.Split(nameRaw, "\n")

// TODO: i should have a more generic function that can render a bunch of
// bytes to a x,y,w,h
func (r *renderer) renderName() {
	nameX := r.w - len(nameTiles[len(nameTiles)-1]) - 1
	nameY := r.h - len(nameTiles) - 1
	for _, tile := range nameTiles {
		rx := nameX
		for _, ch := range tile {
			fmt.Println(rx, nameY)
			r.sc.SetContent(rx, nameY, ch, nil, r.nameStyle)
			rx++
		}
		nameY++
	}
}

func (r *renderer) renderBubbles() {
	if r.bubbles == nil {
		// TODO: what do do here.. should not happen
		return
	}
	for i, l := range r.bubbles {
		bx := 0
		if l.velo < 0 {
			bx = l.x + 1
		} else {
			bx = l.x + l.asset.Width
		}
		if bx > 0 && (bx%(r.w/4) == 0) {
			b := newRandomBubble(r.w, r.h)
			b.x = bx
			b.y = l.y - 1
			r.bubbles[i] = b
		}

		l.drawFunc(l, r.sc)
	}
}

func (r *renderer) renderSwarm() {
	for _, l := range r.swarm {
		internal.Logln("LAYER DRAW %v", l)
		l.drawFunc(l, r.sc)
	}
}

func (r *renderer) render() {
	r.running = true
	r.sc.Clear()
	for r.running {
		r.renderName()
		// r.renderBubbles()
		// r.renderSwarm()
		// r.sc.Show()
		time.Sleep(renderTickDelay)
	}
}

func newRenderer(sc tcell.Screen, swarmSize int) *renderer {
	r := renderer{sc: sc, swarmSize: swarmSize}
	r.reseed()
	return &r
}
