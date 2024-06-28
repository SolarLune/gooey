package gooey

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"

	_ "embed"
)

var screenBuffer *ebiten.Image
var defaultFont text.Face = text.NewGoXFace(basicfont.Face7x13)
var currentFontLineHeight = defaultFont.Metrics().HAscent + defaultFont.Metrics().HDescent

var inputHighlightedUIRect Rect
var areasInUse = []*Area{}

var textBuffer *ebiten.Image
var textShader *ebiten.Shader
var textStyle TextStyle

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
	AnchorTextLeft
	AnchorTextRight
	AnchorTextAbove
	AnchorTextBelow
)

//go:embed text.kage
var textKage []byte

func init() {
	shader, err := ebiten.NewShader(textKage)
	if err != nil {
		panic(err)
	}
	textShader = shader

	SetTextStyle(NewDefaultTextStyle())
}

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

func Clear() {
	screenBuffer.Clear()
}

func Texture() *ebiten.Image {
	return screenBuffer
}

func DrawDebug(screen *ebiten.Image) {
	for _, area := range areasInUse {
		vector.DrawFilledRect(screen, area.Rect.X, area.Rect.Y, area.Rect.W, area.Rect.H, color.RGBA{0, 0, 0, 100}, false)
		vector.StrokeRect(screen, area.Rect.X, area.Rect.Y, area.Rect.W, area.Rect.H, 1, color.White, false)
		opt := &text.DrawOptions{}
		opt.GeoM.Translate(float64(area.Rect.X)+4, float64(area.Rect.Y)+4)
		opt.ColorScale.Scale(0, 0, 0, 1)
		text.Draw(screen, area.String(), defaultFont, opt)
		opt.GeoM.Translate(-1, -1)
		opt.ColorScale.Reset()
		text.Draw(screen, area.String(), defaultFont, opt)
	}
}

