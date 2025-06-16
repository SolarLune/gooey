package gooey

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"

	_ "embed"
)

// UIIDReusePolicyType specifies what should happen when a stateful UI element ID (i.e. a button, checkbox, etc) is reused by a UI element.
type UIIDReusePolicyType int

const (
	UIIDReusePolicyPanic UIIDReusePolicyType = iota // panic() when a UI ID is used multiple times in a single frame. This is the default behavior.
	UIIDReusePolicyWarn                             // log.Println when a UI ID is used multiple times in a single frame.
	UIIDReusePolicyNone                             // Don't do anything when a UI ID is used multiple times in a single frame.
)

// UIIDReusePolicy sets what should happen when an ID is used more than once in a single frame. The default behavior is to panic.
var UIIDReusePolicy UIIDReusePolicyType

var screenBuffer *ebiten.Image
var defaultFont text.Face = text.NewGoXFace(basicfont.Face7x13)

var inputHighlightedUIRect Rect
var areasInUse = []*Area{}
var visibleAreas = []*Area{}

var textBuffer *ebiten.Image
var textStyle TextStyle

var ScrollWheelScrollSpeed = float32(1)

// AnchorPosition sets where an element (like an icon) should be placed within another (say, a button).
type AnchorPosition int

const (
	AnchorTopLeft AnchorPosition = iota
	AnchorTopCenter
	AnchorTopRight
	AnchorCenterLeft
	AnchorCenter
	AnchorCenterRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight

	AnchorTextLeft // Text anchor positions naturally only work for elements that have text (like buttons)
	AnchorTextRight
	AnchorTextAbove
	AnchorTextBelow
)

//go:embed text.kage
var textKage []byte
var textShader *ebiten.Shader

var bgPatternVerts []ebiten.Vertex
var bgPatternIndices []uint16

func init() {
	shader, err := ebiten.NewShader(textKage)
	if err != nil {
		panic(err)
	}
	textShader = shader

	bgPatternVerts = []ebiten.Vertex{
		{},
		{},
		{},
		{},
	}

	bgPatternIndices = []uint16{0, 1, 2, 2, 3, 0}

	SetDefaultTextStyle(NewDefaultTextStyle())
}

// Init initializes the screen buffer; this should only need to be called once in an application, or whenever you need to resize the UI.
func Init(w, h int) {

	if screenBuffer != nil {

		if bounds := screenBuffer.Bounds(); bounds.Dx() == w || bounds.Dy() == h {
			return
		}
		screenBuffer.Deallocate()
		textBuffer.Deallocate()

	}

	screenBuffer = ebiten.NewImage(w, h)
	textBuffer = ebiten.NewImage(w, h)

}

var ended = true

// Begin clears the internal screen buffer; this should be called at the beginning of every frame.
func Begin() {
	if !ended {
		log.Println("Warning: gooey.End() may not have been called.")
	}
	ended = false
	visibleAreas = visibleAreas[:0]
	screenBuffer.Clear()
	idsAccessed = idsAccessed[:0]
	selectableUIIDs = selectableUIIDs[:0]
}

func End() {
	queuedInput = queuedInputNone
	ended = true
}

// Texture returns the rendered texture for all UI elements.
func Texture() *ebiten.Image {
	return screenBuffer
}

// DrawDebug will draw debug elements.
func DrawDebug(screen *ebiten.Image, drawAreaText bool) {

	drawText := func(x, y float32, txt string) {
		opt := &text.DrawOptions{}
		opt.GeoM.Translate(float64(x)+4, float64(y)+4)
		opt.ColorScale.Scale(0, 0, 0, 1)
		text.Draw(screen, txt, defaultFont, opt)
		opt.GeoM.Translate(-1, -1)
		opt.ColorScale.Reset()
		text.Draw(screen, txt, defaultFont, opt)
	}

	for _, area := range visibleAreas {
		x := area.Rect.X + area.parentOffset.X
		y := area.Rect.Y + area.parentOffset.Y

		vector.DrawFilledRect(screen, x, y, area.Rect.W, area.Rect.H, color.RGBA{0, 0, 0, 100}, false)
		vector.StrokeRect(screen, x, y, area.Rect.W, area.Rect.H, 1, color.White, false)

		if drawAreaText {
			drawText(x, y, area.String())
		}

		for id, r := range area.prevPlacedElementRects {
			rx := r.X
			ry := r.Y
			vector.StrokeRect(screen, rx, ry, r.W, r.H, 1, color.White, false)
			drawText(rx, ry, fmt.Sprintf("%v", id))
		}
	}
}

func Reset() {
	areasInUse = areasInUse[:0]
	states = states[:0]
}

type Rect struct {
	X, Y, W, H float32
}

func (r Rect) IsZero() bool {
	return r.X == 0 && r.Y == 0 && r.W == 0 && r.H == 0
}

func (r Rect) Center() Position {
	return Position{
		X: r.X + r.W/2,
		Y: r.Y + r.H/2,
	}
}

func (r Rect) Right() float32 {
	return r.X + r.W
}

func (r Rect) Bottom() float32 {
	return r.Y + r.H
}

func (r Rect) Top() float32 {
	return r.Y
}

func (r Rect) Left() float32 {
	return r.X
}

type Position struct {
	X, Y float32
}

func (p Position) DistanceSquaredTo(other Position) float32 {
	p.X -= other.X
	p.Y -= other.Y
	return p.X*p.X + p.Y*p.Y
}

// type PositionIterator struct {
// 	Positions []Position
// 	Index     int
// }

// func (i *PositionIterator) Next() *Position {
// 	if i.Index >= len(i.Positions) {
// 		return nil
// 	}
// 	pos := i.Positions[i.Index]
// 	i.Index++
// 	return &pos
// }

// func (i *PositionIterator) Prev() *Position {
// 	if i.Index == 0 {
// 		return nil
// 	}
// 	return &i.Positions[i.Index-1]
// }

type Layout interface {
	Reset(a *Area)
	Layout(a *Area, id any) (x, y, absX, absY float32)
}

type LayoutGridDirection int

const (
	LayoutGridDirectionDown LayoutGridDirection = iota
	LayoutGridDirectionRight
	LayoutGridDirectionLeft
	LayoutGridDirectionUp
)

type LayoutGrid struct {
	layoutStartID      int
	ItemWidth          float32             // The width of the items in the grid
	ItemHeight         float32             // The height of the items in the grid
	ItemAnchorPosition AnchorPosition      // The positioning to anchor the items if the size is given
	OffsetX            float32             // The offset (in pixels) to use for UI elements on the X axis
	OffsetY            float32             // The offset (in pixels) to use for UI elements on the Y axis
	Direction          LayoutGridDirection // The direction to move in, either horizontally, making rows, or vertically, making columns.
	// DividerCount is the number of evenly-spaced divisions (columns when vertical, rows when horizontal).
	// For each divider in divider count, the elements rendered are interleaved going right or downwards, ad infinitum.
	// For example, with a DividerCount of 2, Horizontal set to false, and six UI elements to draw, you would end up with:
	//
	// | A | B |
	//
	// | C | D |
	//
	// | E | F |
	//
	//
	// With Horizontal set to true, you would get:
	//
	// | A | C | E |
	//
	// | B | D | F |
	//
	// The easiest way to imagine this is that with a DividerCount of 1,
	// you end up with one column (if Horizontal is false), or one row (if Horizontal is true).
	// The minimum value for DividerCount is 1.
	DividerCount   int
	DividerPadding float32 // The padding (in pixels) between each row or column, depending on the grid's primary / growing axis
}

func (l *LayoutGrid) Reset(a *Area) {
	l.layoutStartID = a.uiIDNumber
}

func (l *LayoutGrid) Layout(a *Area, id any) (x, y, absX, absY float32) {

	x = a.Rect.X
	y = a.Rect.Y

	dividerCount := l.DividerCount
	if dividerCount <= 0 {
		dividerCount = 1
	}

	idNumber := a.uiIDNumber - l.layoutStartID

	slotWidth := a.Rect.W / float32(dividerCount)
	slotHeight := a.Rect.H / float32(dividerCount)

	if l.Direction == LayoutGridDirectionRight || l.Direction == LayoutGridDirectionLeft {

		xd := idNumber / dividerCount
		yd := idNumber % dividerCount

		y += slotHeight * float32(yd)

		width := float32(0)
		if len(a.placedElementRects) > 0 {
			width = a.placedElementRects[len(a.placedElementRects)-1].W + l.DividerPadding
		}

		if l.Direction == LayoutGridDirectionLeft {
			xd *= -1
		}

		x += width * float32(xd)

	} else {

		xd := idNumber % dividerCount
		yd := idNumber / dividerCount

		x += slotWidth * float32(xd)

		height := float32(0)
		if len(a.placedElementRects) > 0 {
			height = a.placedElementRects[len(a.placedElementRects)-1].H + l.DividerPadding
		}

		if l.Direction == LayoutGridDirectionUp {
			yd *= -1
		}

		y += height * float32(yd)

	}

	if l.ItemWidth > 0 {
		if l.ItemAnchorPosition == AnchorTopCenter || l.ItemAnchorPosition == AnchorCenter || l.ItemAnchorPosition == AnchorBottomCenter {
			x += slotWidth/2 - (l.ItemWidth / 2)
		} else if l.ItemAnchorPosition == AnchorTopRight || l.ItemAnchorPosition == AnchorCenterRight || l.ItemAnchorPosition == AnchorBottomRight {
			x += slotWidth - l.ItemWidth
		}
	}

	if l.ItemHeight > 0 {
		if l.ItemAnchorPosition == AnchorCenterLeft || l.ItemAnchorPosition == AnchorCenter || l.ItemAnchorPosition == AnchorCenterRight {
			y += slotHeight/2 - (l.ItemHeight / 2)
		} else if l.ItemAnchorPosition == AnchorBottomLeft || l.ItemAnchorPosition == AnchorBottomCenter || l.ItemAnchorPosition == AnchorBottomRight {
			y += slotHeight - l.ItemHeight
		}
	}

	x += l.OffsetX
	y += l.OffsetY

	absX = x
	absY = y

	if !a.scrollRect.IsZero() {
		absX += a.scrollOffset.X
		absY += a.scrollOffset.Y
	}

	absX += a.parentOffset.X + a.Offset.X
	absY += a.parentOffset.Y + a.Offset.Y

	return

}

// LayoutAddSlide allows you to add a slide motion to specific UI elements (i.e. selected buttons).
type LayoutAddSlide struct {
	Softness float32
	// BaseLayout is a pointer to a base layout to use for laying out UI elements.
	// For example, you can use LayoutGrid to create a base layout, and then customize it using a LayoutCustom.
	BaseLayout Layout

	// LayoutFunc is the layout function to run for each UI element rendered in an Area.
	// It should return delta values of movement for each UI element over time from the original location.
	// uiID is the ID of each element being positioned by the layout.
	// uiIndex is the global index of the UI element.
	// uiRects is a slice of rects indicating the previously-rendered UI elements for this frame.
	LayoutFunc func(uiID any, uiIndex int, uiRects []Rect) (dx, dy float32)
}

func (l *LayoutAddSlide) Reset(a *Area) {
	if l.BaseLayout != nil {
		l.BaseLayout.Reset(a)
	}
}

func (l *LayoutAddSlide) Layout(a *Area, id any) (x, y, absX, absY float32) {

	softness := l.Softness
	if softness <= 0 {
		softness = 0.4
	}

	tx := float32(0)
	ty := float32(0)

	if l.BaseLayout != nil {
		tx, ty, _, _ = l.BaseLayout.Layout(a, id)
	}

	if l.LayoutFunc != nil {
		dx, dy := l.LayoutFunc(id, a.uiIDNumber, a.placedElementRects)
		tx += dx
		ty += dy
	}

	pastRect, ok := a.RectByID(id)

	if ok {
		x = pastRect.X + ((tx - pastRect.X) * softness)
		y = pastRect.Y + ((ty - pastRect.Y) * softness)
	} else {
		x = tx
		y = ty
	}

	absX = x
	absY = y

	if !a.scrollRect.IsZero() {
		absX += a.scrollOffset.X
		absY += a.scrollOffset.Y
	}

	absX += a.parentOffset.X + a.Offset.X
	absY += a.parentOffset.Y + a.Offset.Y

	return

}

type Area struct {
	ID                     string
	uiIDNumber             int
	Offset                 Position
	Rect                   Rect
	scrollOffset           Position
	scrollRect             Rect
	placedElementRects     []Rect
	prevPlacedElementRects map[any]Rect

	HighlightOptions map[string]HighlightChoices
	HighlightActive  bool

	parentOffset Position
	// texture            *ebiten.Image
	scrolling bool
	parent    *Area
	children  []*Area

	suspendLayout bool

	layout Layout
}

// NewArea creates a new Area with the specified id.
// x, y, w, and h are the X and Y position and width and height of the Area in pixels.
func NewArea(id string, x, y, w, h float32) *Area {

	for _, a := range areasInUse {
		if a.ID == id {
			a.placedElementRects = a.placedElementRects[:0]
			a.uiIDNumber = 0
			visibleAreas = append(visibleAreas, a)
			return a
		}
	}

	a := &Area{
		ID: id,
		Rect: Rect{
			x, y, w, h,
		},
		prevPlacedElementRects: map[any]Rect{},
	}
	a.layout = &LayoutGrid{}

	areasInUse = append(areasInUse, a)
	visibleAreas = append(visibleAreas, a)

	return a
}

func NewAreaFromImage(id string, screen *ebiten.Image) *Area {
	screenBounds := screen.Bounds()
	area := NewArea(id, 0, 0, float32(screenBounds.Dx()), float32(screenBounds.Dy()))
	return area
}

func (a *Area) Clone() *Area {
	newArea := *a
	newArea.placedElementRects = append([]Rect{}, a.placedElementRects...)
	newArea.children = make([]*Area, 0, len(a.children))
	for _, c := range a.children {
		newArea.children = append(newArea.children, c.Clone())
	}
	return &newArea
}

