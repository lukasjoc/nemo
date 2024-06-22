package layer

import (
	"fmt"
	"unicode"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
	"github.com/lukasjoc/nemo/internal/assets"
)

type Layer struct {
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
	Draw func(l *Layer, sc tcell.Screen)
}

func (l *Layer) X() int              { return l.x }
func (l *Layer) Y() int              { return l.y }
func (l *Layer) Velo() int           { return l.velo }
func (l *Layer) Asset() assets.Asset { return l.asset }
func (l *Layer) SetX(newX int)       { l.x = newX }
func (l *Layer) SetY(newY int)       { l.x = newY }

func FindHidden(layers []*Layer) []int {
	idx := []int{}
	for i, l := range layers {
		if l != nil && l.hidden {
			idx = append(idx, i)
		}
	}
	return idx
}

func (l Layer) String() string {
	return fmt.Sprintf("x:%4d y:%4d velo:%4d hidden:%6t group:%6s",
		l.x, l.y, l.velo, l.hidden, l.asset.Group)
}

func (l *Layer) setDraw(f func(l *Layer, sc tcell.Screen)) {
	l.Draw = f
}

func fishDrawFunc(l *Layer, sc tcell.Screen) {
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
				sc.SetContent(tx, ty, r, nil, bodypartColorMask(r))
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

func NewRandFish(w int, h int) *Layer {
	asset := assets.Random("fish")
	l := Layer{
		velo:       internal.Choose(2, 1, 3),
		style:      internal.Choose(Colors...),
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
	l.setDraw(fishDrawFunc)
	return &l
}

func bubbleDrawFunc(l *Layer, sc tcell.Screen) {
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
	if l.y < -l.asset.Height {
		(*l).hidden = true
	}
	(*l).y += l.velo
}

func NewRandBubble(w int, h int) *Layer {
	asset := assets.Random("bubble")
	l := Layer{
		velo:       -internal.Choose(3, 2, 4, 5),
		style:      internal.Choose(Blues...),
		asset:      asset,
		x:          internal.IntRand(w),
		y:          internal.IntRand(h / 2),
		assetIndex: 0,
	}
	l.setDraw(bubbleDrawFunc)
	return &l
}
