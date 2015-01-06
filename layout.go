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
type Layout struct {
	headerSize int
	footerSize int
	// main viewer's size will determined by calculating
}

func NewLayout() *Layout {
	defaultHeaderSize := 1
	defaultFooterSize := 1
	return &Layout{defaultHeaderSize, defaultFooterSize}
}

func (l *Layout) MainViewerBound() *Area {
	termw, termh := term.Size()

	min := Point{l.headerSize, 0}
	max := Point{termh-l.footerSize, termw}

	return NewArea(min, max)
}

// func (l *Layout) MainViewerSize() image.Point {
// 	return l.MainViewerBound().Size()
// }