func (a *Area) String() string {
	return fmt.Sprintf("%s : { %d, %d, %d, %d } : Scroll : %d, %d", a.ID, int(a.Rect.X), int(a.Rect.Y), int(a.Rect.W), int(a.Rect.H), int(a.scrollOffset.X), int(a.scrollOffset.Y))
}

// appendUIRect essentially says "this is where this object is drawing".
func (a *Area) appendUIRect(id any, rect Rect) {
	if a.suspendLayout {
		return
	}
	a.prevPlacedElementRects[id] = rect
	a.placedElementRects = append(a.placedElementRects,
		rect,
	)
	a.uiIDNumber++
}

// SetLayout sets the active layout (UI element placement tool) for the Area. An Area has only one active layout at a time.
func (a *Area) SetLayout(layout Layout) {
	a.layout = layout
	a.layout.Reset(a)
}

func (a *Area) SetDefaultLayout() {
	a.SetLayout(&LayoutGrid{})
}

// LayoutRow lays out elements into a single horizontal row, with the given padding in pixels between
// UI elements.
// func (a *Area) LayoutRow(elementCount int, padding float32) {
// 	a.FlowDirection = FlowHorizontal
// 	a.FlowElementPadding = padding
// 	a.FlowElementWidth = (a.Rect.W - padding) / float32(elementCount)
// 	a.FlowElementHeight = 0
// 	a.flowStartID = uiIDNumber + 1
// }

// // LayoutColumn lays out elements into a single vertical column, with the given padding in pixels between
// // UI elements.
// func (a *Area) LayoutColumn(elementCount int, padding float32) {
// 	a.FlowDirection = FlowVertical
// 	a.FlowElementPadding = padding
// 	// a.FlowElementPadding = a.Rect.H / float32(elementCount)
// 	// a.FlowElementPadding = padding
// 	// a.FlowElementHeight = (a.Rect.H - padding) / float32(elementCount)
// 	a.FlowElementWidth = 0
// 	a.flowStartID = uiIDNumber + 1
// }

// // LayoutGrid lays out the following UI elements rendered into a grid.
// // cellCountX and cellCountY is how many cells should be in the grid.
// // gridWidth is the width of the grid in pixels. Setting this to <= 0 sets it
// // to the area's width. Controlling this allows you to control which direction the grid scrolls in, if desired.
// // padding is the amount of padding between UI elements in pixels.
// func (a *Area) LayoutGrid(cellCountX, cellCountY int, gridWidth, padding float32) {
// 	a.FlowDirection = FlowGrid
// 	a.FlowElementPadding = padding

// 	a.flowStartID = uiIDNumber + 1
// 	if gridWidth <= 0 {
// 		gridWidth = a.Rect.W
// 	}

// 	a.flowElementXCount = cellCountX

// 	a.flowGridWidth = gridWidth

// 	a.FlowElementWidth = (a.flowGridWidth / float32(cellCountX)) - padding
// 	a.FlowElementHeight = (a.Rect.H / float32(cellCountY)) - padding

// }

// func (a *Area) LayoutFill() {
// 	a.FlowDirection = FlowNone
// 	a.FlowElementPadding = 0
// 	a.FlowElementWidth = a.Rect.W
// 	a.FlowElementHeight = a.Rect.H
// 	a.flowStartID = uiIDNumber + 1
// }

// func (a *Area) LayoutReset() {
// 	a.FlowDirection = FlowVertical
// 	a.FlowElementPadding = 0
// 	a.FlowElementWidth = 0
// 	a.FlowElementHeight = 0
// 	a.flowStartID = -1
// }

// LayoutCustom allows you to easily force a layout where all elements have the same size.
// You would position the elements manually by using area offsetting, for example.
// func (a *Area) LayoutCustom(elementW, elementH float32) {
// 	a.FlowDirection = FlowNone
// 	a.FlowElementPadding = 0
// 	a.FlowElementWidth = elementW
// 	a.FlowElementHeight = elementH
// }

// AlignToArea allows you to reposition an Area to another Area.
// func (a *Area) AlignToArea(other *Area, anchor AnchorPosition) {
// 	// switch anchor {
// 	// 	case Anchor
// 	// }
// }

// AlignToImage aligns an Area to the bounds of the image provided using an AnchorPosition constant,
// with the desired padding in pixels.
// Using any anchor positions that aren't supported will return an error.
func (a *Area) AlignToImage(img *ebiten.Image, anchor AnchorPosition, padding float32) error {
	bounds := img.Bounds()

	switch anchor {
	case AnchorTopLeft:
		a.Rect.X = float32(bounds.Min.X) + padding
		a.Rect.Y = float32(bounds.Min.Y) + padding
	case AnchorTopCenter:
		a.Rect.X = float32(bounds.Min.X) + float32(bounds.Dx())/2 - a.Rect.W/2
		a.Rect.Y = float32(bounds.Min.Y) + padding
	case AnchorTopRight:
		a.Rect.X = float32(bounds.Max.X) - a.Rect.W - padding
		a.Rect.Y = float32(bounds.Min.Y) - padding
	case AnchorCenterLeft:
		a.Rect.X = float32(bounds.Min.X) + padding
		a.Rect.Y = float32(bounds.Min.Y) + float32(bounds.Dy())/2 - a.Rect.H/2
	case AnchorCenter:
		a.Rect.X = float32(bounds.Min.X) + float32(bounds.Dx())/2 - a.Rect.W/2
		a.Rect.Y = float32(bounds.Min.Y) + float32(bounds.Dy())/2 - a.Rect.H/2
	case AnchorCenterRight:
		a.Rect.X = float32(bounds.Max.X) - a.Rect.W - padding
		a.Rect.Y = float32(bounds.Min.Y) + float32(bounds.Dy())/2 - a.Rect.H/2
	case AnchorBottomLeft:
		a.Rect.X = float32(bounds.Min.X) + padding
		a.Rect.Y = float32(bounds.Max.Y) - a.Rect.H - padding
	case AnchorBottomCenter:
		a.Rect.X = float32(bounds.Min.X) + float32(bounds.Dx())/2 - a.Rect.W/2
		a.Rect.Y = float32(bounds.Max.Y) - a.Rect.H - padding
	case AnchorBottomRight:
		a.Rect.X = float32(bounds.Max.X) - a.Rect.W - padding
		a.Rect.Y = float32(bounds.Max.Y) - a.Rect.H - padding
	default:
		return errors.New("can't align area to an image using an unsupported alignment type")
	}

	return nil
}

// AlignToArea aligns an Area to the bounds of the other Area provided using an AnchorPosition constant,
// with the desired padding in pixels.
// Using any anchor positions that aren't supported will return an error.
func (a *Area) AlignToArea(other *Area, anchor AnchorPosition, padding float32) error {

	minX := other.Rect.X
	minY := other.Rect.Y
	maxX := other.Rect.X + other.Rect.W
	maxY := other.Rect.Y + other.Rect.H

	switch anchor {
	case AnchorTopLeft:
		a.Rect.X = float32(minX) + padding
		a.Rect.Y = float32(minY) + padding
	case AnchorTopCenter:
		a.Rect.X = float32(minX) + float32(other.Rect.W)/2 - a.Rect.W/2
		a.Rect.Y = float32(minY) + padding
	case AnchorTopRight:
		a.Rect.X = float32(maxX) - a.Rect.W - padding
		a.Rect.Y = float32(minY) - padding
	case AnchorCenterLeft:
		a.Rect.X = float32(minX) + padding
		a.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - a.Rect.H/2
	case AnchorCenter:
		a.Rect.X = float32(minX) + float32(other.Rect.W)/2 - a.Rect.W/2
		a.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - a.Rect.H/2
	case AnchorCenterRight:
		a.Rect.X = float32(maxX) - a.Rect.W - padding
		a.Rect.Y = float32(minY) + float32(other.Rect.H)/2 - a.Rect.H/2
	case AnchorBottomLeft:
		a.Rect.X = float32(minX) + padding
		a.Rect.Y = float32(maxY) - a.Rect.H - padding
	case AnchorBottomCenter:
		a.Rect.X = float32(minX) + float32(other.Rect.W)/2 - a.Rect.W/2
		a.Rect.Y = float32(maxY) - a.Rect.H - padding
	case AnchorBottomRight:
		a.Rect.X = float32(maxX) - a.Rect.W - padding
		a.Rect.Y = float32(maxY) - a.Rect.H - padding
	default:
		return errors.New("can't align area to an image using an unsupported alignment type")
	}

	return nil
}

func (a *Area) MouseInAreaBounds() bool {
	mouseX, mouseY := ebiten.CursorPosition()
	bounds := a.subscreen().Bounds()
	return mouseX >= bounds.Min.X && mouseX <= bounds.Max.X && mouseY >= bounds.Min.Y && mouseY <= bounds.Max.Y
}

type Grid struct {
	Cells []*Area
}

func (g *Grid) CellByPosition(x, y int) *Area {
	return g.Cells[(y/4)+(x%4)]
}

// TODO: This is an imperfect solution. A true solution would be to do this, but to make it a layout option, I think.
// That way, we could properly navigate using input.
// func (a *Area) UIGrid(id string, cellCountX, cellCountY int, xOffset, yOffset, w, h float32) Grid {

// 	g := Grid{
// 		Cells: []*Area{},
// 	}

// 	cw := w / float32(cellCountX)
// 	ch := h / float32(cellCountY)

// 	for y := 0; y < cellCountY; y++ {
// 		for x := 0; x < cellCountX; x++ {
// 			a := NewArea(id+":"+strconv.Itoa(y)+":"+strconv.Itoa(x), a.Rect.X+(float32(x)*cw), a.Rect.Y+(float32(y)*ch), cw, ch)
// 			a.LayoutFill()
// 			g.Cells = append(g.Cells, a)
// 		}
// 	}

// 	dx, dy, _, _, _, _ := a.Layout.Layout()

// 	a.placedElementRects = append(a.placedElementRects,
// 		&Rect{
// 			X: dx + xOffset,
// 			Y: dy + yOffset,
// 			W: w,
// 			H: h,
// 		},
// 	)

// 	return g
// }

// UIArea creates a sub area to display elements.
// xOffset and yOffset is the position relative to the calling area, and w and h are
// the width and height. All units are in pixels.
// If the width and height are less than or equal to zero, they'll default to the
// original Area's width or height.
func (a *Area) UIArea(id string, xOffset, yOffset, w, h float32) *Area {

	if w <= 0 {
		w = a.Rect.W
	}

	if h <= 0 {
		h = a.Rect.H
	}

	dx, dy, _, _ := a.layout.Layout(a, id)

	newArea := NewArea(id, dx+xOffset, dy+yOffset, w, h)

	r := newArea.Rect
	r.X += xOffset
	r.Y += yOffset
	a.appendUIRect(id, r)

	newArea.parent = a
	a.children = append(a.children, newArea)

	return newArea
}

func (a *Area) subscreen() *ebiten.Image {
	if a.parent != nil {
		return a.parent.subscreen().SubImage(image.Rect(int(a.Rect.X+a.parentOffset.X), int(a.Rect.Y+a.parentOffset.Y), int(a.Rect.X)+int(a.Rect.W), int(a.Rect.Y)+int(a.Rect.H))).(*ebiten.Image)
	}
	return screenBuffer.SubImage(image.Rect(int(a.Rect.X), int(a.Rect.Y), int(a.Rect.X)+int(a.Rect.W), int(a.Rect.Y)+int(a.Rect.H))).(*ebiten.Image)
}

func (a *Area) parentSubscreen() *ebiten.Image {
	if a.parent != nil {
		return a.parent.subscreen().SubImage(image.Rect(int(a.parent.Rect.X+a.parentOffset.X), int(a.parent.Rect.Y+a.parentOffset.Y), int(a.parent.Rect.X)+int(a.parent.Rect.W), int(a.parent.Rect.Y)+int(a.parent.Rect.H))).(*ebiten.Image)
	}
	return screenBuffer
}

func (a *Area) textSubscreen() *ebiten.Image {
	if a.parent != nil {
		return a.parent.textSubscreen().SubImage(image.Rect(int(a.Rect.X+a.parentOffset.X), int(a.Rect.Y+a.parentOffset.Y), int(a.Rect.X)+int(a.Rect.W), int(a.Rect.Y)+int(a.Rect.H))).(*ebiten.Image)
	}
	return textBuffer.SubImage(image.Rect(int(a.Rect.X), int(a.Rect.Y), int(a.Rect.X)+int(a.Rect.W), int(a.Rect.Y)+int(a.Rect.H))).(*ebiten.Image)
}

var HorizontalScrollKey ebiten.Key = ebiten.KeyShift

