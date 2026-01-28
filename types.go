package gooey

type DrawCall struct {
	ElementIndex int                // The index of the Element being drawn in the Layout.
	Instance     *uiElementInstance // The individual instance of the UI element being drawn.
	Color        Color              // The color to draw things with; inherited from previous UIElements.
	Rect         Rect               // The rectangle to draw things in.
	/*
		The base rect that the layout displays; this influences scrolling for Layouts (e.g.
		ArrangerGrid provides Rects that are sliced up including padding; we don't want to
		include padding when it comes to scrolling).
		If this is a zero Rect, then it's set to the DrawCall's Rect.
	*/
	SpacingRect        Rect
	InfluenceScrolling bool // If the elements being drawn should influence scrolling or not.

	isHighlighted bool
	rectSet       bool
}

func (dc *DrawCall) Clone() *DrawCall {
	newDC := *dc
	return &newDC
}

// Returns if the element being drawn is highlighted.
func (dc *DrawCall) IsHighlighted() bool {
	return dc.isHighlighted
}

func (l *Layout) newDefaultDrawcall() *DrawCall {
	dc := &DrawCall{Color: NewColor(1, 1, 1, 1), InfluenceScrolling: true}
	// Set up the default starting rectangle
	// dc.Rect = l.Rect
	// dc.ElementIndex = l.elementIndex
	// dc.isHighlighted = highlightedElement == dc.Instance
	// // Use the layout function to position / partition the starting rectangle
	// dc = l.ItemRect(dc)
	return dc
}

// UIElement is an interface of properties UI elements must have.
type UIElement interface {
	highlightable() bool     // Informs as to whether the UI element is highlightable (buttons) or not
	draw(drawCall *DrawCall) // Draws using the information in the draw call struct, and then returns a state
}

type uiElementInstance struct {
	id          string
	currentRect Rect
	prevRect    Rect
	layout      *Layout
	drawable    UIElement
	state       any
	wasDrawn    bool
	data        any
}

func (u *uiElementInstance) Clone() *uiElementInstance {
	newInst := *u
	return &newInst
}

func (u *uiElementInstance) ID() string {
	return u.id
}

func (u *uiElementInstance) UIElement() UIElement {
	return u.drawable
}

// State returns the persistent state of the UI Element.
// Some elements have no states (e.g. UIColor), while others do (e.g. UISlider).
func (u *uiElementInstance) State() any {
	return u.state
}

// Returns the Layout drawing the UI element.
func (u *uiElementInstance) Layout() *Layout {
	return u.layout
}

// Rect returns the last-drawn rectangle for the UI element.
func (u *uiElementInstance) Rect() Rect {
	return u.currentRect
}

// Returns the last-drawn rectangle for the UI element on the previous game frame.
func (u *uiElementInstance) PrevRect() Rect {
	return u.prevRect
}

func (u *uiElementInstance) Data() any {
	return u.data
}

func (u *uiElementInstance) SetData(data any) {
	u.data = data
}

// func (i *uiElementInstance) Clone() *uiElementInstance {
// 	newI := *i
// 	return &newI
// }
