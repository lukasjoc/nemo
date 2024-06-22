package layer

import (
	"fmt"
	"unicode"

	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
	"github.com/lukasjoc/nemo/internal/assets"
)

type Layer struct {
	X          int
	Y          int
	Velo       int
	hidden     bool
	style      tcell.Style
	Asset      assets.Asset
	AssetIndex int
	// NOTE: that the drawFunc doesnt actual.Y update the screen
	// it just computes the next l.Yer. Its up to the renderer to sync
	// the changes to the screen. This effectively allows for double buffering.
	Draw func(l *Layer, sc tcell.Screen)
}

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
		l.X, l.Y, l.Velo, l.hidden, l.Asset.Group)
}

func fishDrawFunc(l *Layer, sc tcell.Screen) {
	drawW, _ := sc.Size()
	initialX := l.X
	initialY := l.Y
	ty := initialY
	for _, tile := range l.Asset.Sources[l.AssetIndex] {
		tlen := len(tile)
		if tlen == 0 {
			continue
		}
		tx := initialX
		for _, r := range tile {
			// clear any garbage from the previous draw
			if l.Velo > 0 {
				for i := initialX - (l.Velo) - 1; i < initialX; i++ {
					sc.SetContent(i, ty, ' ', nil, tcell.StyleDefault)
				}
			}
			if l.Velo < 0 {
				for i := (initialX + tlen); i < (initialX + tlen + -l.Velo); i++ {
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
	if l.Velo > 0 && l.X > drawW+l.Asset.Width ||
		l.Velo < 0 && l.X < -l.Asset.Width {
		(*l).hidden = true
	}
	(*l).X += l.Velo
}

func NewRandFish(w int, h int) *Layer {
	asset := assets.Random("fish")
	l := Layer{
		Velo:       internal.Choose(2, 1, 3),
		style:      internal.Choose(Colors...),
		Asset:      asset,
		AssetIndex: internal.Choose(0, 1),
	}
	leftSide := l.AssetIndex == 0
	if leftSide {
		l.X = -(internal.IntRand((asset.Width*8)-asset.Width) + asset.Width)
		l.Y = internal.IntRand(h - asset.Height)
	} else {
		l.X = (internal.IntRand((w+asset.Width*8)-(w+asset.Width)) + w + asset.Width)
		l.Y = internal.IntRand(h - asset.Height)
		l.Velo *= -1
	}
	l.Draw = fishDrawFunc
	return &l
}

func bubbleDrawFunc(l *Layer, sc tcell.Screen) {
	(*l).Asset = assets.Random("bubble")
	initialX := l.X
	initialY := l.Y
	ty := initialY
	for _, tile := range l.Asset.Sources[l.AssetIndex] {
		tlen := len(tile)
		if tlen == 0 {
			continue
		}
		tx := initialX
		for _, r := range tile {
			// TODO: dont rely on asset size implicitly
			sc.SetContent(tx, ty-l.Velo, ' ', nil, tcell.StyleDefault)
			if !unicode.IsSpace(r) {
				sc.SetContent(tx, ty, r, nil, l.style)
			} else {
				sc.SetContent(tx, ty, r, nil, tcell.StyleDefault)
			}
			tx++
		}
		ty++
	}
	if l.Y < -l.Asset.Height {
		(*l).hidden = true
	}
	(*l).Y += l.Velo
}

func NewRandBubble(w int, h int) *Layer {
	asset := assets.Random("bubble")
	l := Layer{
		Velo:       -internal.Choose(3, 2, 4, 5),
		style:      internal.Choose(Blues...),
		Asset:      asset,
		X:          internal.IntRand(w),
		Y:          internal.IntRand(h / 2),
		AssetIndex: 0,
	}
	l.Draw = bubbleDrawFunc
	return &l
}