func (a *Area) HandleScrolling() {

	if len(a.placedElementRects) == 0 {
		a.scrollRect = Rect{}
		return
	}

	start := a.placedElementRects[0]
	end := a.placedElementRects[len(a.placedElementRects)-1]
	a.scrollRect.X = min(start.X, end.X)
	a.scrollRect.Y = min(start.Y, end.Y)
	a.scrollRect.W = max(start.X+start.W, end.X+end.W) - a.scrollRect.X
	a.scrollRect.H = max(start.Y+start.H, end.Y+end.H) - a.scrollRect.Y

	if a.scrollRect.H <= a.Rect.H && a.scrollRect.W <= a.Rect.W {
		a.scrollRect = Rect{}
		return
	}

	scrollbarWidth := float32(8)
	scrollAmount := float32(0)

	if !inputHighlightedUIRect.IsZero() {
		// if a.scrollingVertically() {
		a.scrollOffset.Y = -inputHighlightedUIRect.Center().Y + (a.Rect.H / 2)
		// } else {
		// 	a.scrollOffset.X = -inputHighlightedUIRect.Center().X + (a.Rect.W / 2)
		// }
	}

	// TODO: Allow scrolling horizontally AND vertically

	// if a.scrollingVertically() {
	scrollAmount = a.scrollOffset.Y / -(a.scrollRect.H - a.Rect.H)
	scrollAmount = a.drawScrollbar(a.Rect.Right(), a.Rect.Y, scrollbarWidth, a.Rect.H, scrollAmount)

	if a.MouseInAreaBounds() {

		_, wheel := ebiten.Wheel()

		if wheel > 0 {
			scrollAmount -= 0.1 * ScrollWheelScrollSpeed
		} else if wheel < 0 {
			scrollAmount += 0.1 * ScrollWheelScrollSpeed
		}
		scrollAmount = clamp(scrollAmount, 0, 1)

	}

	a.scrollOffset.Y = -scrollAmount * (a.scrollRect.H - a.Rect.H)

	// } else {
	// 	scrollAmount = a.scrollOffset.X / -(a.scrollRect.W - a.Rect.W)
	// 	scrollAmount = a.drawScrollbar(a.Rect.X, a.Rect.Bottom(), a.Rect.W, scrollbarWidth, scrollAmount)

	// 	if a.MouseInAreaBounds() {

	// 		xWheel, yWheel := ebiten.Wheel()

	// 		if ebiten.IsKeyPressed(HorizontalScrollKey) {
	// 			if yWheel != 0 {
	// 				xWheel = yWheel
	// 			}
	// 		}

	// 		if xWheel > 0 {
	// 			scrollAmount -= 0.1 * ScrollWheelScrollSpeed
	// 		} else if xWheel < 0 {
	// 			scrollAmount += 0.1 * ScrollWheelScrollSpeed
	// 		}
	// 		scrollAmount = clamp(scrollAmount, 0, 1)

	// 	}

	// 	a.scrollOffset.X = -scrollAmount * (a.scrollRect.W - a.Rect.W)
	// }

	for _, c := range a.children {
		c.parentOffset.X = a.scrollOffset.X
		c.parentOffset.Y = a.scrollOffset.Y
	}

	// if scrolling {
	// 	if a.Rect.H >= a.scrollRect.H {
	// 		a.Offset.Y = 0
	// 		return
	// 	}
	// 	a.Offset.Y = a.Rect.Y - scrollTo + (a.Rect.H / 2)
	// 	if a.Offset.Y > 0 {
	// 		a.Offset.Y = 0
	// 	} else if a.Offset.Y < -(a.scrollRect.H - a.Rect.H) {
	// 		a.Offset.Y = -(a.scrollRect.H - a.Rect.H)
	// 	}
	// }

}

func (a *Area) ResetScrolling() {
	a.scrollOffset.Y = 0
	a.scrollOffset.X = 0
}

func (a *Area) SetRect(rect Rect) {

	a.Rect.X = rect.X
	a.Rect.Y = rect.Y

	if a.Rect.W == rect.W && a.Rect.H == rect.H {
		return
	}

	// Default to the existing w / h values
	if rect.W > 0 {
		a.Rect.W = rect.W
	}
	if rect.H > 0 {
		a.Rect.H = rect.H
	}

	// if a.texture != nil {
	// 	a.texture.Deallocate()
	// }

	// a.texture = ebiten.NewImage(int(a.Rect.W), int(a.Rect.H))

}

// func (a *AreaOptions) SetScroll(perc float32) {
// 	if perc < 0 {
// 		perc = 0
// 	} else if perc > 1 {
// 		perc = 1
// 	}
// 	a.scrollPerc = perc
// }

// func ScrollTo(y float32) {
// 	scrolling = true
// 	scrollTo = y
// }

// func ClearScroll() {
// 	scrolling = false
// }

func (a *Area) Split(perc float32, horizontal bool) (left, right *Area) {

	if perc < 0 {
		perc = 0
	}
	if perc > 1 {
		perc = 1
	}

	if horizontal {
		left = NewArea("split_left", a.Rect.X, a.Rect.Y, a.Rect.W*perc, a.Rect.H)
		right = NewArea("split_right", a.Rect.X+left.Rect.W, a.Rect.Y, a.Rect.W*(1-perc), a.Rect.H)
		return
	}
	// else
	left = NewArea("split_bottom", a.Rect.X, a.Rect.Y+left.Rect.H, a.Rect.W, a.Rect.H*(1-perc))
	right = NewArea("split_top", a.Rect.X, a.Rect.Y, a.Rect.W, a.Rect.H*perc)
	return

}

var selectableUIIDs []any
var highlightingUIID any
var focusedUIElement bool // If there's a UI element that's focused

const (
	queuedInputNone = iota
	queuedInputRight
	queuedInputLeft
	queuedInputUp
	queuedInputDown
	queuedInputPrev
	queuedInputNext
	queuedInputSelect
)

var queuedInput = queuedInputNone
var prevQueuedInput = queuedInputNone
var usingMouse = true

// HighlightControlSettings is a struct indicating boolean values used to determine how you cycle through UI elements.
type HighlightControlSettings struct {
	LeftInput   bool
	RightInput  bool
	UpInput     bool
	DownInput   bool
	NextInput   bool
	PrevInput   bool
	AcceptInput bool // Selecting ("clicking") a UI element
	CancelInput bool // Pressing cancel
	UseMouse    bool // Whether or not to use the mouse for selecting and clicking UI elements
}

// HighlightControlRepeatInitialDelay is how long it takes holding a highlight control input to repeat.
var HighlightControlRepeatInitialDelay time.Duration = time.Second / 4

// HighlightControlRepeatDelay is how frequently holding a highlight control input repeats after the initial delay.
var HighlightControlRepeatDelay time.Duration = time.Second / 25

var highlightControlInitialTime time.Time
var highlightControlStartTime time.Time

var inputChars []rune
var regexString string
var caretPos int
var targetText *[]rune

// Update is called every Update() to update the highlighting controls.
// After using this here, you would use HighlightControlBegin() / HighlightControlEnd() to indicate which options
// should have highlighting control enabled.

var repeatTimer time.Time

// Update updates input-related things from gooey.
func Update(settings HighlightControlSettings) {

	if targetText == nil {

		if settings.RightInput {
			queuedInput = queuedInputRight
		}

		if settings.LeftInput {
			queuedInput = queuedInputLeft
		}
		if settings.UpInput {
			queuedInput = queuedInputUp
		}
		if settings.DownInput {
			queuedInput = queuedInputDown
		}

		if settings.NextInput {
			queuedInput = queuedInputNext
		}
		if settings.PrevInput {
			queuedInput = queuedInputPrev
		}

		if settings.AcceptInput {
			queuedInput = queuedInputSelect
		}

		if settings.CancelInput && !focusedUIElement {
			if opt, exists := Highlights[highlightingUIID]; exists && opt.OnCancel != nil {
				opt.OnCancel()
			}
		}

	}

	usingMouse = settings.UseMouse

	if queuedInput != queuedInputNone {

		if queuedInput != prevQueuedInput {

			highlightControlStartTime = time.Now()
			highlightControlInitialTime = time.Now()

			prevQueuedInput = queuedInput

		} else {

			if time.Since(highlightControlInitialTime) < HighlightControlRepeatInitialDelay {
				prevQueuedInput = queuedInput
				queuedInput = queuedInputNone
			} else if time.Since(highlightControlStartTime) < HighlightControlRepeatDelay {
				prevQueuedInput = queuedInput
				queuedInput = queuedInputNone
			} else {
				highlightControlStartTime = time.Now() // Let through one input
			}

		}

	} else {
		prevQueuedInput = queuedInputNone
		queuedInput = queuedInputNone
	}

	if highlightingUIID != nil && !focusedUIElement {

		if opt, exists := Highlights[highlightingUIID]; exists {

			if queuedInput == queuedInputDown && opt.Options[DirectionDown] != nil {
				highlightingUIID = opt.Options[DirectionDown]
			} else if queuedInput == queuedInputUp && opt.Options[DirectionUp] != nil {
				highlightingUIID = opt.Options[DirectionUp]
			} else if queuedInput == queuedInputRight && opt.Options[DirectionRight] != nil {
				highlightingUIID = opt.Options[DirectionRight]
			} else if queuedInput == queuedInputLeft && opt.Options[DirectionLeft] != nil {
				highlightingUIID = opt.Options[DirectionLeft]
			} else if queuedInput == queuedInputNext && opt.Options[DirectionNext] != nil {
				highlightingUIID = opt.Options[DirectionNext]
			} else if queuedInput == queuedInputPrev && opt.Options[DirectionPrev] != nil {
				highlightingUIID = opt.Options[DirectionPrev]
			}

		}

	}

	if targetText != nil {

		text := *targetText

		inputChars = inputChars[:0]
		inputChars = ebiten.AppendInputChars(inputChars)

		if len(inputChars) > 0 {

			ta := string(text[:caretPos])
			tb := string(text[caretPos:])
			for _, c := range inputChars {
				regexOK, _ := regexp.MatchString(regexString, string(c))
				if regexString == "" || regexOK {
					*targetText = append(append([]rune(ta), c), []rune(tb)...)
				}
			}
			caretPos += len(inputChars)

		}

		regexOK, _ := regexp.MatchString(regexString, "\n")
		if regexString == "" || regexOK {
			if (keyPressed(ebiten.KeyEnter) || keyPressed(ebiten.KeyKPEnter) || keyPressed(ebiten.KeyNumpadEnter)) && len(text) > 0 {
				ta := string(text[:caretPos])
				tb := string(text[caretPos:])
				*targetText = append(append([]rune(ta), '\n'), []rune(tb)...)
				caretPos++
			}
		}

		if keyPressed(ebiten.KeyBackspace) && len(text) > 0 {
			ta := string(text[:max(0, caretPos-1)])
			tb := string(text[caretPos:])
			caretPos--
			*targetText = append([]rune(ta), []rune(tb)...)
		}

		if keyPressed(ebiten.KeyDelete) && len(text) > 0 {
			ta := string(text[:max(0, caretPos)])
			tb := string(text[min(len(text), caretPos+1):])
			*targetText = append([]rune(ta), []rune(tb)...)

		}

		if keyPressed(ebiten.KeyRight) {
			caretPos++
		}

		if keyPressed(ebiten.KeyLeft) {
			caretPos--
		}

		// TODO: Go up and down
		// if keyPressed(ebiten.KeyUp) {
		// 	caretPos++
		// }

		// if keyPressed(ebiten.KeyDown) {
		// 	caretPos--
		// }

		if caretPos < 0 {
			caretPos = 0
		}

		if caretPos > len(*targetText) {
			caretPos = len(*targetText)
		}

	}

	targetText = nil

	// Clear highlighting ID
	if settings.UseMouse && highlightingUIID != nil && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		highlightingUIID = nil
		inputHighlightedUIRect = Rect{}
	}

	selectable := false
	for _, id := range selectableUIIDs {
		if highlightingUIID == id {
			selectable = true
			break
		}
	}

	if !selectable {
		highlightingUIID = nil
	}

	if highlightingUIID == nil && len(selectableUIIDs) > 0 {
		highlightingUIID = selectableUIIDs[0]
	}

	focusedUIElement = false

}

func keyPressed(key ebiten.Key) bool {

	if inpututil.IsKeyJustPressed(key) || (ebiten.IsKeyPressed(key) && time.Since(repeatTimer) > time.Second/4) {

		if inpututil.IsKeyJustPressed(key) {
			repeatTimer = time.Now()
		}

		return true

	}

	return false

}

type Direction int

const (
	DirectionUp = iota + 1
	DirectionRight
	DirectionNext
)

const (
	DirectionDown = -iota - 1
	DirectionLeft
	DirectionPrev
)

type HighlightChoices struct {
	Options  map[Direction]any
	OnCancel func()
}

var Highlights map[any]HighlightChoices = map[any]HighlightChoices{}

type HighlightOptions struct {
	Direction     Direction // The direction to use for setting highlighting
	IDs           []any     // The IDs to set highlight options for; this slice goes from one to the next.
	Bidirectional bool      // Whether to make highlights set bidirectional (so if going right from option #1 > #2, you can go left from #2 > #1).
	Loop          bool      // Whether to make highlights loop (so if option #1 > #2 > #3 works, then #3 > #1 also works).
	OnCancel      func()    // The function to call for canceling when one of the IDs provided is selected
}

// SetHighlightOptions sets the options for a group of UI ids to control where the highlight moves.
// If the IDs even between Areas are unique, then this allows you to jump between Areas as well.
func SetHighlightOptions(options HighlightOptions) {

	for _, id := range options.IDs {
		if _, exists := Highlights[id]; !exists {
			Highlights[id] = HighlightChoices{
				Options:  map[Direction]any{},
				OnCancel: options.OnCancel,
			}
		}
	}

	for i := 0; i < len(options.IDs)-1; i++ {

		fromID := options.IDs[i]
		toID := options.IDs[i+1]

		Highlights[fromID].Options[options.Direction] = toID
		if options.Bidirectional {
			Highlights[toID].Options[-options.Direction] = fromID
		}

	}

	if options.Loop {
		fromID := options.IDs[len(options.IDs)-1]
		toID := options.IDs[0]
		Highlights[fromID].Options[options.Direction] = toID
		if options.Bidirectional {
			Highlights[toID].Options[-options.Direction] = fromID
		}
	}
}

type GridHighlightOptions struct {
	XCount        int                // The number of cells horizontally
	YCount        int                // The number of cells vertically
	IDFunc        func(x, y int) any // The function that identifies the ID for the item in the cell provided
	Bidirectional bool               // Whether to make highlights set bidirectional (so if going right from option #1 > #2, you can go left from #2 > #1).
	Loop          bool               // Whether to make highlights loop (so if option #1 > #2 > #3 works, then #3 > #1 also works).
	OnCancel      func()             // The function to call for canceling when one of the IDs provided is selected
	BorderOptions map[Direction]any  // The options to use on the outer edges (for example, going from the top of a grid to an option at the bottom)
}

