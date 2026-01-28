package gooey

import (
	"fmt"
	"image"
	"log"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// ArrangeFunc is a function type used to take a draw call (rectangle, color, etc.),
// and slice it up / reposition it as necessary.
type ArrangeFunc func(drawCall *DrawCall)

// An Arranger is an object that has a function that is used to determine how and where
// UI elements are rendered to the screen through a Layout.
type Arranger interface {
	Arrange(drawCall *DrawCall)
}

// Layout represents an object that is used as a target for UI elements to draw to.
// Each Layout has an Arranger that controls the behavior for positioning / spacing
// UI elements using the Layout's element index, which increments as you add UI elements
// to it.
type Layout struct {
	ID                 string
	Rect               Rect // Rect indicates where and how large the Layout is.
	HighlightingLocked bool

	AutoScrollSpeed        float32 // How smoothly to automatically scroll to the highlighted UI element for layouts that draw beyond the Layout's boundary Rect to the right and downwards.
	AutoScrollAcceleration float32 // The acceleration to the top speed (AutoScrollSpeed) for scrolling layouts when automatically scorolling.
	autoScrollCurrentSpeed Vector2

	// A custom highlighting order. If set to nil or an empty slice, then highlighting is done based on UI elements' positions.
	// Otherwise, the elements are highlighted in this given order. If an ID is given that doesn't exist, then it will
	// revert to automatic position-based highlighting when attempting to highlight that element.
	CustomHighlightingOrder []string

	committedMaxRect   Rect
	currentMaxRect     Rect
	elementIndex       int
	arranger           Arranger
	Offset             Vector2
	existingUIElements *sortedElementInstanceMap
}

// NewLayout creates a new Layout object for laying out elements in the given rectangle.
func NewLayout(id string, x, y, w, h float32) *Layout {

	for _, l := range existingLayouts {
		if l.ID == id {

			for _, v := range visibleLayouts {
				if v.ID == id {
					log.Println("gooey: cannot specify a new layout with the same ID multiple times")
				}
			}

			l.arranger = &ArrangerFull{}

			visibleLayouts = append(visibleLayouts, l)

			l.Reset()
			// l.uiDrawables = l.uiDrawables[:0]
			return l
		}
	}

	l := &Layout{
		ID:                     id,
		Rect:                   Rect{X: x, Y: y, W: w, H: h},
		existingUIElements:     newSortedElementInstanceMap(),
		arranger:               &ArrangerFull{},
		AutoScrollSpeed:        8,
		AutoScrollAcceleration: 0.5,
	}
	visibleLayouts = append(visibleLayouts, l)
	existingLayouts = append(existingLayouts, l)
	return l
}

// Creates a new Layout from a given Rectangle and gives it an ID string.
func NewLayoutFromRect(id string, rect Rect) *Layout {
	return NewLayout(id, rect.X, rect.Y, rect.W, rect.H)
}

var layoutsFromStrings = map[string]map[rune]*Layout{}

/*
	 NewLayoutsFromStrings creates a new series of layouts from a Rect and strings indicating the positioning and relative
	 proportions of each partition.
	 The runes / characters used in the `mappingStrings` strings indicate rectangles.
	 `baseRect` is the base rectangle indicating the position and overall size of the space to be partitioned.
	 For example, with a given `mappingStrings` set of:

	 aa b
	 aa b
		b
	 cccb

	 `NewLayoutsFromStrings` would create and return three Layouts - one for the Layout at the
	 top-left quarter (a), the next at the very right, extending from the top to the bottom of
	 the base Rect (b), and the last at the bottom-left, extending from 0 to 75% of the way across (c).
	 For an ID, they all have the same base ID (`idBase`) with their representative character
	 appended (e.g. "idBase_a", "idBase_b", etc).

	 Note that each Layout can only be initially created with this function once; after that, it just returns them.
	 Also note that Layouts created through this method are cached by the idBase string, so the idBase string should not change
	 for these specific Layouts.
*/
func NewLayoutsFromStrings(idBase string, baseRect Rect, mappingStrings ...string) map[rune]*Layout {

	chars := []rune{}

	if results, ok := layoutsFromStrings[idBase]; ok {
		for c := range results {
			chars = append(chars, c)
		}
		// Sort for consistency
		sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })
		for _, c := range chars {
			NewLayout(results[c].ID, 0, 0, 0, 0)
		}
		return results
	}

	resultingLayouts := map[rune]*Layout{}

	for _, row := range mappingStrings {
		for _, c := range row {

			found := false
			for _, c2 := range chars {
				if c == c2 {
					found = true
					break
				}
			}
			if !found {
				chars = append(chars, c)
			}

		}

	}

	// Sort for consistency
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })

	for _, c := range chars {

		if c == ' ' {
			continue
		}

		set := false
		rect := Rect{}

		for y, row := range mappingStrings {

			ty := baseRect.Y + (float32(y)/float32(len(mappingStrings)))*baseRect.H
			ty2 := baseRect.Y + (float32(y+1)/float32(len(mappingStrings)))*baseRect.H

			for x, char := range row {
				tx := baseRect.X + (float32(x)/float32(len(row)))*baseRect.W
				tx2 := baseRect.X + (float32(x+1)/float32(len(row)))*baseRect.W

				if c == char {

					if !set || tx < rect.X {
						rect.X = tx
					}

					if !set || ty < rect.Y {
						rect.Y = ty
					}

					if tx2 > rect.X+rect.W {
						rect.W = tx2 - rect.X
					}

					if ty2 > rect.Y+rect.H {
						rect.H = ty2 - rect.Y
					}

					set = true
				}

			}

		}

		if set {
			resultingLayouts[c] = NewLayout(idBase+"_"+string(c), rect.X, rect.Y, rect.W, rect.H)
			// resultingLayouts = append(resultingLayouts, NewLayout(idBase+"_"+string(c), rect.X, rect.Y, rect.W, rect.H))
		}

	}

	layoutsFromStrings[idBase] = resultingLayouts

	return resultingLayouts

}

