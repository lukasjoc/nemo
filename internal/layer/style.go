package layer

import (
	"github.com/gdamore/tcell"
	"github.com/lukasjoc/nemo/internal"
)

var Blues = []tcell.Style{
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCyan),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightBlue),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSkyBlue),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSteelBlue),
}

var Colors = append([]tcell.Style{
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorOrchid),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGoldenrod),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleGreen),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleTurquoise),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPaleVioletRed),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPapayaWhip),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorPeachPuff),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightCoral),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGoldenrodYellow),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGray),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightGreen),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightPink),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSalmon),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSlateGray),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightYellow),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLightSeaGreen),
	tcell.StyleDefault.Dim(true).Bold(true).Foreground(tcell.ColorLimeGreen),
}, Blues...)

func bodypartColorMask(ch rune) tcell.Style {
	style := tcell.StyleDefault.Dim(true).Bold(true)
	switch ch {
	case '\\', '/', '#', '~', '-', '_', '<', '(', ')':
		return internal.Choose(
			style.Foreground(tcell.ColorLightYellow),
			style.Foreground(tcell.ColorLightGreen))
	case 'C', '@', 'o':
		return style.Foreground(tcell.ColorPaleVioletRed)
	case ',', '"', '\'', ';', ':', '=':
		return style.Foreground(tcell.ColorLightCoral)
	}
	return style
}