// SetHighlightOptionsGrid sets the highlight options for items placed in a grid.
// This is doable without this function; it's just a convenience function.
func SetHighlightOptionsGrid(options GridHighlightOptions) {

	ids := []any{}

	for y := 0; y < options.YCount; y++ {

		for x := 0; x < options.XCount; x++ {

			ids = ids[:0]

			for i := 0; i < options.YCount; i++ {

				if opt, ok := options.BorderOptions[DirectionUp]; ok && i == 0 {
					ids = append(ids, opt)
				}

				ids = append(ids, options.IDFunc(x, i))

				if opt, ok := options.BorderOptions[DirectionDown]; ok && i == options.YCount-1 {
					ids = append(ids, opt)
				}

			}

			SetHighlightOptions(HighlightOptions{
				Direction:     DirectionDown,
				Bidirectional: options.Bidirectional,
				Loop:          options.Loop,
				IDs:           ids,
				OnCancel:      options.OnCancel,
			})

		}

		ids = ids[:0]

		for i := 0; i < options.XCount; i++ {

			if opt, ok := options.BorderOptions[DirectionLeft]; ok && i == 0 {
				ids = append(ids, opt)
			}

			ids = append(ids, options.IDFunc(i, y))

			if opt, ok := options.BorderOptions[DirectionRight]; ok && i == options.XCount-1 {
				ids = append(ids, opt)
			}

		}

		SetHighlightOptions(HighlightOptions{
			Direction:     DirectionRight,
			Bidirectional: options.Bidirectional,
			Loop:          options.Loop,
			IDs:           ids,
			OnCancel:      options.OnCancel,
		})

	}

}

func SetHighlight(id any) {
	highlightingUIID = id
}

func SetDefaultHighlight(id any) {
	if highlightingUIID == nil {
		highlightingUIID = id
	}
}

// func (a *Area) HighlightControlBegin() {
// 	inputHighlightedIDStart = uiIDNumber
// }

// func (a *Area) HighlightControlEnd() {

// 	// directionalPrev := (a.FlowDirection == FlowHorizontal && queuedInput == queuedInputLeft) || (a.FlowDirection == FlowVertical && queuedInput == queuedInputUp)
// 	// directionalNext := (a.FlowDirection == FlowHorizontal && queuedInput == queuedInputRight) || (a.FlowDirection == FlowVertical && queuedInput == queuedInputDown)

// 	nextInput := false
// 	prevInput := false
// 	upInput := false
// 	downInput := false

// 	switch a.FlowDirection {
// 	case FlowHorizontal:
// 		nextInput = queuedInput == queuedInputRight || queuedInput == queuedInputNext
// 		prevInput = queuedInput == queuedInputLeft || queuedInput == queuedInputPrev
// 	case FlowVertical:
// 		nextInput = queuedInput == queuedInputDown || queuedInput == queuedInputNext
// 		prevInput = queuedInput == queuedInputUp || queuedInput == queuedInputPrev
// 	case FlowGrid:
// 		nextInput = queuedInput == queuedInputRight || queuedInput == queuedInputNext
// 		prevInput = queuedInput == queuedInputLeft || queuedInput == queuedInputPrev
// 		upInput = queuedInput == queuedInputUp
// 		downInput = queuedInput == queuedInputDown
// 	}

// 	prevInputHighlightedID = inputHighlightIDNumber

// 	if inputHighlightIDNumber == -999 {

// 		if prevInput || nextInput || queuedInput == queuedInputSelect {
// 			inputHighlightIDNumber = inputHighlightedIDStart
// 			prevInputHighlightedID = inputHighlightIDNumber
// 		}

// 	} else {

// 		if prevInput {
// 			inputHighlightIDNumber--
// 		}

// 		if nextInput {
// 			inputHighlightIDNumber++
// 		}

// 		if a.FlowDirection == FlowGrid {

// 			if upInput {
// 				inputHighlightIDNumber -= a.flowElementXCount
// 			}
// 			if downInput {
// 				inputHighlightIDNumber += a.flowElementXCount
// 			}
// 			// fmt.Println(inputHighlightedID, inputHighlightedID%a.flowElementXCount)

// 			if nextInput && inputHighlightIDNumber%a.flowElementXCount == 0 {
// 				inputHighlightIDNumber -= a.flowElementXCount
// 			} else if prevInput && (inputHighlightIDNumber%a.flowElementXCount == a.flowElementXCount-1 || inputHighlightIDNumber < a.flowGridStartID-1) {
// 				inputHighlightIDNumber += a.flowElementXCount
// 			}

// 		}

// 	}

// 	// If inputHighlightedID == -999, that means nothing is highlighted through input, but rather
// 	// UI is being selected by mouse input
// 	if inputHighlightIDNumber != -999 {

// 		if inputHighlightIDNumber < inputHighlightedIDStart {
// 			inputHighlightIDNumber += uiIDNumber - 1
// 		} else if inputHighlightIDNumber >= uiIDNumber {
// 			inputHighlightIDNumber -= uiIDNumber - inputHighlightedIDStart
// 		}

// 	}
// 	queuedInput = queuedInputNone

// }

// WithOffset runs the function provided with a temporary offset of the given x and y values.
// Once the function finishes, the offset is returned to its original value.
func (a *Area) WithOffset(x, y float32, uiFunc func()) {

	ogX := x
	ogY := y

	a.Offset.X = x
	a.Offset.Y = y

	uiFunc()

	a.Offset.X = ogX
	a.Offset.Y = ogY

}

// WithAbsolutePositioning runs the function provided with a temporary reset to all offsetting, layouting, and rectangle positioning.
// This allows you to easily draw a UI element exactly where you want it to be (i.e. over another element).
func (a *Area) WithAbsolutePositioning(uiFunc func(areaRect *Rect)) {

	ogOffset := a.Offset
	ogLayout := a.layout
	ogRect := a.Rect
	ogIDNumber := a.uiIDNumber

	a.layout = &LayoutGrid{}
	a.Offset = Position{}
	a.uiIDNumber += 10000

	ogElements := a.placedElementRects[:]
	a.placedElementRects = a.placedElementRects[:0]

	uiFunc(&a.Rect)

	a.placedElementRects = ogElements

	a.layout = ogLayout
	a.Offset = ogOffset
	a.Rect = ogRect
	a.uiIDNumber = ogIDNumber

}

func HighlightedUIElementID() any {
	if highlightingUIID != nil {
		return highlightingUIID
	}
	return nil
}

// HighlightedUIElementRect returns a rect showing the UI element currently highlighted by keyboard / gamepad input.
func HighlightedUIElementRect() *Rect {
	if highlightingUIID != nil {
		return &inputHighlightedUIRect
	}
	return nil
}

func (a *Area) ForEachRect(scrolling bool, f func(index, maxCount int, rect Rect)) {
	maxCount := len(a.placedElementRects)
	for index, rect := range a.placedElementRects {
		if scrolling {

			if !a.scrollRect.IsZero() {
				rect.X += a.scrollOffset.X
				rect.Y += a.scrollOffset.Y
			}

			rect.X += a.parentOffset.X + a.Offset.X
			rect.Y += a.parentOffset.Y + a.Offset.Y
		}
		f(index, maxCount, rect)
	}
}

func (a *Area) RectByID(id any) (Rect, bool) {
	r, ok := a.prevPlacedElementRects[id]
	if ok {
		return r, true
	}
	return Rect{}, false
}

type RepeatMode int

const (
	RepeatModeRelease RepeatMode = iota
	RepeatModeHold
)

// ButtonOptions is a struct
type ButtonOptions struct {
	Background     ImageDrawer // A reference to any ImageDrawer implementer to render behind this element
	BaseColor      Color       // The base color of the Button
	HighlightColor Color       // The highlight color for the Button. This color is used to draw the Button when the mouse hovers over the button or the button is highlighted using keyboard / gamepad input)
	ClickColor     Color       // The click color for the button. This color is used to draw the button when the mouse is clicking on the button
	ToggleColor    Color       // The color of the button when toggled (i.e. "clicked in"). If empty, it defaults to the click color.

	Text               string  // The text label rendered over the button
	PaddingX, PaddingY float32 // Horizontal and vertical padding (in pixels)
	Width, Height      float32 // The width and height for the Button (in pixels); if set to 0 (the default), the button will expand to fit the text

	Icon *ebiten.Image // The Icon used for the Button
	// IconColor                  Color          // The color used for blending the icon
	IconAnchor                 AnchorPosition // The AnchorPosition used for the Icon
	IconPaddingX, IconPaddingY float32
	// IconOffset Position
	IconDrawOptions *ebiten.DrawImageOptions

	RepeatMode         RepeatMode
	RepeatInitialDelay time.Duration
	RepeatDelay        time.Duration
	Toggle             bool

	TextStyle TextStyle // The text style to use to overwrite the default style.
}

// WithText returns a copy of the ButtonOptions with the label set to the specified
// textStr string. If args are provided the text string is formatted using those arguments.
func (b ButtonOptions) WithText(textStr string, args ...any) ButtonOptions {
	if len(args) > 0 {
		b.Text = fmt.Sprintf(textStr, args...)
	} else {
		b.Text = textStr
	}
	return b
}

// WithIcon returns a copy of the ButtonOptions with the icon provided.
func (b ButtonOptions) WithIcon(icon *ebiten.Image) ButtonOptions {
	b.Icon = icon
	return b
}

// WithIconDrawOptions returns a copy of the ButtonOptions with the icon draw options provided.
func (b ButtonOptions) WithIconDrawOptions(opt *ebiten.DrawImageOptions) ButtonOptions {
	b.IconDrawOptions = opt
	return b
}

// WithToggle returns a copy of the ButtonOptions set to enable or disable toggling according
// to the toggleEnabled argument.
func (b ButtonOptions) WithToggle(toggleEnabled bool) ButtonOptions {
	b.Toggle = toggleEnabled
	return b
}

// WithPadding returns a copy of the ButtonOptions object with the padding set on both axes.
func (t ButtonOptions) WithPadding(padding float32) ButtonOptions {
	t.PaddingX = padding
	t.PaddingY = padding
	return t
}

// WithPaddingX returns a copy of the ButtonOptions object with the padding set on the X axis.
func (t ButtonOptions) WithPaddingX(padding float32) ButtonOptions {
	t.PaddingX = padding
	return t
}

// WithPaddingY returns a copy of the ButtonOptions object with the padding set on the Y axis.
func (t ButtonOptions) WithPaddingY(padding float32) ButtonOptions {
	t.PaddingY = padding
	return t
}

func (t ButtonOptions) WithWidth(width float32) ButtonOptions {
	t.Width = width
	return t
}

func (t ButtonOptions) WithHeight(height float32) ButtonOptions {
	t.Height = height
	return t
}

// func (t ButtonOptions) WithAutomaticPadding(maxWidth, maxHeight float32) ButtonOptions {
// 	if maxWidth > 0 {
// 		t.PaddingX = (maxWidth - t.Width)
// 		if t.PaddingX < 0 {
// 			t.PaddingX = 0
// 		}
// 	}
// 	if maxHeight > 0 {
// 		t.PaddingY = (maxHeight - t.Height)
// 		if t.PaddingY < 0 {
// 			t.PaddingY = 0
// 		}
// 	}
// 	return t
// }

func (t ButtonOptions) WithSize(width, height float32) ButtonOptions {
	t.Width = width
	t.Height = height
	return t
}

func (t ButtonOptions) WidthWithPadding() float32 {
	if t.Width <= 0 {
		return 0
	}
	return t.Width + (t.PaddingX)
}

func (t ButtonOptions) HeightWithPadding() float32 {
	if t.Height <= 0 {
		return 0
	}
	return t.Height + (t.PaddingY)
}

type ButtonState struct {
	id               any
	WasPressed       bool
	Toggled          bool
	InitialClickTime time.Time
	ClickTime        time.Time
}

func (b *ButtonState) ID() any {
	return b.id
}