func (l *Layout) Clone(newID string) *Layout {
	n := NewLayout(newID, l.Rect.X, l.Rect.Y, l.Rect.W, l.Rect.H)
	n.arranger = l.arranger
	n.AutoScrollAcceleration = l.AutoScrollAcceleration
	n.AutoScrollSpeed = l.AutoScrollSpeed
	n.CustomHighlightingOrder = l.CustomHighlightingOrder
	return n
}

func (l *Layout) String() string {
	return fmt.Sprintf("%v : { %d, %d, %d, %d }", l.ID, int(l.Rect.X), int(l.Rect.Y), int(l.Rect.W), int(l.Rect.H))
}

// Clone creates a clone of the given Layout with a new ID
// func (l *Layout) Clone(newID string) *Layout {
// 	newLayout := *l
// 	newLayout.ID = newID
// 	newLayout.existingUIElements = newLayout.existingUIElements.Clone()
// 	return &newLayout
// }

// AlignToScreenbuffer aligns an Area to the bounds of gooey's screenbuffer using an Alignment constant,
// with the desired padding in pixels.
func (l *Layout) AlignToScreenbuffer(alignment Alignment, padding float32) *Layout {
	l.Rect = l.Rect.AlignToScreenbuffer(alignment, padding)
	return l
}

// AlignToImage aligns an Area to the bounds of the image provided using an Alignment constant,
// with the desired padding in pixels.
func (l *Layout) AlignToImage(img *ebiten.Image, alignment Alignment, padding float32) *Layout {
	l.Rect = l.Rect.AlignToImage(img, alignment, padding)
	return l
}

