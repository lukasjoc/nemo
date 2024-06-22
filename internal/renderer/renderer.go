package renderer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
	"github.com/lukasjoc/nemo/internal/layer"
)

const DefaultTickDelay = time.Millisecond * 120

type Renderer struct {
	mu        sync.RWMutex
	done      chan bool
	t         *time.Ticker
	w         int
	h         int
	nameStyle tcell.Style
	swarm     []*layer.Layer
	bubbles   []*layer.Layer
	// A initialized tcell screen instance.
	Screen tcell.Screen
	// The amount of random fish to generate. This is static for now, but
	// planned to be dynamic in the future.
	SwarmSize int
	// A delay to reduce the render speed with.
	// As defined in `render.DefaultTickDelay` the default delay is 120ms.
	TickDelay time.Duration
	// Signals if the renderer has been stopped recently. This can be used
	// as a hook to stop and start the renderer.
	Stopped chan bool
}

func (r *Renderer) Stop() {
	select {
	case <-r.Stopped:
		go func() { r.Stopped <- true }()
	default:
		r.mu.Lock()
		r.t.Stop()
		r.mu.Unlock()
		go func() { r.done <- true }()
		go func() { r.Stopped <- true }()
		if internal.DebugEnabled {
			r.renderStats(time.Now())
			r.mu.Lock()
			r.Screen.Show()
			r.mu.Unlock()
		}
	}

}

func (r *Renderer) Start() {
	r.mu.Lock()
	r.t.Reset(r.TickDelay)
	r.mu.Unlock()
	go r.render()
}

func (r *Renderer) Restart() {
	select {
	case <-r.Stopped:
	default:
		r.Stop()
		<-r.Stopped
	}
	r.refresh()
	r.Reset()
	r.Start()
}

func (r *Renderer) Destroy() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.swarm = nil
	r.bubbles = nil
}

func (r *Renderer) refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, h := r.Screen.Size()
	r.w = w
	r.h = h
	r.Screen.Clear()
}

func (r *Renderer) Reset() {
	r.Destroy()
	r.refresh()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nameStyle = internal.Choose(layer.Colors...)
	r.swarm = make([]*layer.Layer, r.SwarmSize)
	for i := 0; i < r.SwarmSize; i++ {
		r.swarm[i] = layer.NewRandFish(r.w, r.h)
	}
	r.bubbles = make([]*layer.Layer, r.SwarmSize)
	// NOTE: the bubbles will be created and rendered as the fish moves
	// and the x,y of the fish is known..
}

var nameRaw = `	
  ___  ___ __ _  ___
 / _ \/ -_)  ' \/ _ \
/_//_/\__/_/_/_/\___/ 1.1`

var nameTiles = strings.Split(nameRaw, "\n")

// TODO: i should have a more generic function that can render a bunch of
// bytes to a x,y,w,h
// func (r *Renderer) renderText(x int, y int, w int, h int, text string) { }

func (r *Renderer) renderName() {
	r.mu.Lock()
	defer r.mu.Unlock()
	nameX := r.w - len(nameTiles[len(nameTiles)-1]) - 4
	nameY := r.h - len(nameTiles) - 1
	for _, tile := range nameTiles {
		rx := nameX
		for _, ch := range tile {
			r.Screen.SetContent(rx, nameY, ch, nil, r.nameStyle)
			rx++
		}
		nameY++
	}
}

func (r *Renderer) renderStats(ts time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fishCount := 0
	bubbleCount := 0
	for _, l := range r.swarm {
		if l == nil {
			continue
		}
		fishCount++
	}
	for _, l := range r.bubbles {
		if l == nil {
			continue
		}
		bubbleCount++
	}
	stats := fmt.Sprintf("TS: %5d\nFish: %5d\nBubbles: %5d", ts.Unix(), fishCount, bubbleCount)
	statsTiles := strings.Split(stats, "\n")
	nameX := r.w - len(statsTiles[len(statsTiles)-1]) - (1)
	nameY := 0 + len(statsTiles) - 1
	for _, tile := range statsTiles {
		rx := nameX
		for _, ch := range tile {
			r.Screen.SetContent(rx, nameY, ch, nil, tcell.StyleDefault)
			rx++
		}
		nameY++
	}
}

func (r *Renderer) renderBubbles() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.bubbles == nil {
		return
	}
	for _, layerIndex := range layer.FindHidden(r.bubbles) {
		r.bubbles[layerIndex] = nil
	}
	for i, l := range r.swarm {
		if l == nil {
			continue
		}
		bx := 0
		if l.Velo() < 0 {
			bx = l.X() + 1
		} else {
			bx = l.X() + l.Asset().Width
		}
		if bx > 0 && (bx%(r.w/4) == 0) {
			// TODO: move the layer stuff into here
			b := layer.NewRandBubble(r.w, r.h)
			b.SetX(bx)
			b.SetY(l.Y() - 1)
			r.bubbles[i] = b
		}
	}
	for _, l := range r.bubbles {
		if l == nil {
			continue
		}
		internal.Logln("LAYER DRAW %v", l)
		l.Draw(l, r.Screen)
	}
}

func (r *Renderer) renderSwarm() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.swarm == nil {
		return
	}
	for _, layerIndex := range layer.FindHidden(r.swarm) {
		r.swarm[layerIndex] = layer.NewRandFish(r.w, r.h)
	}
	for _, l := range r.swarm {
		if l == nil {
			continue
		}
		internal.Logln("LAYER DRAW %v", l)
		l.Draw(l, r.Screen)
	}
}

// TODO: i should have a draw loop and a update loop with different
// tick delays. I think that would make it even smoother.
func (r *Renderer) render() {
	for {
		select {
		case <-r.done:
			return
		case ts := <-r.t.C:
			r.renderName()
			r.renderSwarm()
			r.renderBubbles()
			if internal.DebugEnabled {
				r.renderStats(ts)
			}
			r.Screen.Show()
		}
	}
}

func New(sc tcell.Screen, swarmSize int, tickDelay time.Duration) *Renderer {
	r := Renderer{
		Screen:    sc,
		SwarmSize: swarmSize,
		TickDelay: tickDelay,
		mu:        sync.RWMutex{},
		t:         time.NewTicker(tickDelay),
		done:      make(chan bool),
		Stopped:   make(chan bool),
		swarm:     nil,
		bubbles:   nil,
	}
	return &r
}