// UIButton draws a button in the Area, utilizing the ButtonOptions to style and shape the Button.
// The id should be a unique identifying variable (string, number, whatever), used to identify and
// keep track of the button's internal state.
func (a *Area) UIButton(id any, options ButtonOptions) bool {

	// TODO: Support manual newlines and automatic newlines for Buttons

	s := internalStateAccessOnce(id)

	if s == nil {
		s = &ButtonState{id: id}
		states = append(states, s)
	}

	state := s.(*ButtonState)

	subscreen := a.subscreen()
	bounds := subscreen.Bounds()

	// bounds := a.texture.Bounds()

	ogStyle := textStyle

	if !options.TextStyle.IsZero() {
		SetDefaultTextStyle(options.TextStyle)
	}

	x, y, absX, absY := a.layout.Layout(a, id)

	labelW, labelH := text.Measure(options.Text, textStyle.Font, textStyle.lineHeight)

	w := options.Width + options.PaddingX
	h := options.Height + options.PaddingY

	if options.Width == 0 {
		w = float32(labelW) + options.PaddingX
	}
	if options.Height == 0 {
		h = float32(labelH) + options.PaddingY
	}

	if options.Icon != nil {

		iconBounds := options.Icon.Bounds()

		if w < float32(iconBounds.Dx()) {
			w = float32(iconBounds.Dx())
		}

		if h < float32(iconBounds.Dy()) {
			h = float32(iconBounds.Dy())
		}

	}

	a.appendUIRect(id, Rect{
		x, y, w, h,
	})

	if id == highlightingUIID {
		inputHighlightedUIRect = Rect{x, y, w, h}
	}

	x = absX
	y = absY

	mouseX, mouseY := ebiten.CursorPosition()

	baseColor := options.BaseColor
	clickColor := options.ClickColor
	highlightColor := options.HighlightColor

	if baseColor.IsZero() {
		baseColor = NewColor(0.7, 0.7, 0.7, 1)
	}

	if highlightColor.IsZero() {
		highlightColor = baseColor.AddRGBA(0.3, 0.3, 0.3, 0)
	}

	if clickColor.IsZero() {
		clickColor = baseColor.SubRGBA(0.3, 0.3, 0.3, 0)
	}

	buttonColor := baseColor

	if options.Toggle && state.Toggled {

		if options.ToggleColor.IsZero() {
			buttonColor = clickColor
		} else {
			buttonColor = options.ToggleColor
		}

	}

	clicked := false

	inputSelect := queuedInput == queuedInputSelect

	mouseInAreaBounds := mouseX >= bounds.Min.X && mouseX <= bounds.Max.X && mouseY >= bounds.Min.Y && mouseY <= bounds.Max.Y

	now := time.Now()

	if !a.scrolling && (mouseOverArea(mouseX, mouseY, x, y, w, h) && mouseInAreaBounds) || id == highlightingUIID {
		buttonColor = highlightColor

		if mousePressed() || inputSelect {

			buttonColor = clickColor

			if !state.WasPressed {
				state.WasPressed = true
				state.ClickTime = now
				state.InitialClickTime = now
			}

		}

	}

	switch options.RepeatMode {

	case RepeatModeRelease:

		if !mousePressed() && !inputSelect && state.WasPressed {

			if mouseOverArea(mouseX, mouseY, x, y, w, h) || highlightingUIID != nil {
				clicked = true
			}
			state.WasPressed = false

		}

	case RepeatModeHold:

		if (mouseOverArea(mouseX, mouseY, x, y, w, h) && state.WasPressed) || inputSelect {

			buttonColor = clickColor
			if options.RepeatDelay == 0 || state.InitialClickTime == now || (time.Since(state.InitialClickTime) > options.RepeatInitialDelay && time.Since(state.ClickTime) > options.RepeatDelay) {
				if options.RepeatDelay > 0 {
					state.ClickTime = time.Now()
				}
				clicked = true
			}

		}

		if (!mousePressed() || !mouseOverArea(mouseX, mouseY, x, y, w, h)) && !inputSelect {
			state.WasPressed = false
			// state.ClickTime = time.Time{} // Zero it out
		}

	}

	if options.Toggle && clicked {
		state.Toggled = !state.Toggled
	}

	// screen := a.texture.SubImage(image.Rect(int(a.Offset.X), int(a.Offset.Y), int(a.Rect.W), int(a.Rect.H))).(*ebiten.Image)

	textDrawOptions := &text.DrawOptions{}
	textDrawOptions.GeoM.Translate(float64(x), float64(y))
	textDrawOptions.GeoM.Translate(float64(w/2)-(labelW/2), float64(h/2)-(labelH/2))

	// textDrawOptions.ColorScale.ScaleWithColor(buttonColor.ToNRGBA64())
	textDrawOptions.ColorScale.ScaleWithColor(buttonColor.ToNRGBA64())

	if options.Background != nil {
		options.Background.Draw(subscreen, buttonColor.ToNRGBA64(), x, y, w, h)
	}

	if options.Icon != nil {

		iconDrawImageOptions := options.IconDrawOptions
		if options.IconDrawOptions == nil {
			iconDrawImageOptions = &ebiten.DrawImageOptions{}
		}

		// De-reference it to make a copy, making it not possible to destructively edit it
		opt := *iconDrawImageOptions

		opt.ColorScale.ScaleWithColor(buttonColor.ToNRGBA64())

		// opt.GeoM.Translate(float64(x), float64(y))

		bounds := options.Icon.Bounds()
		iconW := float32(bounds.Dx())
		iconH := float32(bounds.Dy())

		switch options.IconAnchor {
		case AnchorTopLeft:
			opt.GeoM.Translate(float64(x+options.IconPaddingX), float64(y+options.IconPaddingY))
		case AnchorTopCenter:
			opt.GeoM.Translate(float64(x+(w/2)-(iconW/2)), float64(y+options.IconPaddingY))
		case AnchorTopRight:
			opt.GeoM.Translate(float64(x+w-iconW-options.IconPaddingX), float64(y+options.IconPaddingY))

		case AnchorCenterLeft:
			opt.GeoM.Translate(float64(x+options.IconPaddingX), float64(y+(h/2)-(iconH/2)))
		case AnchorCenter:
			opt.GeoM.Translate(float64(x+(w/2)-(iconW/2)), float64(y+(h/2)-(iconH/2)))
		case AnchorCenterRight:
			opt.GeoM.Translate(float64(x+w-iconW-options.IconPaddingX), float64(y+(h/2)-(iconH/2)))

		case AnchorBottomLeft:
			opt.GeoM.Translate(float64(x+options.IconPaddingX), float64(y+(h-iconH-options.IconPaddingY)))
		case AnchorBottomCenter:
			opt.GeoM.Translate(float64(x+(w/2)-(iconW/2)), float64(y+(h-iconH-options.IconPaddingY)))
		case AnchorBottomRight:
			opt.GeoM.Translate(float64(x+w-iconW-options.IconPaddingX), float64(y+(h-iconH-options.IconPaddingY)))
		case AnchorTextLeft:
			opt.GeoM.Translate(float64(x), float64(y))
			opt.GeoM.Translate(float64(w/2)-(labelW/2)-float64(iconW)-float64(options.IconPaddingX), float64(h/2)-float64(iconH/2))
		case AnchorTextRight:
			opt.GeoM.Translate(float64(x), float64(y))
			opt.GeoM.Translate(float64(w/2)+(labelW/2)+float64(options.IconPaddingX), float64(h/2)-float64(iconH/2))
		case AnchorTextAbove:
			opt.GeoM.Translate(float64(x), float64(y))
			opt.GeoM.Translate(float64(w/2)-float64(iconW/2), float64(h/2)-(labelH/2)-float64(iconH)-float64(options.IconPaddingY))
		case AnchorTextBelow:
			opt.GeoM.Translate(float64(x), float64(y))
			opt.GeoM.Translate(float64(w/2)-float64(iconW/2), float64(h/2)+(labelH/2)+float64(options.IconPaddingY))

		}

		// opt.GeoM.Translate(float64(options.IconOffset.X), float64(options.IconOffset.Y))

		// if !options.IconColor.IsZero() {
		// 	opt.ColorScale.Scale(options.IconColor.ToFloat32s())
		// }

		subscreen.DrawImage(options.Icon, &opt)

	}

	a.drawTextClear()
	a.drawText(options.Text, textDrawOptions)
	a.drawTextFlush(textDrawOptions)

	if !options.TextStyle.IsZero() {
		SetDefaultTextStyle(ogStyle)
	}

	selectableUIIDs = append(selectableUIIDs, id)

	return clicked
}

type SliderOptions struct {
	Background ImageDrawer
	LineImage  ImageDrawer
	HeadImage  *ebiten.Image

	StepSize   float32
	HeadMargin float32
	Width      float32
	Height     float32

	BaseColor      Color
	HighlightColor Color
	FocusedColor   Color // The color for the slider when activated

	// LineColor and LineThickness are only used if an image isn't used
	LineColor     Color
	LineThickness float32
	ValuePointer  *float32
}

func (s SliderOptions) WithWidth(w float32) SliderOptions {
	s.Width = w
	return s
}

func (s SliderOptions) WithHeight(h float32) SliderOptions {
	s.Height = h
	return s
}

func (s SliderOptions) WithHeadImage(headImg *ebiten.Image) SliderOptions {
	s.HeadImage = headImg
	return s
}

func (s SliderOptions) WithHeadMargin(margin float32) SliderOptions {
	s.HeadMargin = margin
	return s
}

func (s SliderOptions) WithBackground(bg ImageDrawer) SliderOptions {
	s.Background = bg
	return s
}

func (s SliderOptions) WithLineImage(img ImageDrawer) SliderOptions {
	s.LineImage = img
	return s
}

func (s SliderOptions) WithStepSize(stepSize float32) SliderOptions {
	s.StepSize = stepSize
	return s
}

func (s SliderOptions) WithLineColor(color Color) SliderOptions {
	s.LineColor = color
	return s
}

func (s SliderOptions) WithLineThickness(thickness float32) SliderOptions {
	s.LineThickness = thickness
	return s
}

func (s SliderOptions) WithPointer(perc *float32) SliderOptions {
	s.ValuePointer = perc
	return s
}

// func (s SliderOptions) WithDefaultValue(value float32) SliderOptions {
// 	s.DefaultValue = value
// 	return s
// }

type SliderState struct {
	id         any
	perc       *float32
	visualperc float32

	Focused bool // If the slider is focused (being clicked) or not
	HeadX   float32
	HeadY   float32
}

func (s *SliderState) ID() any {
	return s.id
}

func (s *SliderState) Percentage() float32 {
	return *s.perc
}

func (s *SliderState) SetPercentageVariable(perc *float32) {
	s.perc = perc
}

// UIButton draws a button in the Area, utilizing the ButtonOptions to style and shape the Button.
// The id should be a unique identifying variable (string, number, whatever), used to identify and
// keep track of the button's internal state.
func (a *Area) UISlider(id any, options SliderOptions) float32 {

	// TODO: Support manual newlines and automatic newlines for Buttons

	s := internalStateAccessOnce(id)

	if s == nil {
		p := float32(0)
		perc := &p
		if options.ValuePointer != nil {
			perc = options.ValuePointer
		}
		s = &SliderState{id: id, perc: perc}
		states = append(states, s)
	}

	state := s.(*SliderState)

	subscreen := a.subscreen()
	bounds := subscreen.Bounds()

	// bounds := a.texture.Bounds()

	x, y, absX, absY := a.layout.Layout(a, id)

	w := options.Width
	h := options.Height

	if options.Width == 0 {
		w = 128
	}
	if options.Height == 0 {
		h = 32
	}

	a.appendUIRect(id, Rect{
		x, y, w, h,
	})

	if id == highlightingUIID {
		inputHighlightedUIRect = Rect{x, y, w, h}
	}

	x = absX
	y = absY

	mouseX, mouseY := ebiten.CursorPosition()

	baseColor := options.BaseColor
	highlightColor := options.HighlightColor
	focusedColor := options.FocusedColor

	if baseColor.IsZero() {
		baseColor = NewColor(0.7, 0.7, 0.7, 1)
	}

	if highlightColor.IsZero() {
		highlightColor = baseColor.AddRGBA(0.2, 0.2, 0.2, 0)
	}

	if focusedColor.IsZero() {
		focusedColor = baseColor.AddRGBA(0.3, 0.3, 0.3, 0)
	}

	color := baseColor

	inputSelect := queuedInput == queuedInputSelect

	mouseInAreaBounds := mouseX >= bounds.Min.X && mouseX <= bounds.Max.X && mouseY >= bounds.Min.Y && mouseY <= bounds.Max.Y

	if !a.scrolling && !state.Focused {

		if usingMouse {

			if mouseOverArea(mouseX, mouseY, x, y, w, h) && mouseInAreaBounds {

				color = highlightColor
				if mousePressed() || inputSelect {
					state.Focused = true
					focusedUIElement = true
					inputSelect = false
				}

			}

		} else if id == highlightingUIID {
			color = highlightColor
			if inputSelect {
				state.Focused = true
				focusedUIElement = false
				inputSelect = false
			}
		}

	}

	margin := float32(options.HeadMargin)

	lineStartX := x + margin
	lineEndX := x + w - margin
	lineY := y + h/2

	headImgWidth := float32(0)

	if options.HeadImage != nil {
		headImgWidth = float32(options.HeadImage.Bounds().Dx()) / 2
		lineStartX += headImgWidth
		lineEndX -= headImgWidth
	}

	// maxValue := options.MaxValue
	// minValue := options.MinValue

	// if minValue == 0 && options.MaxValue == 0 {
	// 	maxValue = 1
	// }

	perc := state.perc

	if state.Focused {

		color = focusedColor

		focusedUIElement = true

		stepSize := options.StepSize
		if stepSize == 0 {
			stepSize = 0.05
		}

		if usingMouse {
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				*perc = (float32(mouseX) - lineStartX) / (lineEndX - lineStartX)
			} else {
				state.Focused = false
			}
		} else {
			if queuedInput == queuedInputRight {
				*perc += stepSize
			} else if queuedInput == queuedInputLeft {
				*perc -= stepSize
			}

			if inputSelect {
				state.Focused = false
			}
		}

		if *perc > 1 {
			*perc = 1
		} else if *perc < 0 {
			*perc = 0
		}

		if !state.Focused {
			focusedUIElement = false
		}

		if stepSize > 0 {
			*perc = float32(math.Round(float64(*perc/stepSize))) * stepSize
		}

	}

	state.visualperc += (*perc - state.visualperc) * 0.5

	// textDrawOptions := &text.DrawOptions{}
	// textDrawOptions.GeoM.Translate(float64(x), float64(y))
	// textDrawOptions.GeoM.Translate(float64(w/2)-(labelW/2), float64(h/2)-(labelH/2))

	// textDrawOptions.ColorScale.ScaleWithColor(buttonColor.ToNRGBA64())

	if options.Background != nil {
		options.Background.Draw(subscreen, color.ToNRGBA64(), x, y, w, h)
	}

	if options.LineImage != nil {
		options.LineImage.Draw(subscreen, color.ToNRGBA64(), x, y, w, h)
	} else {

		thickness := options.LineThickness

		if thickness == 0 {
			thickness = 2
		}

		if thickness > 0 {
			vector.StrokeLine(subscreen, lineStartX, lineY, lineEndX, lineY, thickness, options.LineColor.ToNRGBA64(), true)
		}

	}

	if options.HeadImage != nil {
		opt := &colorm.DrawImageOptions{}
		sx := float32(lineStartX+(float32(lineEndX-lineStartX)*state.visualperc)) - float32(headImgWidth)
		sy := float32(lineY - float32(options.HeadImage.Bounds().Dy())/2)
		opt.GeoM.Translate(float64(sx), float64(sy))
		cm := colorm.ColorM{}
		cm.ScaleWithColor(color.ToNRGBA64())
		colorm.DrawImage(subscreen, options.HeadImage, cm, opt)
		state.HeadX = sx
		state.HeadY = sy
	}

	selectableUIIDs = append(selectableUIIDs, id)

	return *state.perc

}

type SpacerOptions struct {
	Width, Height float32
}

func (a *Area) UISpacer(id any, options SpacerOptions) {

	x, y, _, _ := a.layout.Layout(a, id)

	w := options.Width
	h := options.Height

	a.appendUIRect(id, Rect{
		x, y, w, h,
	})

}

func (a *Area) drawTextClear() {
	textBuffer.Clear()
}

func (a *Area) drawText(textStr string, textDrawOptions *text.DrawOptions) {
	text.Draw(a.textSubscreen(), textStr, textStyle.Font, textDrawOptions)
}