// AlignToLayout aligns a Layout to the bounds of the other Layout provided using an Alignment constant,
// with the desired padding in pixels.
func (l *Layout) AlignToLayout(other *Layout, alignment Alignment, padding float32) *Layout {

	minX := other.Rect.X
	minY := other.Rect.Y
	maxX := other.Rect.X + other.Rect.W
	maxY := other.Rect.Y + other.Rect.H

	switch alignment {
	case AlignmentTopLeft:
		l.Rect.X = float32(minX) + padding
		l.Rect.Y = float32(minY) + padding
	case AlignmentTopCenter:
		l.Rect.X = float32(minX) + float32(other.Rect.W)/2 - l.Rect.W/2
		l.Rect.Y = float32(minY) + padding
	case AlignmentTopRight:
		l.Rect.X = float32(maxX) - l.Rect.W - padding
		l.Rect.Y = float32(minY) - padding
	case AlignmentCenterLeft:
		l.Rect.X = float32(minX) + padding
		l.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - l.Rect.H/2
	case AlignmentCenterCenter:
		l.Rect.X = float32(minX) + float32(other.Rect.W)/2 - l.Rect.W/2
		l.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - l.Rect.H/2
	case AlignmentCenterRight:
		l.Rect.X = float32(maxX) - l.Rect.W - padding
		l.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - l.Rect.H/2
	case AlignmentBottomLeft:
		l.Rect.X = float32(minX) + padding
		l.Rect.Y = float32(maxY) - l.Rect.H - padding
	case AlignmentBottomCenter:
		l.Rect.X = float32(minX) + float32(other.Rect.W)/2 - l.Rect.W/2
		l.Rect.Y = float32(maxY) - l.Rect.H - padding
	case AlignmentBottomRight:
		l.Rect.X = float32(maxX) - l.Rect.W - padding
		l.Rect.Y = float32(maxY) - l.Rect.H - padding
	}

	return l
}

func (l *Layout) subscreen() *ebiten.Image {
	// return screenBuffer
	return screenBuffer.SubImage(image.Rect(int(l.Rect.X), int(l.Rect.Y), int(l.Rect.X)+int(l.Rect.W), int(l.Rect.Y)+int(l.Rect.H))).(*ebiten.Image)
}

// Reset resets the Layout so that any additionally drawn UI elements' positions
// go back to the start.
func (l *Layout) Reset() {
	l.elementIndex = 0
	l.committedMaxRect = Rect{}
	l.currentMaxRect = Rect{}
}

// Advance advances the Layout by the element number given. You shouldn't need to call this, but can
// be useful if you want to skip spaces.
func (l *Layout) Advance(delta int) {
	l.elementIndex += delta
}

// itemRect returns the current item's rectangle when transformed by the Layout's layout function.
func (l *Layout) itemRect(drawCall *DrawCall) {
	l.arranger.Arrange(drawCall)
	drawCall.Rect = drawCall.Rect.MoveVec(l.Offset)
	// If the spacing rectangle isn't set, then set it to be whatever Rect is set to.
	if drawCall.SpacingRect.IsZero() {
		drawCall.SpacingRect = drawCall.Rect
	}
	drawCall.SpacingRect = drawCall.SpacingRect.MoveVec(l.Offset)
}

// Arranger returns the layout function for the Layout.
func (l *Layout) Arranger() Arranger {
	return l.arranger
}

// SetArranger sets the layout function for laying out elements in the Layout's rectangle.
// Setting the arranger resets the maximum rectangle of drawn elements (so the Layout won't scroll).
func (l *Layout) SetArranger(arranger Arranger) *Layout {
	l.arranger = arranger
	l.committedMaxRect = Rect{}
	l.currentMaxRect = Rect{}
	l.elementIndex = 0
	return l
}

type customArranger struct {
	ArrangeFunc ArrangeFunc
}

func (c customArranger) Arrange(drawcall *DrawCall) {
	c.ArrangeFunc(drawcall)
}

// Sets a custom layouting function for laying out elements in the Layout's rectangle.
func (l *Layout) SetCustomArranger(layoutFunc ArrangeFunc) *Layout {
	return l.SetArranger(customArranger{
		ArrangeFunc: layoutFunc,
	})
}

