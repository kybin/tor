package main

import(
	term "github.com/nsf/termbox-go"
)

// Layout means divided section in a terminal window.
// For now, it divied by three section. header, main and footer.
//
// header (for tab)
// ---------------------------------
// main (for edit)
//
//
//
//
//
// ---------------------------------
// footer (for information)
// 
type layout struct {
	headerSize int
	footerSize int
	// main viewer's size will determined by calculating
}

func newLayout() *layout {
	defaultHeaderSize := 1
	defaultFooterSize := 1
	return &layout{defaultHeaderSize, defaultFooterSize}
}

func (l *layout) mainViewerBound() *Area {
	termw, termh := term.Size()

	min := Point{l.headerSize, 0}
	max := Point{termh-l.footerSize, termw}

	return NewArea(min, max)
}

// func (l *layout) mainViewerSize() image.Point {
// 	return l.mainViewerBound().Size()
// }