func (a *Area) drawTextFlush(textDrawOptions *text.DrawOptions) {

	rounded := float32(0)
	if textStyle.OutlineRounded {
		rounded = 1
	}

	shadX := textStyle.ShadowDirectionX
	shadY := textStyle.ShadowDirectionY

	shadX, shadY = vecNormalized(shadX, shadY)

	uniformMap := map[string]interface{}{
		"OutlineThickness": float32(textStyle.OutlineThickness),
		"OutlineRounded":   rounded,
		"ShadowVector":     [2]float32{shadX, shadY},
		"ShadowLength":     float32(textStyle.ShadowLength),
		"TextColor":        textStyle.TextColor.MultiplyRGBA(textDrawOptions.ColorScale.R(), textDrawOptions.ColorScale.G(), textDrawOptions.ColorScale.B(), textDrawOptions.ColorScale.A()).ToFloat32Slice(),
		"OutlineColor":     textStyle.OutlineColor.ToFloat32Slice(),
		"ShadowColorNear":  textStyle.ShadowColorNear.ToFloat32Slice(),
		"ShadowColorFar":   textStyle.ShadowColorFar.ToFloat32Slice(),
	}

	if !textStyle.ShadowColorFar.IsZero() {
		uniformMap["ShadowColorFarSet"] = 1.0
	}

	screenBuffer.DrawRectShader(textBuffer.Bounds().Dx(), textBuffer.Bounds().Dy(), textShader, &ebiten.DrawRectShaderOptions{
		Images: [4]*ebiten.Image{
			textBuffer,
		},
		Uniforms: uniformMap,
	})

	// screenBuffer.DrawImage(textBuffer, nil)
}

type ImageDrawer interface {
	Draw(screen *ebiten.Image, tint color.Color, x, y, w, h float32)
}

type BGImage struct {
	Background  ImageDrawer                  // A reference to any ImageDrawer implementer to render behind this BGImage
	Image       *ebiten.Image                // The source image to draw
	DrawOptions *colorm.DrawTrianglesOptions // Controls some of the rendering features for the image
	ColorM      colorm.ColorM                // The colorm.ColorM matrix for transforming the color of the image
	SrcGeoM     *ebiten.GeoM                 // The source GeoM object for transforming the source texture UV values for the BGImage
	DstGeoM     *ebiten.GeoM                 // The source GeoM object for transforming the destination drawn locations for the BGImage

	// When set to true, TileOverDst sets a flag to automatically tile the image's source texture coordinates
	// over the target render location and size.
	//
	// Here's an example. Let's say you have a Textbox that uses a BGImage as a background.
	// The BGImage is 128x128 and the textbox covers an area of 360x160.
	// If BGImage.TileOverDst == false, the image (which covers 128x128) will be stretched over the area.
	// If BGImage.TileOverDst == true (and DrawOptions.Address is set to ebiten.AddressRepeat), the image UV values
	// will be extended over the area (so it will extend over 360x160).
	// This is, in line with the naming of the variable, best used for patterns that tile behind UI elements.
	TileOverDst bool
}

func NewBGImage(img *ebiten.Image) *BGImage {
	return &BGImage{
		Image:       img,
		DrawOptions: &colorm.DrawTrianglesOptions{},
		ColorM:      colorm.ColorM{},
		DstGeoM:     &ebiten.GeoM{},
		SrcGeoM:     &ebiten.GeoM{},
	}
}

func (img *BGImage) Draw(screen *ebiten.Image, tint color.Color, x, y, w, h float32) {

	if img.Background != nil {
		img.Background.Draw(screen, tint, x, y, w, h)
	}

	srcGeom := *img.SrcGeoM
	dstGeom := *img.DstGeoM
	opt := *img.DrawOptions

	if img.TileOverDst {
		srcGeom.Scale(float64(w)/float64(img.Image.Bounds().Dx()), float64(h)/float64(img.Image.Bounds().Dy()))
	}

	cm := img.ColorM // Make a copy for tinting

	if tint != nil {
		cm.ScaleWithColor(tint)
	}

	for i := range bgPatternVerts {
		bgPatternVerts[i].ColorR = 1
		bgPatternVerts[i].ColorG = 1
		bgPatternVerts[i].ColorB = 1
		bgPatternVerts[i].ColorA = 1
	}

	bgPatternVerts[0].DstX = x
	bgPatternVerts[0].DstY = y

	bgPatternVerts[1].DstX = x + w
	bgPatternVerts[1].DstY = y

	bgPatternVerts[2].DstX = x + w
	bgPatternVerts[2].DstY = y + h

	bgPatternVerts[3].DstX = x
	bgPatternVerts[3].DstY = y + h

	for i := range bgPatternVerts {
		x, y := dstGeom.Apply(float64(bgPatternVerts[i].DstX), float64(bgPatternVerts[i].DstY))
		bgPatternVerts[i].DstX = float32(x)
		bgPatternVerts[i].DstY = float32(y)
	}

	bounds := img.Image.Bounds()

	bgPatternVerts[0].SrcX = float32(bounds.Min.X)
	bgPatternVerts[0].SrcY = float32(bounds.Min.Y)

	bgPatternVerts[1].SrcX = float32(bounds.Max.X)
	bgPatternVerts[1].SrcY = float32(bounds.Min.Y)

	bgPatternVerts[2].SrcX = float32(bounds.Max.X)
	bgPatternVerts[2].SrcY = float32(bounds.Max.Y)

	bgPatternVerts[3].SrcX = float32(bounds.Min.X)
	bgPatternVerts[3].SrcY = float32(bounds.Max.Y)

	for i := range bgPatternVerts {
		x, y := srcGeom.Apply(float64(bgPatternVerts[i].SrcX), float64(bgPatternVerts[i].SrcY))
		bgPatternVerts[i].SrcX = float32(x)
		bgPatternVerts[i].SrcY = float32(y)
	}

	colorm.DrawTriangles(screen, bgPatternVerts, bgPatternIndices, img.Image, cm, &opt)

}

type BGNinepatch struct {
	Background ImageDrawer // A reference to any ImageDrawer implementer to render behind this element
	Image      *ebiten.Image
	ColorM     colorm.ColorM
	Options    *colorm.DrawImageOptions
}

func NewBGNinepatch(img *ebiten.Image) *BGNinepatch {
	return &BGNinepatch{
		Image:   img,
		ColorM:  colorm.ColorM{},
		Options: &colorm.DrawImageOptions{},
	}
}

func (img *BGNinepatch) Draw(screen *ebiten.Image, tint color.Color, x, y, w, h float32) {
	if img.Background != nil {
		img.Background.Draw(screen, tint, x, y, w, h)
	}

	cm := img.ColorM
	if tint != nil {
		cm.ScaleWithColor(tint)
	}
	DrawNinepatch(screen, img.Image, x, y, w, h, cm, img.Options)
}

// BGThreePatch draws a three-patch image (i.e. an image that is composed of a left side,
// stretched middle, and right side, or a top, stretched middle, and bottom).
type BGThreePatch struct {
	Background ImageDrawer // A reference to any ImageDrawer implementer to render behind this element
	Image      *ebiten.Image
	ColorM     colorm.ColorM
	Options    *colorm.DrawImageOptions
	Horizontal bool
}

func NewBGThreePatch(img *ebiten.Image, horizontal bool) *BGThreePatch {
	return &BGThreePatch{
		Image:      img,
		ColorM:     colorm.ColorM{},
		Options:    &colorm.DrawImageOptions{},
		Horizontal: horizontal,
	}
}

func (img *BGThreePatch) Draw(screen *ebiten.Image, tint color.Color, x, y, w, h float32) {
	if img.Background != nil {
		img.Background.Draw(screen, tint, x, y, w, h)
	}

	cm := img.ColorM
	if tint != nil {
		cm.ScaleWithColor(tint)
	}

	DrawThreepatch(screen, img.Image, x, y, w, h, img.Horizontal, cm, img.Options)
}

type BGColor struct {
	Background ImageDrawer // A reference to any ImageDrawer implementer to render behind this element
	ColorM     colorm.ColorM
	Color      Color
	Antialias  bool
}

func NewBGColor(color Color) *BGColor {
	return &BGColor{
		Color: color,
	}
}

func (img *BGColor) Draw(screen *ebiten.Image, tint color.Color, x, y, w, h float32) {
	if img.Background != nil {
		img.Background.Draw(screen, tint, x, y, w, h)
	}
	cm := img.ColorM
	if tint != nil {
		cm.ScaleWithColor(tint)
	}
	vector.DrawFilledRect(screen, x, y, w, h, cm.Apply(img.Color.ToNRGBA64()), img.Antialias)
}

func NewBGWhite() *BGColor {
	return NewBGColor(NewColor(1, 1, 1, 1))
}

type TextboxOptions struct {
	Background ImageDrawer // A reference to any ImageDrawer implementer to render behind this element

	Text            string // Text to display in the Textbox. Can be modified if Editable is set.
	TextAnchor      AnchorPosition
	TypewriterIndex int
	TypewriterOn    bool
	LineSpacing     float64
	PaddingLeft     float32
	PaddingRight    float32
	PaddingTop      float32
	PaddingBottom   float32
	MaxCharCount    int

	Width  float32
	Height float32

	// Icon       *ebiten.Image
	// IconColor  Color
	// IconAnchor AnchorPosition
	// IconOffset Position

	TextStyle TextStyle

	// When Editable is true, you can click on the Textbox to begin editing and change the textbox's Text string.
	// A known issue is that you can't manually change the text of an editable textbox after creation, so try not to do that.
	Editable          bool
	AllowedCharacters string // Regex string of allowed characters

	CaretDrawFunc func(screen *ebiten.Image, x, y, lineHeight float32) // A function to draw the caret; otherwise, the caret is just a black vertical bar
}

func (t TextboxOptions) WithText(textStr string, args ...any) TextboxOptions {
	if len(args) > 0 {
		t.Text = fmt.Sprintf(textStr, args...)
	} else {
		t.Text = textStr
	}
	return t
}

func (t TextboxOptions) WithPadding(padding float32) TextboxOptions {
	t.PaddingTop = padding
	t.PaddingBottom = padding
	t.PaddingLeft = padding
	t.PaddingRight = padding
	return t
}

func (t TextboxOptions) WithPaddingLeft(padding float32) TextboxOptions {
	t.PaddingLeft = padding
	return t
}

func (t TextboxOptions) WithPaddingRight(padding float32) TextboxOptions {
	t.PaddingRight = padding
	return t
}

func (t TextboxOptions) WithPaddingTop(padding float32) TextboxOptions {
	t.PaddingTop = padding
	return t
}

func (t TextboxOptions) WithPaddingBottom(padding float32) TextboxOptions {
	t.PaddingBottom = padding
	return t
}

func (t TextboxOptions) WithWidth(width float32) TextboxOptions {
	t.Width = width
	return t
}

func (t TextboxOptions) WithHeight(height float32) TextboxOptions {
	t.Height = height
	return t
}

type EditableTextState struct {
	id       any
	Text     []rune
	pastText string
	focused  bool
}

func (e *EditableTextState) ID() any {
	return e.id
}

var parsedText []string