func (l *Layout) add(id string, drawable UIElement, drawCall *DrawCall) {

	inst := l.existingUIElements.Add(id)

	inst.layout = l

	drawCall.ElementIndex = l.elementIndex
	drawCall.Instance = inst

	drawCall.isHighlighted = inst == highlightedElement

	inst.drawable = drawable
	inst.wasDrawn = true

	inst.prevRect = inst.currentRect

	// Use the layout function to position / partition the starting rectangle
	if !drawCall.rectSet {
		drawCall.Rect = l.Rect
		l.itemRect(drawCall)
		drawCall.rectSet = true
	}

	inst.drawable.draw(drawCall) // the state is set here
	inst.currentRect = drawCall.Rect

	if drawCall.InfluenceScrolling {

		emptyRect := l.currentMaxRect.IsZero()

		if emptyRect || drawCall.SpacingRect.X < l.currentMaxRect.X {
			l.currentMaxRect = l.currentMaxRect.ScaleLeftTo(drawCall.SpacingRect.X)
		}

		if emptyRect || drawCall.SpacingRect.Right() > l.currentMaxRect.Right() {
			l.currentMaxRect = l.currentMaxRect.ScaleRightTo(drawCall.SpacingRect.Right())
		}

		if emptyRect || drawCall.SpacingRect.Y < l.currentMaxRect.Y {
			l.currentMaxRect = l.currentMaxRect.ScaleUpTo(drawCall.SpacingRect.Y)
		}

		if emptyRect || drawCall.SpacingRect.Bottom() > l.currentMaxRect.Bottom() {
			l.currentMaxRect = l.currentMaxRect.ScaleDownTo(drawCall.SpacingRect.Bottom())
		}

	}

	l.Advance(1)

}

func (l *Layout) isVisible() bool {
	for _, layout := range visibleLayouts {
		if layout == l {
			return true
		}
	}
	return false
}

// Returns the UI element instance of the given ID string.
// If no such ID is found, the function returns nil.
func (l *Layout) UIElement(id string) *uiElementInstance {
	if element, ok := l.existingUIElements.Data[id]; ok {
		return element
	}
	return nil
}

// ArrangerFull simply is used to place any UI elements over a Layout's entire Rect.
// This is the default arrangement method for a new Layout.
type ArrangerFull struct {
	PaddingLeft   float32
	PaddingRight  float32
	PaddingTop    float32
	PaddingBottom float32
}

func (a ArrangerFull) WithPaddingLeft(padding float32) ArrangerFull {
	a.PaddingLeft = padding
	return a
}

func (a ArrangerFull) WithPaddingRight(padding float32) ArrangerFull {
	a.PaddingRight = padding
	return a
}

func (a ArrangerFull) WithPaddingTop(padding float32) ArrangerFull {
	a.PaddingTop = padding
	return a
}

func (a ArrangerFull) WithPaddingBottom(padding float32) ArrangerFull {
	a.PaddingBottom = padding
	return a
}

func (a ArrangerFull) WithPaddingHorizontal(padding float32) ArrangerFull {
	a.PaddingLeft = padding
	a.PaddingRight = padding
	return a
}

func (a ArrangerFull) WithPaddingVertical(padding float32) ArrangerFull {
	a.PaddingTop = padding
	a.PaddingBottom = padding
	return a
}

func (a ArrangerFull) WithPadding(padding float32) ArrangerFull {
	a.PaddingLeft = padding
	a.PaddingRight = padding
	a.PaddingTop = padding
	a.PaddingBottom = padding
	return a
}

func (a ArrangerFull) Arrange(drawCall *DrawCall) {

	drawCall.Rect.W -= a.PaddingRight + a.PaddingLeft
	drawCall.Rect.H -= a.PaddingBottom + a.PaddingTop
	drawCall.Rect.X += a.PaddingLeft
	drawCall.Rect.Y += a.PaddingTop

}

type ArrangerGridDirection int

const (
	ArrangeDirectionRow    ArrangerGridDirection = iota // ArrangerGrid places elements horizontally first
	ArrangeDirectionColumn                              // ArrangerGrid places elements vertically first
)

const (
	ContainerSize float32 = -9999999999999
)

