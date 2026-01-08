package gooey

import (
	"errors"
	"fmt"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// ArrangeFunc is a function type used to take a draw call (rectangle, color, etc.),
// and slice it up / reposition it as necessary.
type ArrangeFunc func(drawCall DrawCall) DrawCall

// Arranger is an object that has a function that is used to determine how and where
// UI elements are rendered to the screen through a Layout.
type Arranger interface {
	Arrange(drawCall DrawCall) DrawCall
}

// Layout represents an object that is used as a target for UI elements to draw to.
// Each Layout has an Arranger that controls the behavior for positioning / spacing
// UI elements using the Layout's element index, which increments as you add UI elements
// to it.
type Layout struct {
	ID   string
	Rect Rect

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

			visibleLayouts = append(visibleLayouts, l)
			l.Reset()
			for _, d := range l.existingUIElements.Data {
				d.wasDrawn = false
			}
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
*/
func NewLayoutsFromStrings(idBase string, baseRect Rect, mappingStrings ...string) map[rune]*Layout {

	// resultingLayouts = resultingLayouts[:0]

	resultingLayouts := map[rune]*Layout{}

	chars := []rune{}

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

	// sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })

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

	return resultingLayouts

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

// AlignToScreenbuffer aligns an Area to the bounds of gooey's screenbuffer using an AnchorPosition constant,
// with the desired padding in pixels.
func (l *Layout) AlignToScreenbuffer(anchor AnchorPosition, padding float32) {
	l.Rect = l.Rect.AlignToScreenbuffer(anchor, padding)
}

// AlignToImage aligns an Area to the bounds of the image provided using an AnchorPosition constant,
// with the desired padding in pixels.
// Using any anchor positions that aren't supported will return an error.
func (l *Layout) AlignToImage(img *ebiten.Image, anchor AnchorPosition, padding float32) {
	l.Rect = l.Rect.AlignToImage(img, anchor, padding)
}

// AlignToLayout aligns a Layout to the bounds of the other Layout provided using an AnchorPosition constant,
// with the desired padding in pixels.
// Using any anchor positions that aren't supported will return an error.
func (l *Layout) AlignToLayout(other *Layout, anchor AnchorPosition, padding float32) error {

	minX := other.Rect.X
	minY := other.Rect.Y
	maxX := other.Rect.X + other.Rect.W
	maxY := other.Rect.Y + other.Rect.H

	switch anchor {
	case AnchorTopLeft:
		l.Rect.X = float32(minX) + padding
		l.Rect.Y = float32(minY) + padding
	case AnchorTopCenter:
		l.Rect.X = float32(minX) + float32(other.Rect.W)/2 - l.Rect.W/2
		l.Rect.Y = float32(minY) + padding
	case AnchorTopRight:
		l.Rect.X = float32(maxX) - l.Rect.W - padding
		l.Rect.Y = float32(minY) - padding
	case AnchorCenterLeft:
		l.Rect.X = float32(minX) + padding
		l.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - l.Rect.H/2
	case AnchorCenter:
		l.Rect.X = float32(minX) + float32(other.Rect.W)/2 - l.Rect.W/2
		l.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - l.Rect.H/2
	case AnchorCenterRight:
		l.Rect.X = float32(maxX) - l.Rect.W - padding
		l.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - l.Rect.H/2
	case AnchorBottomLeft:
		l.Rect.X = float32(minX) + padding
		l.Rect.Y = float32(maxY) - l.Rect.H - padding
	case AnchorBottomCenter:
		l.Rect.X = float32(minX) + float32(other.Rect.W)/2 - l.Rect.W/2
		l.Rect.Y = float32(maxY) - l.Rect.H - padding
	case AnchorBottomRight:
		l.Rect.X = float32(maxX) - l.Rect.W - padding
		l.Rect.Y = float32(maxY) - l.Rect.H - padding
	default:
		return errors.New("can't align area to an image using an unsupported alignment type")
	}

	return nil
}

func (l *Layout) subscreen() *ebiten.Image {
	// return screenBuffer
	return screenBuffer.SubImage(image.Rect(int(l.Rect.X), int(l.Rect.Y), int(l.Rect.X)+int(l.Rect.W), int(l.Rect.Y)+int(l.Rect.H))).(*ebiten.Image)
}

// Reset resets the Layout so that any additionally drawn UI elements' positions
// go back to the start.
func (l *Layout) Reset() {
	l.elementIndex = 0
}

// Advance advances the Layout by the element number given. You shouldn't need to call this, but can
// be useful if you want to skip spaces.
func (l *Layout) Advance(delta int) {
	l.elementIndex += delta
}

// ItemRect returns the current item's rectangle when transformed by the Layout's layout function.
func (l *Layout) ItemRect(drawCall DrawCall) DrawCall {
	res := l.arranger.Arrange(drawCall)
	res.Rect = res.Rect.MoveVec(l.Offset)
	return res
}

// Arranger returns the layout function for the Layout.
func (l *Layout) Arranger() Arranger {
	return l.arranger
}

// SetArranger sets the layout function for laying out elements in the Layout's rectangle.
func (l *Layout) SetArranger(arranger Arranger) *Layout {
	l.arranger = arranger
	l.elementIndex = 0
	return l
}

type customArranger struct {
	ArrangeFunc ArrangeFunc
}

func (c customArranger) Arrange(drawcall DrawCall) DrawCall {
	return c.ArrangeFunc(drawcall)
}

// Sets a custom layouting function for laying out elements in the Layout's rectangle.
func (l *Layout) SetCustomArranger(layoutFunc ArrangeFunc) *Layout {
	return l.SetArranger(customArranger{
		ArrangeFunc: layoutFunc,
	})
}

func (l *Layout) add(id string, drawable UIElement, drawCall DrawCall) DrawCall {

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
		drawCall = l.ItemRect(drawCall)
		drawCall.rectSet = true
	}

	inst.drawable.draw(drawCall) // the state is set here
	inst.currentRect = drawCall.Rect

	emptyRect := l.currentMaxRect.IsZero()

	if emptyRect || inst.currentRect.X < l.currentMaxRect.X {
		l.currentMaxRect = l.currentMaxRect.ScaleLeftTo(inst.currentRect.X)
	}

	if emptyRect || inst.currentRect.Right() > l.currentMaxRect.Right() {
		l.currentMaxRect = l.currentMaxRect.ScaleRightTo(inst.currentRect.Right())
	}

	if emptyRect || inst.currentRect.Y < l.currentMaxRect.Y {
		l.currentMaxRect = l.currentMaxRect.ScaleUpTo(inst.currentRect.Y)
	}

	if emptyRect || inst.currentRect.Bottom() > l.currentMaxRect.Bottom() {
		l.currentMaxRect = l.currentMaxRect.ScaleDownTo(inst.currentRect.Bottom())
	}

	l.Advance(1)

	return drawCall

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
	Padding Vector2
}

func (l ArrangerFull) Arrange(drawCall DrawCall) DrawCall {

	drawCall.Rect.W -= l.Padding.X
	drawCall.Rect.H -= l.Padding.Y
	drawCall.Rect.X += l.Padding.X / 2
	drawCall.Rect.Y += l.Padding.Y / 2

	return drawCall
}

type ArrangerGridDirection int

const (
	ArrangerGridOrderRowMajor = iota
	ArrangerGridOrderColumnMajor
)

// ArrangerGrid arranges UI elements in an easily extendible grid.
// Can be used for single columns or rows by specifying the DivisionSize
// to be 1 or below.
type ArrangerGrid struct {
	OuterPadding   Vector2 // How many pixels should be given as padding between elements and the borders
	ElementPadding Vector2 // How many pixels should be given as padding between elements and each other
	ElementSize    Vector2 // The size of the Elements in the Grid in pixels
	DivisionSize   int     // Number of elements per division (column or row).

	/*
	   Whether elements increase across (RowMajor) or vertically (ColumnMajor).

	   As an example drawing eight (8) elements with a DivisionSize of 3:

	   ArrangerGridOrderRowMajor:

	   [ 0 ] [ 1 ] [ 2 ]

	   [ 3 ] [ 4 ] [ 5 ]

	   [ 6 ] [ 7 ]

	   ArrangerGridOrderColumnMajor:

	   [ 0 ] [ 3 ] [ 6 ]

	   [ 1 ] [ 4 ] [ 7 ]

	   [ 2 ] [ 5 ]
	*/
	DivisionDirection ArrangerGridDirection
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithElementPaddingW(paddingW float32) ArrangerGrid {
	l.ElementPadding.X = paddingW
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithElementPaddingH(paddingH float32) ArrangerGrid {
	l.ElementPadding.Y = paddingH
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithElementPadding(padding float32) ArrangerGrid {
	l.ElementPadding.X = padding
	l.ElementPadding.Y = padding
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithOuterPaddingW(paddingW float32) ArrangerGrid {
	l.OuterPadding.X = paddingW
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithOuterPaddingH(paddingH float32) ArrangerGrid {
	l.OuterPadding.Y = paddingH
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithOuterPadding(padding float32) ArrangerGrid {
	l.OuterPadding.X = padding
	l.OuterPadding.Y = padding
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithElementSizeW(elementWidth float32) ArrangerGrid {
	l.ElementSize.X = elementWidth
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithElementSizeH(elementHeight float32) ArrangerGrid {
	l.ElementSize.Y = elementHeight
	return l
}

// Returns the ArrangerGrid with the given property set.
func (l ArrangerGrid) WithDivisionSize(divisionSize int) ArrangerGrid {
	l.DivisionSize = divisionSize
	return l
}

// Arranges UI elements given a DrawCall.
func (l ArrangerGrid) Arrange(drawCall DrawCall) DrawCall {

	drawCall.Rect.W -= l.OuterPadding.X
	drawCall.Rect.H -= l.OuterPadding.Y

	cellWidth := drawCall.Rect.W / l.ElementSize.X
	cellHeight := drawCall.Rect.H / l.ElementSize.Y

	if l.ElementPadding.X < 0 {
		l.ElementPadding.X = 0
	}

	if l.ElementPadding.Y < 0 {
		l.ElementPadding.Y = 0
	}

	if l.ElementSize.X <= 0 {
		l.ElementSize.X = cellWidth
	}

	if l.ElementSize.Y <= 0 {
		l.ElementSize.Y = cellHeight
	}

	// Add in difference between element size and the cell size if it's been set
	diffWidth := cellWidth - l.ElementSize.X
	diffHeight := cellHeight - l.ElementSize.Y

	if diffWidth < 0 {
		diffWidth = 0
	}

	if diffHeight < 0 {
		diffHeight = 0
	}

	lx := float32(0)
	ly := float32(0)

	elementCount := l.DivisionSize
	if elementCount <= 1 {
		elementCount = 1
	}

	if l.DivisionDirection == ArrangerGridOrderRowMajor {
		lx = float32(drawCall.ElementIndex%elementCount) * l.ElementSize.X
		ly = float32(drawCall.ElementIndex/elementCount) * l.ElementSize.Y
	} else {
		lx = float32(drawCall.ElementIndex/elementCount) * l.ElementSize.X
		ly = float32(drawCall.ElementIndex%elementCount) * l.ElementSize.Y
	}

	elementRect := Rect{
		X: drawCall.Rect.X + lx + (l.ElementPadding.X / 2) + (diffWidth / 2) + (l.OuterPadding.X / 2),
		Y: drawCall.Rect.Y + ly + (l.ElementPadding.Y / 2) + (diffHeight / 2) + (l.OuterPadding.Y / 2),
		W: l.ElementSize.X - l.ElementPadding.X,
		H: l.ElementSize.Y - l.ElementPadding.Y,
	}

	drawCall.Rect = elementRect

	return drawCall

}