func (a *Area) UITextbox(id any, options TextboxOptions) string {

	ogStyle := textStyle

	if !options.TextStyle.IsZero() {
		SetDefaultTextStyle(options.TextStyle)
	}

	subscreen := a.subscreen()

	x, y, absX, absY := a.layout.Layout(a, id)

	s := internalStateAccessOnce(id)

	if s == nil {
		s = &EditableTextState{id: id, Text: []rune(options.Text)}
		states = append(states, s)
	}

	state := s.(*EditableTextState)

	if !options.Editable && state.pastText != options.Text {
		state.Text = []rune(options.Text)
		state.pastText = options.Text
	}

	/////

	w := options.Width
	h := options.Height

	if w == 0 {
		w = a.Rect.W
	}

	if h == 0 {
		h = a.Rect.H
	}

	parsedText = parsedText[:0]
	textW, textH := text.Measure(string(state.Text), textStyle.Font, textStyle.lineHeight)

	if textW > float64(w-options.PaddingLeft-options.PaddingRight) || strings.ContainsRune(string(state.Text), '\n') {

		for _, s := range strings.Split(string(state.Text), "\n") {

			out := []string{""}
			lineWidth := 0.0

			res := splitWithSeparator(s, " -")
			if len(res) == 1 {
				for _, letter := range res[0] {
					width, _ := text.Measure(string(letter), textStyle.Font, textStyle.lineHeight)
					if lineWidth+width > float64(w)-float64(options.PaddingLeft+options.PaddingRight) {
						out[len(out)-1] = strings.TrimRight(out[len(out)-1], " ")
						out = append(out, "")
						lineWidth = 0
					}
					out[len(out)-1] += string(letter)
					lineWidth += width
				}
			} else {

				for _, word := range res {
					width, _ := text.Measure(word, textStyle.Font, textStyle.lineHeight)
					if lineWidth+width > float64(w)-float64(options.PaddingLeft+options.PaddingRight) {
						out[len(out)-1] = strings.TrimRight(out[len(out)-1], " ")
						out = append(out, "")
						lineWidth = 0
					}
					out[len(out)-1] += word
					lineWidth += width
				}

			}

			parsedText = append(parsedText, out...)

		}

	} else {
		parsedText = append(parsedText, string(state.Text))
	}

	// If height isn't auto-set, set it to the default necessary to fully draw the existing text
	if h == 0 {
		h = float32(textStyle.lineHeight*float64(len(parsedText))) + (options.PaddingTop + options.PaddingBottom)
	}

	a.appendUIRect(id, Rect{
		x, y, w, h,
	})

	x = absX
	y = absY

	if options.Background != nil {
		options.Background.Draw(subscreen, nil, x, y, w, h)
	}

	textDrawOptions := &text.DrawOptions{}
	// textDrawOptions.GeoM.Translate(float64(x)+float64(options.PaddingLeft), float64(y)+float64(options.PaddingTop)-textStyle.Font.Metrics().HDescent)

	tw := float32(textW)
	th := float32(textH)
	descent := float32(textStyle.Font.Metrics().HDescent)

	switch options.TextAnchor {
	case AnchorTopLeft:
		x += options.PaddingLeft
		y += options.PaddingTop
	case AnchorTopCenter:
		x += w/2 - tw/2
		y += options.PaddingTop
	case AnchorTopRight:
		x += w - tw - options.PaddingRight
		y += options.PaddingTop

	case AnchorCenterLeft:
		x += options.PaddingLeft
		y += h/2 - th/2 + (descent / 2)
	case AnchorCenter:
		x += w/2 - tw/2
		y += h/2 - th/2 + (descent / 2)
	case AnchorCenterRight:
		x += w - tw - options.PaddingRight
		y += h/2 - th/2 + (descent / 2)

	case AnchorBottomLeft:
		x += options.PaddingLeft
		y += h - th + (descent / 2) - options.PaddingBottom

	case AnchorBottomCenter:
		x += w/2 - tw/2
		y += h - th + (descent / 2) - options.PaddingBottom

	case AnchorBottomRight:
		x += w - tw - options.PaddingRight
		y += h - th + (descent / 2) - options.PaddingBottom

	default:
		// UNSUPPORTED
	}

	textDrawOptions.GeoM.Translate(float64(x), float64(y)-textStyle.Font.Metrics().HDescent)

	lineSpacing := textStyle.lineHeight
	if options.LineSpacing != 0 {
		lineSpacing = options.LineSpacing
	}

	t := options.TypewriterIndex

	if !options.TypewriterOn || options.Editable {
		t = len(state.Text)
	}

	t = clamp(t, 0, len(state.Text))

	a.drawTextClear()

	for _, p := range parsedText {

		cut := false

		if t > len(p) {
			t -= len(p)
		} else {
			p = p[:t]
			cut = true
		}

		a.drawText(p, textDrawOptions)

		// text.Draw(subscreen, p, textStyle.Font, textDrawOptions)
		textDrawOptions.GeoM.Translate(0, lineSpacing)

		if cut {
			break
		}
	}

	caretY := 0.0
	caretX := 0.0

	if options.Editable && state.focused {

		ci := caretPos

		parsedLineIndex := 0

		for lineIndex, line := range parsedText {
			lineLength := len(line)
			if ci > lineLength {
				ci -= lineLength + 1
				caretY += lineSpacing
				parsedLineIndex = lineIndex + 1
			} else {
				break
			}
		}

		caretX, _ = text.Measure(string(parsedText[parsedLineIndex][:ci]), textStyle.Font, lineSpacing)

	}

	caretX += float64(options.PaddingLeft)
	caretY += float64(options.PaddingTop)

	a.drawTextFlush(textDrawOptions)

	/////

	if options.Editable {

		var clickedArea int

		if !state.focused {
			clickedArea = a.ClickedArea(id, false, Rect{x, y, w, h})
		} else {
			clickedArea = a.ClickedArea(id, true, Rect{x, y, w, h})
		}

		if clickedArea == ClickInArea {
			state.focused = true
			highlightingUIID = id
		} else if clickedArea == ClickOutOfArea {
			state.focused = false
			if highlightingUIID == id {
				highlightingUIID = nil
			}
		}

		mouseX, mouseY := ebiten.CursorPosition()

		if mousePressed() {

			closest := 0
			closestDist := float32(math.MaxFloat32)
			point := Position{}
			runeIndex := 0

			for _, line := range parsedText {

				for _, rune := range line {
					w, _ := text.Measure(string(rune), textStyle.Font, lineSpacing)
					point.X += float32(w)
					if dist := point.DistanceSquaredTo(Position{float32(mouseX), float32(mouseY - int(lineSpacing))}); dist < closestDist {
						closest = runeIndex
						closestDist = dist
					}
					runeIndex++
				}

				w, _ := text.Measure(string(" "), textStyle.Font, lineSpacing)
				point.X += float32(w)
				if dist := point.DistanceSquaredTo(Position{float32(mouseX), float32(mouseY - int(lineSpacing))}); dist < closestDist {
					closest = runeIndex
					closestDist = dist
				}

				point.X = 0
				point.Y += float32(lineSpacing)
				runeIndex++

			}

			caretPos = closest

		}

		if caretPos < 0 {
			caretPos = 0
		}
		if caretPos > len(state.Text) {
			caretPos = len(state.Text)
		}

		if state.focused {

			targetText = &state.Text
			regexString = options.AllowedCharacters

			if options.CaretDrawFunc != nil {
				options.CaretDrawFunc(subscreen, float32(caretX), float32(caretY), float32(textStyle.lineHeight))
			} else {

				// Blink
				if time.Now().UnixMilli()%1000 < 800 {
					vector.DrawFilledRect(subscreen, float32(caretX), float32(caretY), 2, float32(textStyle.lineHeight), color.Black, false)
				}

			}

		}

	} else {
		state.focused = false
	}

	if !options.TextStyle.IsZero() {
		SetDefaultTextStyle(ogStyle)
	}

	// Editable textboxes should be selectable, I think?
	if options.Editable {
		selectableUIIDs = append(selectableUIIDs, id)
	}

	return string(state.Text)

}

type LabelOptions struct {
	Background ImageDrawer // A reference to any ImageDrawer implementer to render behind this element
	Text       string

	Anchor AnchorPosition

	OffsetX float32
	OffsetY float32

	TextPaddingLeft   float32
	TextPaddingRight  float32
	TextPaddingTop    float32
	TextPaddingBottom float32

	TextStyle TextStyle // A custom text style to override the default
}

func (l LabelOptions) WithBackground(bg ImageDrawer) LabelOptions {
	l.Background = bg
	return l
}

func (l LabelOptions) WithOffset(x, y float32) LabelOptions {
	l.OffsetX = x
	l.OffsetY = y
	return l
}

func (l LabelOptions) WithText(textStr string, args ...any) LabelOptions {
	if len(args) > 0 {
		l.Text = fmt.Sprintf(textStr, args...)
	} else {
		l.Text = textStr
	}
	return l
}

func (l LabelOptions) WithTextPadding(padding float32) LabelOptions {
	l.TextPaddingBottom = padding
	l.TextPaddingRight = padding
	l.TextPaddingLeft = padding
	l.TextPaddingTop = padding
	return l
}

func (l LabelOptions) WithTextStyle(style TextStyle) LabelOptions {
	l.TextStyle = style
	return l
}

func (a *Area) UILabel(options LabelOptions) {

	// _, _, _, _, x, y := a.layout.Layout()

	rect := a.Rect
	if len(a.placedElementRects) > 0 {
		rect = a.placedElementRects[len(a.placedElementRects)-1]
	}

	x := rect.X
	y := rect.Y
	w, h := text.Measure(options.Text, textStyle.Font, textStyle.lineHeight)

	ww := float32(w) + options.TextPaddingLeft + options.TextPaddingRight
	hh := float32(h) + options.TextPaddingTop + options.TextPaddingBottom

	switch options.Anchor {
	case AnchorTopCenter:
		x = float32(rect.X) + float32(rect.W)/2 - ww/2
		y = float32(rect.Y)
	case AnchorTopRight:
		x = float32(rect.Right()) - ww
		y = float32(rect.Y)
	case AnchorCenterLeft:
		x = float32(rect.X)
		y = float32(rect.Y) + float32(rect.H)/2 - hh/2
	case AnchorCenter:
		x = float32(rect.X) + float32(rect.W)/2 - ww/2
		y = float32(rect.Y) + float32(rect.H)/2 - hh/2
	case AnchorCenterRight:
		x = float32(rect.Right()) - ww
		y = float32(rect.Y) + float32(rect.H)/2 - hh/2
	case AnchorBottomLeft:
		x = float32(rect.X)
		y = float32(rect.Bottom()) - hh
	case AnchorBottomCenter:
		x = float32(rect.X) + float32(rect.W)/2 - ww/2
		y = float32(rect.Bottom()) - hh
	case AnchorBottomRight:
		x = float32(rect.Right()) - ww
		y = float32(rect.Bottom()) - hh
	}

	x += a.Offset.X + a.parentOffset.X + options.OffsetX
	y += a.Offset.Y + a.parentOffset.Y + options.OffsetY

	opt := &text.DrawOptions{}
	opt.GeoM.Translate(float64(x+options.TextPaddingLeft), float64(y+options.TextPaddingTop))

	ogStyle := textStyle

	if !options.TextStyle.IsZero() {
		SetDefaultTextStyle(options.TextStyle)
	}

	if options.Background != nil {
		options.Background.Draw(a.subscreen(), nil, x, y, ww, hh)
	}

	a.drawTextClear()
	a.drawText(options.Text, opt)
	a.drawTextFlush(opt)

	if !options.TextStyle.IsZero() {
		SetDefaultTextStyle(ogStyle)
	}

}

// UIBackground draws a background image completely behind the Area.
func (a *Area) UIBackground(img ImageDrawer) {
	img.Draw(a.subscreen(), nil, a.Rect.X, a.Rect.Y, a.Rect.W, a.Rect.H)
}

type ImageOptions struct {
	Background       ImageDrawer // A reference to any ImageDrawer implementer to render behind this element
	Image            *ebiten.Image
	DrawImageOptions *ebiten.DrawImageOptions
}

func (a *Area) UIImage(id any, options ImageOptions) {

	sub := a.subscreen()

	x, y, absX, absY := a.layout.Layout(a, id)

	bounds := options.Image.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	img := options.Image

	h := float32(imgH)
	w := float32(imgW)

	// This used to resize the space according to the image's scale; this doesn't seem that useful, so I'm taking it out.

	// scaleW := 1.0
	// scaleH := 1.0

	// if options.DrawImageOptions != nil {

	// 	a := options.DrawImageOptions.GeoM.Element(0, 0)
	// 	b := options.DrawImageOptions.GeoM.Element(0, 1)
	// 	c := options.DrawImageOptions.GeoM.Element(1, 0)
	// 	d := options.DrawImageOptions.GeoM.Element(1, 1)

	// 	scaleW = math.Sqrt((a * a) + c*c)
	// 	scaleH = math.Sqrt((b * b) + d*d)

	// }

	// if a.FlowElementWidth == 0 {

	// 	if newW := float32(imgW) * float32(scaleW); newW > float32(w) {
	// 		w = newW
	// 	}

	// 	if newH := float32(imgH) * float32(scaleH); newH > float32(h) {
	// 		h = newH
	// 	}

	// }

	a.appendUIRect(id, Rect{
		x, y, w, h,
	})

	x = absX
	y = absY

	// sx := float32(1)
	// sy := float32(1)

	// switch options.ScaleMode {
	// case ImageScaleModeStretch:
	// 	sx = float32(w / float32(imgW))
	// 	sy = float32(h / float32(imgH))

	// case ImageScaleModeScale:
	// 	// sx = float32(w / float32(imgW))
	// 	// sy = float32(h / float32(imgH))
	// }

	sub = sub.SubImage(image.Rect(int(math.Round(float64(x))), int(math.Round(float64(y))), int(math.Round(float64(x+w))), int(math.Round(float64(y+h))))).(*ebiten.Image)

	x += w/2 - (float32(imgW) / 2)
	y += h/2 - (float32(imgH) / 2)

	// x += w/2 - (float32(imgW) / 2 * float32(scaleW))
	// y += h/2 - (float32(imgH) / 2 * float32(scaleH))

	opt := &ebiten.DrawImageOptions{}

	if options.DrawImageOptions != nil {
		opt = options.DrawImageOptions
	}

	opt.GeoM.Translate(float64(x), float64(y))

	if options.Background != nil {
		options.Background.Draw(sub, nil, x, y, w, h)
	}

	sub.DrawImage(img, opt)

}

type UIState interface {
	ID() any
}

var states = []UIState{}

// State returns the state of UI elements that have a persistent state over frames.
// This is usually input elements (e.g. Buttons, Checkboxes, etc).
func State(id any) UIState {
	for _, s := range states {
		if s.ID() == id {
			return s
		}
	}
	return nil
}

var idsAccessed []any

// internalStateAccessOnce accesses a state associated with an ID. If the state was accessed before in the current frame, then
// this either panics or warns with a log print.
func internalStateAccessOnce(id any) UIState {

	for _, i := range idsAccessed {
		if i == id {
			switch UIIDReusePolicy {
			case UIIDReusePolicyPanic:
				panic(fmt.Sprint("gooey: UI element ID [", id, "] is used multiple times. Each UI element should have a unique ID."))
			case UIIDReusePolicyWarn:
				log.Println("gooey: UI element ID", id, "is used multiple times. Each UI element should have a unique ID.")
			}
		}
	}

	idsAccessed = append(idsAccessed, id)

	return State(id)

}

// ForEachState runs a function for each State that persists in Gooey.
func ForEachState(forEach func(state UIState)) {
	for _, s := range states {
		forEach(s)
	}
}

// type CheckboxState struct {
// 	id         any
// 	wasPressed bool
// 	Checked    bool
// 	Changed    bool
// }

// func (c *CheckboxState) ID() any {
// 	return c.id
// }

// type CheckboxOptions struct {
// 	Background ImageDrawer

// 	Text string

// 	CheckIcon            *ebiten.Image
// 	CheckIconDrawOptions *ebiten.DrawImageOptions

// 	PaddingLeft   float32
// 	PaddingRight  float32
// 	PaddingTop    float32
// 	PaddingBottom float32

// 	BaseColor      Color
// 	ClickColor     Color
// 	HighlightColor Color

// 	TextAnchorPosition AnchorPosition
// 	CheckboxSize       float32
// 	TextStyle          TextStyle
// }

// func (c CheckboxOptions) WithPadding(padding float32) CheckboxOptions {
// 	c.PaddingLeft = padding
// 	c.PaddingRight = padding
// 	c.PaddingTop = padding
// 	c.PaddingBottom = padding
// 	return c
// }

// // UICheckbox renders a checkbox.
// func (a *Area) UICheckbox(id string, options CheckboxOptions) *CheckboxState {

// 	subscreen := a.subscreen()

// 	s := internalStateAccessOnce(id)

// 	if s == nil {
// 		s = &CheckboxState{id: id}
// 		states = append(states, s)
// 	}

// 	state := s.(*CheckboxState)

// 	ogStyle := textStyle