// ArrangerGrid arranges UI elements in an easily extendible grid.
// Can be used for single columns or rows by specifying the DivisionSize
// to be 1 or below.
type ArrangerGrid struct {
	OuterPaddingLeft   float32 // How many pixels should be given as padding between elements and the left border
	OuterPaddingRight  float32 // How many pixels should be given as padding between elements and the right border
	OuterPaddingTop    float32 // How many pixels should be given as padding between elements and the top border
	OuterPaddingBottom float32 // How many pixels should be given as padding between elements and the bottom border

	ElementPadding Vector2 // How many pixels should be given as padding between elements

	// The size of the Elements in the Grid in pixels.
	// You can set any element to be the const `gooey.ContainerSize` or a percentage of it to easily
	// set element sizes to be fractions of the container size.
	// A size of 0 is the same as `gooey.ContainerSize`.
	ElementSize      Vector2
	ElementCount     int  // Number of elements per division (column or row).
	NoCenterElements bool // Center elements in the division if their combined size is less than the rectangle being arranged

	/*
	   Whether elements increase across (RowMajor) or vertically (ColumnMajor).

	   As an example, drawing eight (8) elements with an ElementNumber of 3:

	   ArrangeDirectionRow:

	   [ 0 ] [ 1 ] [ 2 ]

	   [ 3 ] [ 4 ] [ 5 ]

	   [ 6 ] [ 7 ]

	   ArrangeDirectionColumn:

	   [ 0 ] [ 3 ] [ 6 ]

	   [ 1 ] [ 4 ] [ 7 ]

	   [ 2 ] [ 5 ]
	*/
	Direction ArrangerGridDirection
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithElementPaddingWidth(paddingW float32) ArrangerGrid {
	a.ElementPadding.X = paddingW
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithElementPaddingHeight(paddingH float32) ArrangerGrid {
	a.ElementPadding.Y = paddingH
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithElementPadding(padding float32) ArrangerGrid {
	a.ElementPadding.X = padding
	a.ElementPadding.Y = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithElementPaddingVec(padding Vector2) ArrangerGrid {
	a.ElementPadding = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithOuterPaddingLeft(padding float32) ArrangerGrid {
	a.OuterPaddingLeft = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithOuterPaddingRight(padding float32) ArrangerGrid {
	a.OuterPaddingRight = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithOuterPaddingTop(padding float32) ArrangerGrid {
	a.OuterPaddingTop = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithOuterPaddingBottom(padding float32) ArrangerGrid {
	a.OuterPaddingBottom = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithOuterPaddingWidth(padding float32) ArrangerGrid {
	a.OuterPaddingLeft = padding
	a.OuterPaddingRight = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithOuterPaddingHeight(padding float32) ArrangerGrid {
	a.OuterPaddingTop = padding
	a.OuterPaddingBottom = padding
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithOuterPadding(padding float32) ArrangerGrid {
	a.OuterPaddingTop = padding
	a.OuterPaddingBottom = padding
	a.OuterPaddingLeft = padding
	a.OuterPaddingRight = padding
	return a
}

// Returns the ArrangerGrid with the size of the elements set to the given values in pixels.
// You can set either part of the size to be the const `gooey.ContainerSize` or a percentage of it to easily
// set element sizes to be fractions of the container size.
// A size of 0 is the same as `gooey.ContainerSize`.
func (a ArrangerGrid) WithElementSize(w, h float32) ArrangerGrid {
	a.ElementSize.X = w
	a.ElementSize.Y = h
	return a
}

// Returns the ArrangerGrid with the size of the elements set to the given values in pixels.
// You can set either part of the size to be the const `gooey.ContainerSize` or a percentage of it to easily
// set element sizes to be fractions of the container size.
// A size of 0 is the same as `gooey.ContainerSize`.
func (a ArrangerGrid) WithElementSizeVec(size Vector2) ArrangerGrid {
	a.ElementSize = size
	return a
}

// Returns the ArrangerGrid with the size of the elements set to the given values in pixels.
// You can set either part of the size to be the const `gooey.ContainerSize` or a percentage of it to easily
// set element sizes to be fractions of the container size.
// A size of 0 is the same as `gooey.ContainerSize`.
func (a ArrangerGrid) WithElementSizeW(elementWidth float32) ArrangerGrid {
	a.ElementSize.X = elementWidth
	return a
}

// Returns the ArrangerGrid with the size of the elements set to the given values in pixels.
// You can set either part of the size to be the const `gooey.ContainerSize` or a percentage of it to easily
// set element sizes to be fractions of the container size.
// A size of 0 is the same as `gooey.ContainerSize`.
func (a ArrangerGrid) WithElementSizeH(elementHeight float32) ArrangerGrid {
	a.ElementSize.Y = elementHeight
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithElementCount(count int) ArrangerGrid {
	a.ElementCount = count
	return a
}

// Returns the ArrangerGrid with the given property set.
func (a ArrangerGrid) WithNoCenterElements(noCenterElements bool) ArrangerGrid {
	a.NoCenterElements = noCenterElements
	return a
}

// Arranges UI elements given a DrawCall.
func (a ArrangerGrid) Arrange(drawCall *DrawCall) {

	drawCall.Rect.W -= a.OuterPaddingLeft + a.OuterPaddingRight
	drawCall.Rect.H -= a.OuterPaddingTop + a.OuterPaddingBottom

	cellWidth := float32(0)
	cellHeight := float32(0)

	if a.ElementCount <= 1 {
		a.ElementCount = 1
	}

	if a.ElementPadding.X < 0 {
		a.ElementPadding.X = 0
	}

	if a.ElementPadding.Y < 0 {
		a.ElementPadding.Y = 0
	}

	// If the size of each element is set, then go with that.
	// Otherwise, figure it out from the rectangle's size
	if a.ElementSize.X > 0 {
		cellWidth = a.ElementSize.X
	} else {
		if a.Direction == ArrangeDirectionRow {
			cellWidth = drawCall.Rect.W / float32(a.ElementCount)
		} else {
			cellWidth = drawCall.Rect.W
		}
		containerMulti := float32(1)
		if a.ElementSize.X < 0 {
			containerMulti = a.ElementSize.X / ContainerSize
		}
		a.ElementSize.X = cellWidth * containerMulti
	}

	if a.ElementSize.Y > 0 {
		cellHeight = a.ElementSize.Y
	} else {
		if a.Direction == ArrangeDirectionColumn {
			cellHeight = drawCall.Rect.H / float32(a.ElementCount)
		} else {
			cellHeight = drawCall.Rect.H
		}
		containerMulti := float32(1)
		if a.ElementSize.Y < 0 {
			containerMulti = a.ElementSize.Y / ContainerSize
		}
		a.ElementSize.Y = cellHeight * containerMulti

	}

	lx := float32(0)
	ly := float32(0)

	esx := a.ElementSize.X + a.ElementPadding.X
	esy := a.ElementSize.Y + a.ElementPadding.Y

	if a.Direction == ArrangeDirectionRow {
		lx = float32(drawCall.ElementIndex%a.ElementCount) * esx
		ly = float32(drawCall.ElementIndex/a.ElementCount) * esy
	} else {
		lx = float32(drawCall.ElementIndex/a.ElementCount) * esx
		ly = float32(drawCall.ElementIndex%a.ElementCount) * esy
	}

	drawCall.SpacingRect = Rect{
		X: drawCall.Rect.X + lx,
		Y: drawCall.Rect.Y + ly,
		W: a.ElementSize.X + a.OuterPaddingRight,
		H: a.ElementSize.Y + a.OuterPaddingBottom,
	}.
		ScaleLeftTo(drawCall.Rect.X - a.OuterPaddingLeft).
		ScaleUpTo(drawCall.Rect.Y - a.OuterPaddingTop)

	// Center elements if their combined width or height is less than the rectangle size
	// and they're being arranged horizontally or vertically
	if !a.NoCenterElements {

		if a.Direction == ArrangeDirectionRow {
			totalWidth := float32(a.ElementCount)*esx - a.ElementPadding.X // Disregard padding on the right side
			if totalWidth < drawCall.Rect.W {
				lx += (drawCall.Rect.W - totalWidth) / 2
			}
		} else {
			totalHeight := float32(a.ElementCount)*esy - a.ElementPadding.Y // Disregard padding on the bottomright side
			if totalHeight < drawCall.Rect.H {
				ly += (drawCall.Rect.H - totalHeight) / 2
			}
		}

	}

	elementRect := Rect{
		X: drawCall.Rect.X + lx + a.OuterPaddingLeft,
		Y: drawCall.Rect.Y + ly + a.OuterPaddingTop,
		W: a.ElementSize.X,
		H: a.ElementSize.Y,
	}

	drawCall.Rect = elementRect

}
