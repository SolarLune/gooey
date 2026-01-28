package gooey

// Page represents an object that controls highlighting unidirectional highlighting flow for multiple Layouts.
// If you have a menu system that goes from selecting an option in a menu in layout A, to another menu in layout B, and finally an option in layout C,
// then Page would help you advance through those Layouts by controlling if their highlighting is locked.
type Page struct {
	Layouts     []*Layout
	activeIndex int
}

func NewPage(layouts ...*Layout) Page {
	p := Page{
		Layouts: layouts,
	}

	if len(layouts) == 0 {
		return p
	}

	set := false
	active := -1
	for i, l := range layouts {
		if l.HighlightingLocked {
			set = true
		} else {
			active = i
		}
	}

	if !set {
		p.MakeActive(layouts[0])
	} else {
		p.activeIndex = active
	}

	return p
}

func (p *Page) MakeActive(l *Layout) {

	for i, other := range p.Layouts {
		if other == l {
			l.HighlightingLocked = false
		} else {
			other.HighlightingLocked = true
		}
		p.activeIndex = i
	}

}

func (p *Page) Advance(advance int) {

	// Advance forward
	if advance != 0 {
		next := clamp(p.activeIndex+advance, 0, len(p.Layouts)-1)
		if p.activeIndex != next {
			p.activeIndex = next
			p.MakeActive(p.Layouts[next])
			highlightedElement = nil
		}
	}

}