// 	if !options.TextStyle.IsZero() {
// 		SetTextStyle(options.TextStyle)
// 	}

// 	x, y, w, h, absX, absY := a.layout.Layout()

// 	if w == 0 || h == 0 {
// 		w = 16
// 		h = 16
// 	}

// 	w += options.PaddingLeft + options.PaddingRight
// 	h += options.PaddingTop + options.PaddingBottom

// 	x = absX
// 	y = absY

// 	mouseX, mouseY := ebiten.CursorPosition()

// 	checkboxColor := options.BaseColor
// 	clickColor := options.ClickColor
// 	highlightColor := options.HighlightColor

// 	if checkboxColor.IsZero() {
// 		checkboxColor = NewColor(0.8, 0.8, 0.8, 1)
// 	}

// 	if highlightColor.IsZero() {
// 		highlightColor = checkboxColor.AddRGBA(0.2, 0.2, 0.2, 0)
// 	}

// 	if clickColor.IsZero() {
// 		clickColor = checkboxColor.SubRGBA(0.2, 0.2, 0.2, 0)
// 	}

// 	inputSelect := queuedInput == queuedInputSelect

// 	if inputHighlightIDNumber >= 0 && !inputSelect && mousePressed() {
// 		inputHighlightIDNumber = -999
// 		inputHighlightedUIRect = Rect{}
// 	}

// 	mouseInAreaBounds := float32(mouseX) >= a.Rect.X && float32(mouseX) <= a.Rect.X+a.Rect.W && float32(mouseY) >= a.Rect.Y && float32(mouseY) <= a.Rect.Y+a.Rect.H
// 	mouseOverButton := mouseX >= int(x) && mouseX <= int(x+w) && mouseY >= int(y) && mouseY <= int(y+h)

// 	if !a.scrolling && (mouseOverButton && mouseInAreaBounds) || uiIDNumber == inputHighlightIDNumber {

// 		checkboxColor = highlightColor

// 		if mousePressed() || inputSelect {
// 			state.wasPressed = true
// 			checkboxColor = clickColor
// 		}

// 		if state.wasPressed && (inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && !inputSelect) {
// 			state.Checked = !state.Checked
// 			state.Changed = true
// 		}

// 	}

// 	if options.Background != nil {
// 		options.Background.Draw(subscreen, x, y, w, h)
// 	}

// 	// x, y, w, h, absX, absY := a.layout.Layout()

// 	labelW, labelH := text.Measure(options.Text, textStyle.Font, textStyle.lineHeight)

// 	// if w == 0 {
// 	// 	w = float32(labelW)
// 	// }
// 	// if h == 0 {
// 	// 	h = float32(labelH)
// 	// }

// 	// a.placedElementRects = append(a.placedElementRects,
// 	// 	&Rect{
// 	// 		X: x,
// 	// 		Y: y,
// 	// 		W: w,
// 	// 		H: h,
// 	// 	},
// 	// )

// 	// if uiIDNumber == inputHighlightIDNumber {
// 	// 	inputHighlightedUIRect = Rect{x, y, w, h}
// 	// }

// 	// r := w
// 	// if h < w {
// 	// 	r = h
// 	// }

// 	// if options.CheckboxSize > 0 {
// 	// 	w = options.CheckboxSize
// 	// 	h = options.CheckboxSize
// 	// }

// 	// x = absX
// 	// y = absY

// 	// padding := float32(4)

// 	// mouseX, mouseY := ebiten.CursorPosition()

// 	// checkboxColor := options.BaseColor
// 	// clickColor := options.ClickColor
// 	// highlightColor := options.HighlightColor

// 	// if checkboxColor.IsZero() {
// 	// 	checkboxColor = NewColor(0.8, 0.8, 0.8, 1)
// 	// }

// 	// if highlightColor.IsZero() {
// 	// 	highlightColor = checkboxColor.AddRGBA(0.2, 0.2, 0.2, 0)
// 	// }

// 	// if clickColor.IsZero() {
// 	// 	clickColor = checkboxColor.SubRGBA(0.2, 0.2, 0.2, 0)
// 	// }

// 	// inputSelect := queuedInput == queuedInputSelect

// 	// if !inputSelect && mousePressed() {
// 	// 	inputHighlightIDNumber = -999
// 	// }

// 	// mouseInAreaBounds := float32(mouseX) >= a.Rect.X && float32(mouseX) <= a.Rect.X+a.Rect.W && float32(mouseY) >= a.Rect.Y && float32(mouseY) <= a.Rect.Y+a.Rect.H
// 	// mouseOverButton := mouseX >= int(x+padding-r) && mouseX <= int(x+padding+r) && mouseY >= int(y+padding-r) && mouseY <= int(y+padding+r)

// 	// // if inputUIID == inputHighlightedID && !a.scrollRect.IsZero() {
// 	// // 	ScrollTo(y)
// 	// // }

// 	// state.Changed = false

// 	// if (mouseOverButton && mouseInAreaBounds) || uiIDNumber == inputHighlightIDNumber {
// 	// 	checkboxColor = highlightColor

// 	// 	if mousePressed() || inputSelect {
// 	// 		checkboxColor = clickColor
// 	// 	}

// 	// 	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) || inputSelect {
// 	// 		state.Checked = !state.Checked
// 	// 		state.Changed = true
// 	// 	}

// 	// }

// 	textDrawOptions := &text.DrawOptions{}

// 	switch options.TextAnchorPosition {
// 	// case AnchorTopCenter:
// 	// 	x = float32(rect.X) + float32(rect.W)/2 - ww/2
// 	// 	y = float32(rect.Y)
// 	// case AnchorTopRight:
// 	// 	x = float32(rect.Right()) - ww
// 	// 	y = float32(rect.Y)
// 	case AnchorCenterLeft:
// 		x += float32(labelW)
// 		y += h/2 - float32(labelH/2)
// 		// case AnchorCenter:
// 		// 	x = float32(rect.X) + float32(rect.W)/2 - ww/2
// 		// 	y = float32(rect.Y) + float32(rect.H)/2 - hh/2
// 		// case AnchorCenterRight:
// 		// 	x = float32(rect.Right()) - ww
// 		// 	y = float32(rect.Y) + float32(rect.H)/2 - hh/2
// 		// case AnchorBottomLeft:
// 		// 	x = float32(rect.X)
// 		// 	y = float32(rect.Bottom()) - hh
// 		// case AnchorBottomCenter:
// 		// 	x = float32(rect.X) + float32(rect.W)/2 - ww/2
// 		// 	y = float32(rect.Bottom()) - hh
// 		// case AnchorBottomRight:
// 		// 	x = float32(rect.Right()) - ww
// 		// 	y = float32(rect.Bottom()) - hh
// 	}

// 	textDrawOptions.GeoM.Translate(float64(x), float64(y))
// 	textDrawOptions.GeoM.Translate(float64(w/2)-(labelW/2), float64(h/2)-(labelH/2))

// 	// if options.Ninepatch != nil {
// 	// 	opt := options.NinepatchDrawOptions
// 	// 	if opt == nil {
// 	// 		opt = &ebiten.DrawImageOptions{}
// 	// 	}
// 	// 	opt.ColorScale.ScaleWithColor(checkboxColor.ToNRGBA64())
// 	// 	DrawNinepatch(subscreen, options.Ninepatch, x, y, w, h, opt)
// 	// } else {
// 	// 	vector.DrawFilledRect(subscreen, x, y, w, h, checkboxColor.ToNRGBA64(), false)
// 	// }

// 	// // vector.StrokeCircle(subscreen, x+r+padding, y+r+padding, r, 4, checkboxColor, false)

// 	// // vector.DrawFilledCircle(subscreen, x+r, y+r, r, color.RGBA{200, 200, 200, 255}, false)
// 	// // vector.DrawFilledCircle(subscreen, x+r, y+r, r, color.RGBA{200, 200, 200, 255}, false)

// 	a.drawTextClear()
// 	a.drawText(options.Text, textDrawOptions)
// 	a.drawTextFlush(textDrawOptions)

// 	if !options.TextStyle.IsZero() {
// 		SetTextStyle(ogStyle)
// 	}

// 	return state

// }

func (a *Area) drawScrollbar(x, y, w, h float32, value float32) float32 {

	x += a.parentOffset.X
	y += a.parentOffset.Y

	blockSize := w

	if w > h {
		blockSize = h
	}

	subscreen := a.parentSubscreen()

	// bounds := currentArea.Bounds()

	mx, my := ebiten.CursorPosition()

	mouseX := float32(mx)
	mouseY := float32(my)

	// mouseInAreaBounds := mouseX >= bounds.X && mouseX <= bounds.X+bounds.W && mouseY >= bounds.Y && mouseY <= bounds.Y+bounds.H
	mouseOverButton := mouseOverArea(int(mouseX), int(mouseY), x, y, w, h)

	mouseX = clamp(mouseX, float32(subscreen.Bounds().Min.X), float32(subscreen.Bounds().Max.X))
	mouseY = clamp(mouseY, float32(subscreen.Bounds().Min.Y), float32(subscreen.Bounds().Max.Y))

	blockColor := color.RGBA{200, 200, 200, 255}

	if mouseOverButton {
		// if mouseInAreaBounds && mouseOverButton {
		blockColor = color.RGBA{255, 255, 255, 255}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			a.scrolling = true
		}
	}

	if a.scrolling {
		blockColor = color.RGBA{255, 255, 255, 255}
		// if a.scrollingVertically() {
		value = (mouseY - (blockSize / 2) - y) / (h - blockSize)
		// } else {
		// value = (mouseX - (blockSize / 2) - x) / (w - blockSize)
		// }
		if !mousePressed() {
			a.scrolling = false
		}
	}

	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	vector.DrawFilledRect(subscreen, x, y, w, h, color.RGBA{40, 40, 40, 255}, false)
	// if a.scrollingVertically() {
	vector.DrawFilledRect(subscreen, x, y+((h-blockSize)*value), blockSize, blockSize, blockColor, false)
	// } else {
	// 	vector.DrawFilledRect(subscreen, x+((w-blockSize)*value), y, blockSize, blockSize, blockColor, false)
	// }

	return value
}

// TextStyle is an object that controls how text is rendered in gooey.
type TextStyle struct {
	Font text.Face // The font face to use for rendering the text. The size is customizeable, but the DPI should be 72.

	TextColor Color // The Foreground color for the text. Defaults to white (1, 1, 1, 1).

	// TODO: Support both shadows and outlines

	ShadowDirectionX float32 // A vector indicating direction of the shadow's heading. Defaults to down-right ( {1, 1}, normalized ).
	ShadowDirectionY float32 // A vector indicating direction of the shadow's heading. Defaults to down-right ( {1, 1}, normalized ).
	ShadowLength     int     // The length of the shadow in pixels. Defaults to 0 (no shadow).
	ShadowColorNear  Color   // The color of the shadow near the letters. Defaults to black (0, 0, 0, 1).
	ShadowColorFar   Color   // The color of the shadow towards the end of the letters. Defaults to black (0, 0, 0, 1).

	OutlineThickness int   // Overall thickness of the outline in pixels. Defaults to 0 (no outline).
	OutlineRounded   bool  // If the outline is rounded or not. Defaults to false (square outlines).
	OutlineColor     Color // Color of the outline. Defaults to black (0, 0, 0, 1).

	lineHeight float64
}

// NewDefaultTextStyle returns a TextStyle set to default values.
func NewDefaultTextStyle() TextStyle {
	return TextStyle{
		Font:      defaultFont,
		TextColor: NewColor(0, 0, 0, 1),

		OutlineColor: NewColor(0, 0, 0, 1),

		ShadowDirectionX: 1,
		ShadowDirectionY: 1,
		ShadowColorNear:  NewColor(0, 0, 0, 1),
		ShadowColorFar:   NewColor(0, 0, 0, 1),
	}
}

// WithFGColor returns a TextStyle with the foreground color modified to be the specified color.
func (s TextStyle) WithFGColor(color Color) TextStyle {
	s.TextColor = color
	return s
}

// WithFontFace returns a TextStyle with the font face set.
func (s TextStyle) WithFontFace(face text.Face) TextStyle {
	s.Font = face
	return s
}

// IsZero returns if a TextStyle is zero (uninitialized).
func (s TextStyle) IsZero() bool {
	return s == TextStyle{}
}

// SetDefaultTextStyle sets the default text style for all UI elements with text, unless they override it.
func SetDefaultTextStyle(style TextStyle) {
	textStyle = style
	if style.Font != nil {
		textStyle.lineHeight = style.Font.Metrics().HAscent + style.Font.Metrics().HDescent
	}
}

// CurrentTextStyle returns the currently active TextStyle.
func CurrentTextStyle() TextStyle {
	return textStyle
}

const (
	ClickNotNear = iota
	ClickHover
	ClickInArea
	ClickOutOfArea
)

func (a *Area) ClickedArea(id any, clickOnly bool, rect Rect) int {

	if !usingMouse {
		return ClickOutOfArea
	}

	mouseX, mouseY := ebiten.CursorPosition()

	bounds := a.subscreen().Bounds()
	mouseInAreaBounds := mouseX >= bounds.Min.X && mouseX <= bounds.Max.X && mouseY >= bounds.Min.Y && mouseY <= bounds.Max.Y
	mouseOverButton := mouseX >= int(rect.X) && mouseX < int(rect.X+rect.W) && mouseY >= int(rect.Y) && mouseY < int(rect.Y+rect.H)

	clicked := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || queuedInput == queuedInputSelect

	if clickOnly {
		clicked = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	}

	if clicked {

		if !a.scrolling && (mouseOverButton && mouseInAreaBounds) || highlightingUIID == id {
			return ClickInArea
		}

		return ClickOutOfArea

	} else if (mouseOverButton && mouseInAreaBounds) || highlightingUIID == id {
		return ClickHover
	}

	return ClickNotNear

}

func mousePressed() bool {
	return usingMouse && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func mouseOverArea(mouseX, mouseY int, x, y, w, h float32) bool {
	return usingMouse && mouseX >= int(x) && mouseX < int(x+w) && mouseY >= int(y) && mouseY < int(y+h)
}
