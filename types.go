package gooey

type DrawCall struct {
	ElementIndex  int
	isHighlighted bool
	Instance      *uiElementInstance
	Color         Color
	Rect          Rect
	rectSet       bool
}

func (dc DrawCall) IsHighlighted() bool {
	return dc.isHighlighted
}

func (l *Layout) newDefaultDrawcall() DrawCall {
	dc := DrawCall{Color: NewColor(1, 1, 1, 1)}
	// Set up the default starting rectangle
	// dc.Rect = l.Rect
	// dc.ElementIndex = l.elementIndex
	// dc.isHighlighted = highlightedElement == dc.Instance
	// // Use the layout function to position / partition the starting rectangle
	// dc = l.ItemRect(dc)
	return dc
}

// UI elements have the following properties.
type UIElement interface {
	highlightable() bool    // Informs as to whether the UI element is highlightable (buttons) or not
	draw(drawCall DrawCall) // Draws using the information in the draw call struct, and then returns a state
}

type uiElementInstance struct {
	id          string
	currentRect Rect
	prevRect    Rect
	layout      *Layout
	drawable    UIElement
	state       any
	wasDrawn    bool
}

func (u *uiElementInstance) Clone() *uiElementInstance {
	newInst := *u
	return &newInst
}

func (u *uiElementInstance) State() any {
	return u.state
}

func (u *uiElementInstance) Layout() *Layout {
	return u.layout
}

func (u *uiElementInstance) Rect() Rect {
	return u.currentRect
}

func (u *uiElementInstance) PrevRect() Rect {
	return u.prevRect
}

// func (i *uiElementInstance) Clone() *uiElementInstance {
// 	newI := *i
// 	return &newI
// }
