package main

type Point struct {
	l int // l
	o int // o
}

func (p Point) Add(q Point) Point {
	return Point{p.l + q.l, p.o + q.o}
}

func (p Point) Sub(q Point) Point {
	return Point{p.l - q.l, p.o - q.o}
}

type Area struct {
	min Point
	max Point
}

func NewArea(a, b Point) *Area {
	minl := a.l
	maxl := b.l
	if minl > maxl {
		minl, maxl = maxl, minl
	}
	mino := a.o
	maxo := b.o
	if mino > maxo {
		mino, maxo = maxo, mino
	}
	return &Area{min: Point{minl, mino}, max: Point{maxl, maxo}}
}

func (a *Area) Size() Point {
	return Point{a.max.l - a.min.l, a.max.o - a.min.o}
}
