package main

import(
	"image"
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

func (l *layout) mainViewerBound() *image.Rectangle {
	termw, termh := term.Size()

	bbmin := image.Point{0, l.headerSize}
	bbmax := image.Point{termw, termh-l.footerSize}

	return &image.Rectangle{bbmin, bbmax}
}

func (l *layout) mainViewerSize() image.Point {
	return l.mainViewerBound().Size()
}