func Reset() {
	areasInUse = areasInUse[:0]
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

type Position struct {
	X, Y float32
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

type Area struct {
	ID                 string
	Offset             Position
	Rect               Rect
	scrollOffset       Position
	scrollRect         Rect
	placedElementRects []*Rect

	parentOffset Position
	// texture            *ebiten.Image
	scrolling bool
	parent    *Area
	children  []*Area

	FlowDirection FlowDirection
	FlowPadding   float32
	FlowWidth     float32
	FlowHeight    float32
}

func NewArea(id string, x, y, w, h float32) *Area {

	for _, a := range areasInUse {
		if a.ID == id {
			a.placedElementRects = a.placedElementRects[:0]
			a.LayoutReset()
			return a
		}
	}

	a := &Area{
		ID: id,
		Rect: Rect{
			x, y, w, h,
		},
	}

	areasInUse = append(areasInUse, a)

	return a
}

func NewAreaFromImage(id string, screen *ebiten.Image) *Area {
	screenBounds := screen.Bounds()
	area := NewArea(id, 0, 0, float32(screenBounds.Dx()), float32(screenBounds.Dy()))
	return area
}

func (a *Area) String() string {
	return fmt.Sprintf("%s : { %d, %d, %d, %d } : Scroll : %d, %d", a.ID, int(a.Rect.X), int(a.Rect.Y), int(a.Rect.W), int(a.Rect.H), int(a.scrollOffset.X), int(a.scrollOffset.Y))
}

func (a *Area) LayoutRow(elementCount int, padding float32) {
	a.FlowDirection = FlowHorizontal
	a.FlowPadding = padding
	a.FlowWidth = a.Rect.W / float32(elementCount)
	a.FlowHeight = 0
}

func (a *Area) LayoutColumn(elementCount int, padding float32) {
	a.FlowDirection = FlowVertical
	a.FlowPadding = padding
	a.FlowHeight = a.Rect.H / float32(elementCount)
	a.FlowWidth = 0
}

func (a *Area) LayoutFill() {
	a.FlowDirection = FlowNone
	a.FlowPadding = 0
	a.FlowWidth = a.Rect.W
	a.FlowHeight = a.Rect.H
}

func (a *Area) LayoutReset() {
	a.FlowDirection = FlowVertical
	a.FlowPadding = 0
	a.FlowWidth = 0
	a.FlowHeight = 0
}

func (a *Area) LayoutCustom(elementW, elementH float32) {
	a.FlowDirection = FlowNone
	a.FlowPadding = 0
	a.FlowWidth = elementW
	a.FlowHeight = elementH
}

type Grid struct {
	Cells []*Area
}

func (g *Grid) CellByPosition(x, y int) *Area {
	return g.Cells[(y/4)+(x%4)]
}

// TODO: This is an imperfect solution. A true solution would be to do this, but to make it a layout option, I think.
// That way, we could properly navigate using input.
func (a *Area) UIGrid(id string, cellCountX, cellCountY int, xOffset, yOffset, w, h float32) Grid {

	g := Grid{
		Cells: []*Area{},
	}

	cw := w / float32(cellCountX)
	ch := h / float32(cellCountY)

	for y := 0; y < cellCountY; y++ {
		for x := 0; x < cellCountX; x++ {
			g.Cells = append(g.Cells, NewArea(id+strconv.Itoa((y*cellCountY)+x), a.Rect.X+(float32(x)*cw), a.Rect.Y+(float32(y)*ch), cw, ch))
		}
	}

	dx, dy, _, _, _, _ := a.uiPosition()

	a.placedElementRects = append(a.placedElementRects,
		&Rect{
			X: dx + xOffset,
			Y: dy + yOffset,
			W: w,
			H: h,
		},
	)

	return g
}

func (a *Area) UIArea(id string, xOffset, yOffset, w, h float32) *Area {

	dx, dy, _, _, _, _ := a.uiPosition()

	newArea := NewArea(id, dx+xOffset, dy+yOffset, w, h)
	// if len(a.placedElementRects) > 0 {
	// 	prevRect := a.placedElementRects[len(a.placedElementRects)-1]

	// 	px := float32(0)
	// 	py := float32(0)

	// 	switch a.Spacer.FlowDirection {
	// 	case FlowRight:
	// 		px += prevRect.W + a.Spacer.FlowPadding
	// 	case FlowLeft:
	// 		px -= prevRect.W + a.Spacer.FlowPadding
	// 	case FlowUp:
	// 		py -= prevRect.H + a.Spacer.FlowPadding
	// 	case FlowDown:
	// 		py += prevRect.H + a.Spacer.FlowPadding
	// 	}

	// 	newArea.parent = a
	// 	newArea.parentOffset = Position{
	// 		px,
	// 		py,
	// 	}

	// }

	// newArea.parentOffset = Position{
	// 	dx,
	// 	dy,
	// }

	a.placedElementRects = append(a.placedElementRects,
		&newArea.Rect,
	)

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

func (a *Area) textSubscreen() *ebiten.Image {
	if a.parent != nil {
		return a.parent.textSubscreen().SubImage(image.Rect(int(a.Rect.X+a.parentOffset.X), int(a.Rect.Y+a.parentOffset.Y), int(a.Rect.X)+int(a.Rect.W), int(a.Rect.Y)+int(a.Rect.H))).(*ebiten.Image)
	}
	return textBuffer.SubImage(image.Rect(int(a.Rect.X), int(a.Rect.Y), int(a.Rect.X)+int(a.Rect.W), int(a.Rect.Y)+int(a.Rect.H))).(*ebiten.Image)
}

func (a *Area) uiPosition() (x, y, w, h, absX, absY float32) {

	x = a.Rect.X
	y = a.Rect.Y
	w = float32(0)
	h = float32(0)

	if a.FlowDirection != FlowNone {

		if len(a.placedElementRects) > 0 {

			prevRect := a.placedElementRects[len(a.placedElementRects)-1]

			switch a.FlowDirection {
			case FlowHorizontal:
				x = prevRect.X + prevRect.W + a.FlowPadding
			case FlowVertical:
				y = prevRect.Y + prevRect.H + a.FlowPadding
			}

		}

	}

	// switch a.ExpandMode {

	// case ExpandFill:
	// 	w = a.Rect.W
	// 	h = a.Rect.H
	// case ExpandMinimalMainAxis:
	// 	switch a.FlowDirection {
	// 	case FlowRight:
	// 		h = a.Rect.H
	// 	case FlowLeft:
	// 		h = a.Rect.H
	// 	case FlowUp:
	// 		w = a.Rect.W
	// 	case FlowDown:
	// 		w = a.Rect.W
	// 	}
	// }

	switch a.FlowDirection {
	case FlowHorizontal:
		h = a.Rect.H
		if a.FlowWidth > 0 {
			w = a.FlowWidth
		}
	case FlowVertical:
		w = a.Rect.W
		if a.FlowHeight > 0 {
			h = a.FlowHeight
		}
	case FlowNone:
		w = a.FlowWidth
		h = a.FlowHeight
	}

	absX = x
	absY = y

	if !a.scrollRect.IsZero() {
		absX += a.scrollOffset.X
		absY += a.scrollOffset.Y
		if a.scrollingVertically() {
			w -= 8
		} else {
			h -= 8
		}
	}

	absX += a.parentOffset.X + a.Offset.X
	absY += a.parentOffset.Y + a.Offset.Y

	return x, y, w, h, absX, absY
}

func (a *Area) HandleScrolling() {

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
		if a.scrollingVertically() {
			a.scrollOffset.Y = -inputHighlightedUIRect.Center().Y + (a.Rect.H / 2)
		} else {
			a.scrollOffset.X = -inputHighlightedUIRect.Center().X + (a.Rect.W / 2)
		}
	}

	// TODO: Allow scrolling horizontally AND vertically

	if a.scrollingVertically() {
		scrollAmount = a.scrollOffset.Y / -(a.scrollRect.H - a.Rect.H)
		scrollAmount = a.drawScrollbar(a.Rect.X+a.Rect.W-scrollbarWidth, a.Rect.Y, scrollbarWidth, a.Rect.H, scrollAmount)
		a.scrollOffset.Y = -scrollAmount * (a.scrollRect.H - a.Rect.H)
	} else {
		scrollAmount = a.scrollOffset.X / -(a.scrollRect.W - a.Rect.W)
		scrollAmount = a.drawScrollbar(a.Rect.X, a.Rect.Y+a.Rect.H-scrollbarWidth, a.Rect.W, scrollbarWidth, scrollAmount)
		a.scrollOffset.X = -scrollAmount * (a.scrollRect.W - a.Rect.W)
	}

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

func (a *Area) scrollingVertically() bool {
	// return a.scrollRect.H > a.scrollRect.W
	return a.FlowDirection == FlowVertical
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

// func (a *Area) Bounds() Rect {
// 	screenBounds := currentScreen.Bounds()
// 	r := Rect{
// 		X: float32(math.Max(float64(a.Rect.X), float64(screenBounds.Min.X))),
// 		Y: float32(math.Max(float64(a.Rect.Y), float64(screenBounds.Min.Y))),
// 	}
// 	r.W = float32(math.Min(float64(a.Rect.X+a.Rect.W), float64(screenBounds.Max.X))) - r.X
// 	r.H = float32(math.Min(float64(a.Rect.Y+a.Rect.H), float64(screenBounds.Max.Y))) - r.Y
// 	return r
// }

type FlowDirection int

const (
	FlowVertical FlowDirection = iota
	FlowHorizontal
	FlowNone // When the flow is set to this, UI elements won't advance in a direction
)

// func WithAnchor(anchorPosition AnchorPosition, uiFunc func()) {

// 	uiFunc()

// }

// func WithGrid(rowCount, columnCount int, uiFunc func()) {

// 	positions := []Position{}

// 	columnSize := currentArea.W / float32(columnCount)
// 	rowSize := currentArea.H / float32(rowCount)

// 	for y := 0; y < rowCount; y++ {
// 		for x := 0; x < columnCount; x++ {
// 			positions = append(positions, Position{
// 				X: currentArea.X + (float32(x) * columnSize),
// 				Y: currentArea.Y + (float32(y) * rowSize),
// 			})
// 		}
// 	}

// 	ogPositionIterator := currentPositionIterator

// 	currentPositionIterator = &PositionIterator{
// 		Positions: positions,
// 	}

// 	uiFunc()

// 	currentPositionIterator = ogPositionIterator

// }

// func WithPadding(padding float32, uiFunc func()) {

// 	ogPositionIterator := currentPositionIterator

// 	currentPositionIterator = &PositionIterator{
// 		Padding:    padding,
// 		PaddingSet: true,
// 	}

// 	uiFunc()

// 	currentPositionIterator = ogPositionIterator

// }

// const (
// 	SpacingAbsolute = iota
// 	SpacingRelative
// 	SpacingGrid
// 	SpacingRow
// 	SpacingColumn
// )

// func SetArea(x, y, w, h int, uiFunc func()) {

// }

// func SpaceGrid(rowCount int, columnCount int, uiFunc func()) {}

// func SpaceRow() {}

// func SpaceColumn() {}

var inputUIID = 0
var inputHighlightedID = -999
var prevInputHighlightedID = -999
var inputSelect = false

// func GamepadControl(gamepadID ebiten.GamepadID, uiFunc func()) {

// 	lsX := ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickHorizontal))
// 	// lsY := ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickVertical))
// 	deadzone := 0.25

// 	leftPressed := lsX < -deadzone || ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftLeft)
// 	rightPressed := lsX > deadzone || ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftRight)

// 	if time.Since(inputTick) > time.Millisecond*100 {

// 		if leftPressed {
// 			inputTick = time.Now()
// 			if inputHighlightedID == -999 {
// 				inputHighlightedID = 0
// 			} else {
// 				inputHighlightedID--
// 			}
// 		}

// 		if rightPressed {
// 			inputTick = time.Now()
// 			if inputHighlightedID == -999 {
// 				inputHighlightedID = 0
// 			} else {
// 				inputHighlightedID++
// 			}
// 		}

// 	}

// 	inputUIID = 0

// 	uiFunc()

// 	if inputHighlightedID != -999 {

// 		if inputHighlightedID < 0 {
// 			inputHighlightedID = inputUIID
// 		} else if inputHighlightedID >= inputUIID {
// 			inputHighlightedID = 0
// 		}

// 	}

// }

// var keypresses = map[ebiten.Key]int{}

// func keypress(key ebiten.Key) bool {

// 	if _, ok := keypresses[key]; !ok {
// 		keypresses[key] = 0
// 	}

// 	if ebiten.IsKeyPressed(key) {
// 		if keypresses[key] == 1 {
// 			keypresses[key] = 2
// 		} else {
// 			keypresses[key] = 1
// 		}
// 	} else {
// 		if keypresses[key] != 0 && keypresses[key] != 3 {
// 			keypresses[key] = 3
// 		} else {
// 			keypresses[key] = 0
// 		}
// 	}

// 	return keypresses[key] == 3
// }

// type KeyControlOptions struct {
// 	Prev, Next, Accept ebiten.Key
// }

// func KeyControlBegin(options KeyControlOptions) {

// 	inputSelect = false

// 	leftPressed := keypress(options.Prev)
// 	rightPressed := keypress(options.Next)

// 	if leftPressed {
// 		if inputHighlightedID == -999 {
// 			inputHighlightedID = 0
// 		} else {
// 			inputHighlightedID--
// 		}
// 	}

// 	if rightPressed {
// 		if inputHighlightedID == -999 {
// 			inputHighlightedID = 0
// 		} else {
// 			inputHighlightedID++
// 		}
// 	}

// 	if keypress(options.Accept) {
// 		inputSelect = true
// 	}

// 	inputUIID = 0

// }

// func KeyControlEnd() {

// 	if inputHighlightedID != -999 {

// 		if inputHighlightedID < 0 {
// 			inputHighlightedID = inputUIID - 1
// 		} else if inputHighlightedID >= inputUIID {
// 			inputHighlightedID = 0
// 		}

// 	}

// }

const (
	queuedInputNone = iota
	queuedInputRight
	queuedInputLeft
	queuedInputUp
	queuedInputDown
	queuedInputPrev
	queuedInputNext
	queuedInputAccept
)

var queuedInput = 0

type HighlightControlSettings struct {
	LeftInput   bool // Row and Grid
	RightInput  bool
	UpInput     bool // Column and Grid
	DownInput   bool
	NextInput   bool
	PrevInput   bool // Tabbing
	AcceptInput bool
}

func HighlightControlUpdate(settings HighlightControlSettings) {

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
		queuedInput = queuedInputAccept
	}

}

func (a *Area) HighlightControlBegin() {

	directionalPrev := (a.FlowDirection == FlowHorizontal && queuedInput == queuedInputLeft) || (a.FlowDirection == FlowVertical && queuedInput == queuedInputUp)
	directionalNext := (a.FlowDirection == FlowHorizontal && queuedInput == queuedInputRight) || (a.FlowDirection == FlowVertical && queuedInput == queuedInputDown)

	prevInputHighlightedID = inputHighlightedID

	if directionalPrev || queuedInput == queuedInputPrev {
		if inputHighlightedID == -999 {
			inputHighlightedID = 0
			prevInputHighlightedID = 0
		} else {
			inputHighlightedID--
		}
	}

	if directionalNext || queuedInput == queuedInputNext {
		if inputHighlightedID == -999 {
			inputHighlightedID = 0
			prevInputHighlightedID = 0
		} else {
			inputHighlightedID++
		}
	}

	if queuedInput == queuedInputAccept {
		inputSelect = true
	}

	inputUIID = 0

	queuedInput = queuedInputNone

}

func HighlightControlEnd() {

	inputSelect = false

	// If inputHighlightedID == -999, that means nothing is highlighted through input, but rather
	// UI is being selected by mouse input
	if inputHighlightedID != -999 {

		if inputHighlightedID < 0 {
			inputHighlightedID = inputUIID - 1
		} else if inputHighlightedID >= inputUIID {
			inputHighlightedID = 0
		}

	}

}

// HighlightedUIElementRect returns a rect showing the UI element currently highlighted by keyboard / gamepad input.
func HighlightedUIElementRect() Rect {
	if inputHighlightedID >= 0 {
		return inputHighlightedUIRect
	}
	return Rect{}
}

type ClickMode int

const (
	ClickModeRelease = iota
	ClickModeHold
)

// ButtonOptions is a struct
type ButtonOptions struct {
	BaseColor      Color // The base color of the Button
	HighlightColor Color // The highlight color for the Button. This color is used to draw the Button when the mouse hovers over the button or the button is highlighted using keyboard / gamepad input)
	ClickColor     Color // The click color for the button. This color is used to draw the button when the mouse is clicking on the button

	Label               string  // The text label rendered over the button
	PaddingX, PaddingY  float32 // Horizontal and vertical padding (in pixels)
	MinWidth, MinHeight float32 // The minimum width and height for the Button (in pixels)

	Icon *ebiten.Image // The Icon used for the Button
	// IconColor                  Color          // The color used for blending the icon
	IconAnchor                 AnchorPosition // The AnchorPosition used for the Icon
	IconPaddingX, IconPaddingY float32
	// IconOffset Position
	IconDrawOptions *ebiten.DrawImageOptions

	Ninepatch            *ebiten.Image
	NinepatchDrawOptions *ebiten.DrawImageOptions

	BGPattern            *ebiten.Image
	BGPatternDrawOptions *ebiten.DrawImageOptions

	ClickMode        ClickMode
	ClickRepeatDelay time.Duration
}

func (b ButtonOptions) WithLabel(textStr string, args ...any) ButtonOptions {
	if len(args) > 0 {
		b.Label = fmt.Sprintf(textStr, args...)
	} else {
		b.Label = textStr
	}
	return b
}

func (b ButtonOptions) WithIcon(icon *ebiten.Image) ButtonOptions {
	b.Icon = icon
	return b
}

func (b ButtonOptions) WithIconDrawOptions(opt *ebiten.DrawImageOptions) ButtonOptions {
	b.IconDrawOptions = opt
	return b
}

// UIButton draws a button in the Area, utilizing the ButtonOptions to style and shape the Button.
// The id should be a unique identifying variable (string, number, whatever), used to identify and
// keep track of the button's internal state.
func (a *Area) UIButton(id any, options ButtonOptions) bool {

	// TODO: Support manual newlines and automatic newlines for Buttons

	if !StateExists(id) {
		states[id] = &ButtonState{}
	}

	state := State(id).(*ButtonState)

	subscreen := a.subscreen()
	bounds := subscreen.Bounds()

	// bounds := a.texture.Bounds()

	x, y, w, h, absX, absY := a.uiPosition()

	labelW, labelH := text.Measure(options.Label, textStyle.Font, currentFontLineHeight)

	if w == 0 {
		w = float32(labelW)
	}
	if h == 0 {
		h += float32(labelH)
	}

	if options.Icon != nil {

		bounds := options.Icon.Bounds()

		if w < float32(bounds.Dx()) {
			w = float32(bounds.Dx())
		}

		if h < float32(bounds.Dy()) {
			h = float32(bounds.Dy())
		}

	}

	if w < options.PaddingX {
		w += options.PaddingX
	}

	if h < options.PaddingY {
		h += options.PaddingY
	}

	if options.MinWidth > 0 && w < options.MinWidth {
		w = options.MinWidth
	}

	if options.MinHeight > 0 && h < options.MinHeight {
		h = options.MinHeight
	}

	a.placedElementRects = append(a.placedElementRects,
		&Rect{
			X: x,
			Y: y,
			W: w,
			H: h,
		},
	)

	if inputUIID == inputHighlightedID {
		inputHighlightedUIRect = Rect{x, y, w, h}
	}

	x = absX
	y = absY

	mouseX, mouseY := ebiten.CursorPosition()

	baseColor := options.BaseColor
	clickColor := options.ClickColor
	highlightColor := options.HighlightColor

	if baseColor.IsZero() {
		baseColor = NewColor(0.8, 0.8, 0.8, 1)
	}

	if highlightColor.IsZero() {
		highlightColor = baseColor.AddRGBA(0.2, 0.2, 0.2, 0)
	}

	if clickColor.IsZero() {
		clickColor = baseColor.SubRGBA(0.2, 0.2, 0.2, 0)
	}

	buttonColor := baseColor

	clicked := false

	if !inputSelect && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		inputHighlightedID = -999
		inputHighlightedUIRect = Rect{}
	}

	mouseInAreaBounds := mouseX >= bounds.Min.X && mouseX <= bounds.Max.X && mouseY >= bounds.Min.Y && mouseY <= bounds.Max.Y
	mouseOverButton := mouseX >= int(x) && mouseX < int(x+w) && mouseY >= int(y) && mouseY < int(y+h)

	now := time.Now()

	if !a.scrolling && (mouseOverButton && mouseInAreaBounds) || inputUIID == inputHighlightedID {
		buttonColor = highlightColor

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {

			buttonColor = clickColor

			if !state.WasPressed {
				state.WasPressed = true
				state.ClickTime = now
			}

		} else if inputSelect {
			clicked = true
		}

	}

	switch options.ClickMode {

	case ClickModeRelease:

		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {

			if mouseOverButton && state.WasPressed {
				clicked = true
			}
			state.WasPressed = false

		}

	case ClickModeHold:

		if mouseOverButton && state.WasPressed {

			buttonColor = clickColor
			if options.ClickRepeatDelay == 0 || (state.ClickTime == now || time.Since(state.ClickTime) > options.ClickRepeatDelay) {
				clicked = true
			}

		}

		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || !mouseOverButton {
			state.WasPressed = false
			// state.ClickTime = time.Time{} // Zero it out
		}

	}

	// screen := a.texture.SubImage(image.Rect(int(a.Offset.X), int(a.Offset.Y), int(a.Rect.W), int(a.Rect.H))).(*ebiten.Image)

	textDrawOptions := &text.DrawOptions{}
	textDrawOptions.GeoM.Translate(float64(x), float64(y))
	textDrawOptions.GeoM.Translate(float64(w/2)-(labelW/2), float64(h/2)-(labelH/2))

	textDrawOptions.ColorScale.ScaleWithColor(buttonColor.ToNRGBA64())

	if options.Ninepatch != nil {
		opt := options.NinepatchDrawOptions
		if opt == nil {
			opt = &ebiten.DrawImageOptions{}
		}
		opt.ColorScale.ScaleWithColor(buttonColor.ToNRGBA64())
		DrawNinepatch(subscreen, options.Ninepatch, x, y, w, h, opt)
	} else {
		vector.DrawFilledRect(subscreen, x, y, w, h, buttonColor.ToNRGBA64(), false)
	}

	if options.BGPattern != nil {

		// TODO: Replace this with a shader, I guess?
		// Multiplicative BGPattern works with a ninepatch, but wouldn't
		// if something else rendered underneath, which could, theoretically, happen.

		opt := options.BGPatternDrawOptions
		if options.BGPatternDrawOptions == nil {
			opt = &ebiten.DrawImageOptions{}
		}
		opt.GeoM.Translate(float64(x), float64(y))
		opt.ColorScale.Scale(buttonColor.ToFloat32s())
		subscreen.DrawImage(options.BGPattern.SubImage(image.Rect(0, 0, int(w), int(h))).(*ebiten.Image), opt)
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
			// opt.GeoM.Translate()
			opt.GeoM.Translate(float64(x), float64(y))
			opt.GeoM.Translate(float64(w/2)-(labelW/2)-float64(iconW)-float64(options.IconPaddingX), float64(h/2)-(labelH/2))
		case AnchorTextRight:
			// opt.GeoM.Translate()
			opt.GeoM.Translate(float64(x), float64(y))
			opt.GeoM.Translate(float64(w/2)+(labelW/2)+float64(options.IconPaddingX), float64(h/2)-(labelH/2))
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
	a.drawText(options.Label, textDrawOptions)
	a.drawTextFlush(textDrawOptions)

	inputUIID++

	return clicked
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
		"FGColor":          textStyle.FGColor.MultiplyRGBA(textDrawOptions.ColorScale.R(), textDrawOptions.ColorScale.G(), textDrawOptions.ColorScale.B(), textDrawOptions.ColorScale.A()).ToFloat32Slice(),
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

type TextboxOptions struct {
	TextboxColor    Color
	Text            string
	TypewriterIndex int
	TypewriterOn    bool
	LineSpacing     float64
	Padding         float32
	Height          float32

	// Icon       *ebiten.Image
	// IconColor  Color
	// IconAnchor AnchorPosition
	// IconOffset Position

	Ninepatch            *ebiten.Image
	NinepatchDrawOptions *ebiten.DrawImageOptions

	BGPattern            *ebiten.Image
	BGPatternDrawOptions *ebiten.DrawImageOptions
}

func (t TextboxOptions) SetLabel(textStr string, args ...any) TextboxOptions {
	if len(args) > 0 {
		t.Text = fmt.Sprintf(textStr, args...)
	} else {
		t.Text = textStr
	}
	return t
}

var parsedText []string

func (a *Area) UITextbox(options TextboxOptions) {

	subscreen := a.subscreen()

	x, y, w, h, absX, absY := a.uiPosition()

	if w == 0 {
		w = a.Rect.W
	}

	parsedText = parsedText[:0]
	textW, _ := text.Measure(options.Text, textStyle.Font, currentFontLineHeight)

	if textW > float64(w) {

		for _, s := range strings.Split(options.Text, "\n") {

			out := []string{""}
			lineWidth := 0.0

			res := splitWithSeparator(s, " -")
			if len(res) == 1 {
				for _, letter := range res[0] {
					width, _ := text.Measure(string(letter), textStyle.Font, currentFontLineHeight)
					if lineWidth+width > float64(w)-float64(options.Padding*2) {
						out = append(out, "")
						lineWidth = 0
					}
					out[len(out)-1] += string(letter)
					lineWidth += width
				}
			} else {

				for _, word := range res {
					width, _ := text.Measure(word, textStyle.Font, currentFontLineHeight)
					if lineWidth+width > float64(w)-float64(options.Padding*2) {
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
		parsedText = append(parsedText, options.Text)
	}

	// targetHeight := float32(currentFontLineHeight*float64(len(parsedText))) + (options.Padding * 2)

	// If there's a pre-set height to aim for, go for that
	if options.Height > 0 {
		h = options.Height
	}

	// If height isn't informed by the area's flow options, set it to the default necessary to fully draw the existing text
	if h == 0 {
		h = float32(currentFontLineHeight*float64(len(parsedText))) + (options.Padding * 2)
	}

	a.placedElementRects = append(a.placedElementRects,
		&Rect{
			X: x,
			Y: y,
			W: w,
			H: h,
		},
	)

	x = absX
	y = absY

	textDrawOptions := &text.DrawOptions{}
	textDrawOptions.GeoM.Translate(float64(x)+float64(options.Padding), float64(y)+float64(options.Padding))

	textboxColor := options.TextboxColor

	if textboxColor.IsZero() {
		textboxColor = NewColor(1, 1, 1, 1)
	}

	if options.Ninepatch != nil {
		opt := options.NinepatchDrawOptions
		if opt == nil {
			opt = &ebiten.DrawImageOptions{}
		}
		opt.ColorScale.ScaleWithColor(textboxColor.ToNRGBA64())
		DrawNinepatch(subscreen, options.Ninepatch, x, y, w, h, opt)
	} else {
		vector.DrawFilledRect(subscreen, x, y, w, h, textboxColor.ToNRGBA64(), false)
	}

	if options.BGPattern != nil {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(float64(x), float64(y))
		if !textboxColor.IsZero() {
			opt.ColorScale.Scale(textboxColor.ToFloat32s())
		}
		subscreen.DrawImage(options.BGPattern.SubImage(image.Rect(0, 0, int(w), int(h))).(*ebiten.Image), opt)
	}

	lineSpacing := currentFontLineHeight
	if options.LineSpacing != 0 {
		lineSpacing = options.LineSpacing
	}

	t := options.TypewriterIndex

	if !options.TypewriterOn {
		t = len(options.Text)
	}

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

	a.drawTextFlush(textDrawOptions)

	// UI element isn't highlightable, so skip
	if inputUIID == inputHighlightedID {
		skipHighlightID()
	}

	inputUIID++

}

func skipHighlightID() {
	if prevInputHighlightedID > inputHighlightedID {
		inputHighlightedID--
	} else {
		inputHighlightedID++
	}
}

type ImageOptions struct {
	Image            *ebiten.Image
	DrawImageOptions *ebiten.DrawImageOptions
}

func (a *Area) UIImage(options ImageOptions) {

	sub := a.subscreen()

	x, y, w, h, absX, absY := a.uiPosition()

	bounds := options.Image.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	img := options.Image

	if h == 0 {
		h = float32(imgH)
	}
	if w == 0 {
		w = float32(imgW)
	}

	scaleW := 1.0
	scaleH := 1.0

	if options.DrawImageOptions != nil {

		a := options.DrawImageOptions.GeoM.Element(0, 0)
		b := options.DrawImageOptions.GeoM.Element(0, 1)
		c := options.DrawImageOptions.GeoM.Element(1, 0)
		d := options.DrawImageOptions.GeoM.Element(1, 1)

		scaleW = math.Sqrt((a * a) + c*c)
		scaleH = math.Sqrt((b * b) + d*d)

	}

	if a.FlowWidth == 0 {

		if newW := float32(imgW) * float32(scaleW); newW > float32(w) {
			w = newW
		}

		if newH := float32(imgH) * float32(scaleH); newH > float32(h) {
			h = newH
		}

	}

	a.placedElementRects = append(a.placedElementRects,
		&Rect{
			X: x,
			Y: y,
			W: w,
			H: h,
			// W: w * float32(scaleW),
			// H: h * float32(scaleH),
		},
	)

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

	sub.DrawImage(img, opt)

	// UI element isn't highlightable, so skip
	if inputUIID == inputHighlightedID {
		skipHighlightID()
	}

	inputUIID++

}

var states = map[any]any{}

func State(id any) any {
	return states[id]
}

func StateExists(id any) bool {
	_, ok := states[id]
	return ok
}

type ButtonState struct {
	WasPressed bool
	ClickTime  time.Time
}

type CheckboxState struct {
	Checked bool
}

func (a *Area) UICheckbox(id string) bool {

	subscreen := a.subscreen()

	if !StateExists(id) {
		states[id] = &CheckboxState{}
	}

	state := State(id).(*CheckboxState)

	x, y, w, h, _, _ := a.uiPosition()

	a.placedElementRects = append(a.placedElementRects,
		&Rect{
			X: x,
			Y: y,
			W: w,
			H: h,
		},
	)

	r := w
	if h < w {
		r = h
	}

	padding := float32(4)

	mouseX, mouseY := ebiten.CursorPosition()

	checkboxColor := color.RGBA{200, 200, 200, 255}

	if !inputSelect && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		inputHighlightedID = -999
	}

	mouseInAreaBounds := float32(mouseX) >= a.Rect.X && float32(mouseX) <= a.Rect.X+a.Rect.W && float32(mouseY) >= a.Rect.Y && float32(mouseY) <= a.Rect.Y+a.Rect.H
	mouseOverButton := mouseX >= int(x+padding-r) && mouseX <= int(x+padding+r) && mouseY >= int(y+padding-r) && mouseY <= int(y+padding+r)

	// if inputUIID == inputHighlightedID && !a.scrollRect.IsZero() {
	// 	ScrollTo(y)
	// }

	if (mouseOverButton && mouseInAreaBounds) || inputUIID == inputHighlightedID {
		checkboxColor = color.RGBA{255, 255, 255, 255}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || inputSelect {
			checkboxColor = color.RGBA{150, 150, 150, 255}
		}

		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) || inputSelect {
			state.Checked = !state.Checked
		}

	}

	vector.StrokeCircle(subscreen, x+r+padding, y+r+padding, r, 4, checkboxColor, false)

	vector.DrawFilledCircle(subscreen, x+r, y+r, r, color.RGBA{200, 200, 200, 255}, false)
	vector.DrawFilledCircle(subscreen, x+r, y+r, r, color.RGBA{200, 200, 200, 255}, false)

	return state.Checked

}

func (a *Area) drawScrollbar(x, y, w, h float32, value float32) float32 {

	x += a.parentOffset.X
	y += a.parentOffset.Y

	blockSize := w

	if w > h {
		blockSize = h
	}

	subscreen := a.subscreen()

	// bounds := currentArea.Bounds()

	mx, my := ebiten.CursorPosition()

	mouseX := float32(mx)
	mouseY := float32(my)

	// mouseInAreaBounds := mouseX >= bounds.X && mouseX <= bounds.X+bounds.W && mouseY >= bounds.Y && mouseY <= bounds.Y+bounds.H
	mouseOverButton := mouseX >= x && mouseX <= x+w && mouseY >= y && mouseY <= y+h

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
		if a.scrollingVertically() {
			value = (mouseY - (blockSize / 2) - y) / (h - blockSize)
		} else {
			value = (mouseX - (blockSize / 2) - x) / (w - blockSize)
		}
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
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
	if a.scrollingVertically() {
		vector.DrawFilledRect(subscreen, x, y+((h-blockSize)*value), blockSize, blockSize, blockColor, false)
	} else {
		vector.DrawFilledRect(subscreen, x+((w-blockSize)*value), y, blockSize, blockSize, blockColor, false)
	}

	return value
}

// TODO: Support both shadows and outlines

type TextStyle struct {
	Font text.Face // The font face to use for rendering the text. The size is customizeable, but the DPI should be 72.

	FGColor Color // The Foreground color for the text. Defaults to white (1, 1, 1, 1).

	ShadowDirectionX float32 // A vector indicating direction of the shadow's heading. Defaults to down-right ( {1, 1}, normalized ).
	ShadowDirectionY float32 // A vector indicating direction of the shadow's heading. Defaults to down-right ( {1, 1}, normalized ).
	ShadowLength     int     // The length of the shadow in pixels. Defaults to 0 (no shadow).
	ShadowColorNear  Color   // The color of the shadow near the letters. Defaults to black (0, 0, 0, 1).
	ShadowColorFar   Color   // The color of the shadow towards the end of the letters. Defaults to black (0, 0, 0, 1).

	OutlineThickness int   // Overall thickness of the outline in pixels. Defaults to 0 (no outline).
	OutlineRounded   bool  // If the outline is rounded or not. Defaults to false (square outlines).
	OutlineColor     Color // Color of the outline. Defaults to black (0, 0, 0, 1).

}

func NewDefaultTextStyle() TextStyle {
	return TextStyle{
		Font:    defaultFont,
		FGColor: NewColor(1, 1, 1, 1),

		OutlineColor: NewColor(0, 0, 0, 1),

		ShadowDirectionX: 1,
		ShadowDirectionY: 1,
		ShadowColorNear:  NewColor(0, 0, 0, 1),
		ShadowColorFar:   NewColor(0, 0, 0, 1),
	}
}

func SetTextStyle(style TextStyle) {
	textStyle = style
	if style.Font != nil {
		currentFontLineHeight = textStyle.Font.Metrics().HAscent + textStyle.Font.Metrics().HDescent
	}
}
