package gooey

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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

var visibleLayouts = []*Layout{}
var existingLayouts = []*Layout{}
var IDsInUse = []any{}

var ScrollWheelScrollSpeed = float32(1)

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

}

// Init initializes the screen buffer; this should only need to be called once in an application, or whenever you need to resize the UI.
func Init(w, h int) {

	if screenBuffer != nil {

		if bounds := screenBuffer.Bounds(); bounds.Dx() == w && bounds.Dy() == h {
			return
		}
		screenBuffer.Deallocate()

	}

	screenBuffer = ebiten.NewImage(w, h)
	existingLayouts = existingLayouts[:0]

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

	for _, layout := range visibleLayouts {

		x := layout.Rect.X
		y := layout.Rect.Y

		vector.FillRect(screen, x, y, layout.Rect.W, layout.Rect.H, color.RGBA{0, 0, 0, 100}, false)
		vector.StrokeRect(screen, x, y, layout.Rect.W, layout.Rect.H, 1, color.White, false)

		if drawAreaText {
			drawText(x, y, layout.String())
		}

		// for id, r := range area.prevPlacedElementRects {
		// 	rx := r.X
		// 	ry := r.Y
		// 	vector.StrokeRect(screen, rx, ry, r.W, r.H, 1, color.White, false)
		// 	drawText(rx, ry, fmt.Sprintf("%v", id))
		// }
	}
}

const (
	queuedInputNone   = 0
	queuedInputRight  = 1
	queuedInputLeft   = -1
	queuedInputUp     = 2
	queuedInputDown   = -2
	queuedInputPrev   = 3
	queuedInputNext   = -3
	queuedInputSelect = 4
)

// UpdateSettings is a struct indicating boolean values used to determine how you cycle through UI elements.
type UpdateSettings struct {
	LeftInput   bool // Input to use for moving the input selector
	RightInput  bool
	UpInput     bool
	DownInput   bool
	NextInput   bool
	PrevInput   bool
	AcceptInput bool // Selecting ("clicking") a UI element
	CancelInput bool // Pressing cancel

	UseMouse       bool // Whether or not to use the mouse for selecting and clicking UI elements
	LeftMouseClick bool // The input to use for clicking (for rebinding)

	NoDefaultHighlightOption bool

	// HighlightControlRepeatInitialDelay is how long it takes holding a highlight control input to repeat.
	HighlightControlRepeatInitialDelay time.Duration

	// HighlightControlRepeatDelay is how frequently holding a highlight control input repeats after the initial delay.
	HighlightControlRepeatDelay time.Duration
}

var queuedInput = queuedInputNone
var prevQueuedInput = queuedInputNone
var repeatingMouseClick = false
var justClicked = false
var prevMouseClick = false
var usingMouse = true
var canMoveHighlight = true

var highlightedElement *uiElementInstance
var highlightControlInitialTime time.Time
var highlightControlStartTime time.Time
var updateSettings UpdateSettings

// var repeatTimer time.Time

// var inputChars []rune
// var regexString string
// var caretPos int
var targetText *[]rune

var begun = false

// Begin ends the frame the updates input-related things from gooey.
func Begin(settings UpdateSettings) error {

	if begun {
		return errors.New("error: gooey.Begin() called without a previously corresponding gooey.End() call")
	}

	begun = true

	queuedInput = queuedInputNone

	// Do this here

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

		// if settings.CancelInput && !focusedUIElement {
		// 	if opt, exists := Highlights[highlightingUIID]; exists && opt.OnCancel != nil {
		// 		opt.OnCancel()
		// 	}
		// }

	}

	updateSettings = settings

	if !settings.UseMouse {
		usingMouse = false
	} else {
		repeatingMouseClick = settings.LeftMouseClick
		if repeatingMouseClick {
			usingMouse = true
			highlightedElement = nil
		}
	}

	if queuedInput != queuedInputNone {
		usingMouse = false
	}

	if settings.HighlightControlRepeatDelay == 0 {
		settings.HighlightControlRepeatDelay = time.Second / 8
	}

	if settings.HighlightControlRepeatInitialDelay == 0 {
		settings.HighlightControlRepeatInitialDelay = time.Second / 4
	}

	if queuedInput != queuedInputNone {

		if queuedInput != prevQueuedInput {

			highlightControlStartTime = time.Now()
			highlightControlInitialTime = time.Now()

			prevQueuedInput = queuedInput

		} else {

			if time.Since(highlightControlInitialTime) < settings.HighlightControlRepeatInitialDelay {
				prevQueuedInput = queuedInput
				queuedInput = queuedInputNone
			} else if time.Since(highlightControlStartTime) < settings.HighlightControlRepeatDelay {
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

	justClicked = false

	if usingMouse {

		if settings.LeftMouseClick {

			if repeatingMouseClick != prevMouseClick {

				highlightControlStartTime = time.Now()
				highlightControlInitialTime = time.Now()
				justClicked = true
				prevMouseClick = repeatingMouseClick

			} else {

				if time.Since(highlightControlInitialTime) < settings.HighlightControlRepeatInitialDelay || time.Since(highlightControlStartTime) < settings.HighlightControlRepeatDelay {
					prevMouseClick = repeatingMouseClick
					repeatingMouseClick = false
				} else {
					highlightControlStartTime = time.Now() // Let through one input
				}

			}

		} else {
			repeatingMouseClick = false
			prevMouseClick = false
		}

	}

	// if targetText != nil {

	// 	text := *targetText

	// 	inputChars = inputChars[:0]
	// 	inputChars = ebiten.AppendInputChars(inputChars)

	// 	if len(inputChars) > 0 {

	// 		ta := string(text[:caretPos])
	// 		tb := string(text[caretPos:])
	// 		for _, c := range inputChars {
	// 			regexOK, _ := regexp.MatchString(regexString, string(c))
	// 			if regexString == "" || regexOK {
	// 				*targetText = append(append([]rune(ta), c), []rune(tb)...)
	// 			}
	// 		}
	// 		caretPos += len(inputChars)

	// 	}

	// 	regexOK, _ := regexp.MatchString(regexString, "\n")
	// 	if regexString == "" || regexOK {
	// 		if (keyPressed(ebiten.KeyEnter) || keyPressed(ebiten.KeyKPEnter) || keyPressed(ebiten.KeyNumpadEnter)) && len(text) > 0 {
	// 			ta := string(text[:caretPos])
	// 			tb := string(text[caretPos:])
	// 			*targetText = append(append([]rune(ta), '\n'), []rune(tb)...)
	// 			caretPos++
	// 		}
	// 	}

	// 	if keyPressed(ebiten.KeyBackspace) && len(text) > 0 {
	// 		ta := string(text[:max(0, caretPos-1)])
	// 		tb := string(text[caretPos:])
	// 		caretPos--
	// 		*targetText = append([]rune(ta), []rune(tb)...)
	// 	}

	// 	if keyPressed(ebiten.KeyDelete) && len(text) > 0 {
	// 		ta := string(text[:max(0, caretPos)])
	// 		tb := string(text[min(len(text), caretPos+1):])
	// 		*targetText = append([]rune(ta), []rune(tb)...)

	// 	}

	// 	if keyPressed(ebiten.KeyRight) {
	// 		caretPos++
	// 	}

	// 	if keyPressed(ebiten.KeyLeft) {
	// 		caretPos--
	// 	}

	// 	// TODO: Go up and down
	// 	// if keyPressed(ebiten.KeyUp) {
	// 	// 	caretPos++
	// 	// }

	// 	// if keyPressed(ebiten.KeyDown) {
	// 	// 	caretPos--
	// 	// }

	// 	if caretPos < 0 {
	// 		caretPos = 0
	// 	}

	// 	if caretPos > len(*targetText) {
	// 		caretPos = len(*targetText)
	// 	}

	// }

	// targetText = nil

	// Clear highlighting ID
	// if settings.UseMouse && highlightingUIID != nil && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
	// 	highlightingUIID = nil
	// 	inputHighlightedUIRect = Rect{}
	// }

	// selectable := false
	// for _, id := range selectableUIIDs {
	// 	if highlightingUIID == id {
	// 		selectable = true
	// 		break
	// 	}
	// }

	// if !selectable {
	// 	highlightingUIID = nil
	// }

	// if highlightingUIID == nil && len(selectableUIIDs) > 0 {
	// 	highlightingUIID = selectableUIIDs[0]
	// }

	// focusedUIElement = false

	if highlightedElement != nil {

		for _, layout := range visibleLayouts {

			layout.committedMaxRect = layout.currentMaxRect.MoveVec(layout.Offset.Invert())
			layout.currentMaxRect = Rect{}

			thisLayoutHasHighlightedElement := false

			for _, e := range layout.existingUIElements.Data {
				if e == highlightedElement {
					thisLayoutHasHighlightedElement = true
					break
				}
			}

			if thisLayoutHasHighlightedElement && layout.AutoScrollSpeed > 0 {

				scrollingX := false
				scrollingY := false

				if layout.committedMaxRect.H > layout.Rect.H {
					scrollingY = true

					centerScreenY := layout.Rect.Y + (layout.Rect.H / 2)
					edgeSlop := layout.Rect.H / 3

					// Basically, if the element is small enough, then scroll the screen to put it wholly onscreen
					// with some extra tolerance (i.e. some distance away from the edge)
					downTooFar := highlightedElement.currentRect.Bottom() > centerScreenY+edgeSlop
					upTooFar := highlightedElement.currentRect.Y < centerScreenY-edgeSlop

					// If it's too big, then we just scroll it so its leading edge is onscreen
					if highlightedElement.currentRect.H > edgeSlop {
						downTooFar = highlightedElement.currentRect.Bottom() > layout.Rect.Y+layout.Rect.H
						upTooFar = highlightedElement.currentRect.Y < layout.Rect.Y
					}

					if downTooFar {
						layout.autoScrollCurrentSpeed.Y -= layout.AutoScrollAcceleration
					} else if upTooFar {
						layout.autoScrollCurrentSpeed.Y += layout.AutoScrollAcceleration
					} else {

						if layout.autoScrollCurrentSpeed.Y >= layout.AutoScrollAcceleration {
							layout.autoScrollCurrentSpeed.Y -= layout.AutoScrollAcceleration
						} else if layout.autoScrollCurrentSpeed.Y <= -layout.AutoScrollAcceleration {
							layout.autoScrollCurrentSpeed.Y += layout.AutoScrollAcceleration
						} else {
							layout.autoScrollCurrentSpeed.Y = 0
						}

					}

				}

				if layout.committedMaxRect.W > layout.Rect.W {
					scrollingX = true

					centerScreenX := layout.Rect.X + (layout.Rect.W / 2)
					edgeSlop := layout.Rect.W / 3

					// Basically, if the element is small enough, then scroll the screen to put it wholly onscreen
					// with some extra tolerance (i.e. some distance away from the edge)
					rightTooFar := highlightedElement.currentRect.Right() > centerScreenX+edgeSlop
					leftTooFar := highlightedElement.currentRect.X < centerScreenX-edgeSlop

					// If it's too big, then we just scroll it so its leading edge is onscreen
					if highlightedElement.currentRect.W >= edgeSlop {
						rightTooFar = highlightedElement.currentRect.Right() > layout.Rect.X+layout.Rect.W
						leftTooFar = highlightedElement.currentRect.X < layout.Rect.X
					}

					if rightTooFar {
						layout.autoScrollCurrentSpeed.X -= layout.AutoScrollAcceleration
					} else if leftTooFar {
						layout.autoScrollCurrentSpeed.X += layout.AutoScrollAcceleration
					} else {

						if layout.autoScrollCurrentSpeed.X >= layout.AutoScrollAcceleration {
							layout.autoScrollCurrentSpeed.X -= layout.AutoScrollAcceleration
						} else if layout.autoScrollCurrentSpeed.X <= -layout.AutoScrollAcceleration {
							layout.autoScrollCurrentSpeed.X += layout.AutoScrollAcceleration
						} else {
							layout.autoScrollCurrentSpeed.X = 0
						}

					}

				}

				if scrollingY {

					layout.autoScrollCurrentSpeed.Y = clamp(layout.autoScrollCurrentSpeed.Y, -layout.AutoScrollSpeed, layout.AutoScrollSpeed)

					ogScrollY := layout.Offset.Y
					layout.Offset.Y = clamp(layout.Offset.Y+layout.autoScrollCurrentSpeed.Y, -(layout.committedMaxRect.H - layout.Rect.H), 0)

					// Scroll's the same as clamped; it hit a barrier, stop speed
					if layout.Offset.Y == ogScrollY {
						layout.autoScrollCurrentSpeed.Y = 0
					}
				}

				if scrollingX {

					layout.autoScrollCurrentSpeed.X = clamp(layout.autoScrollCurrentSpeed.X, -layout.AutoScrollSpeed, layout.AutoScrollSpeed)

					ogScrollX := layout.Offset.X
					layout.Offset.X = clamp(layout.Offset.X+layout.autoScrollCurrentSpeed.X, -(layout.committedMaxRect.W - layout.Rect.W), 0)

					// Scroll's the same as clamped; it hit a barrier, stop speed
					if layout.Offset.X == ogScrollX {
						layout.autoScrollCurrentSpeed.X = 0
					}

				}

			}

		}
	}

	// Reset visible layouts at the end of Begin so we have layouts / drawn UI elements to work
	// with for highlight movement
	visibleLayouts = visibleLayouts[:0]
	clear(IDsInUse)

	screenBuffer.Clear()
	prevMouseClick = updateSettings.LeftMouseClick

	return nil
}

func End() {

	begun = false

	if (highlightedElement == nil || !highlightedElement.layout.isVisible() || !highlightedElement.wasDrawn) && (queuedInput != 0 || !updateSettings.NoDefaultHighlightOption) && !usingMouse {
		for _, layout := range visibleLayouts {
			found := false

			if len(layout.CustomHighlightingOrder) > 0 {
				for _, e := range layout.CustomHighlightingOrder {
					element := layout.existingUIElements.Data[e]
					if element != nil && element.drawable.highlightable() && element.wasDrawn {
						highlightedElement = element
						found = true
					}
				}
			}

			if !found {

				layout.existingUIElements.ForEach(func(element *uiElementInstance) bool {

					if element.drawable.highlightable() && element.wasDrawn {
						highlightedElement = element
						found = true
						return false
					}
					return true

				})

			}

			if found {
				break
			}
		}
	} else if highlightedElement != nil && queuedInput != queuedInputSelect {

		currentElementRect := highlightedElement.currentRect

		visibleHighlightableElements := []*uiElementInstance{}

		for _, layout := range visibleLayouts {

			layout.existingUIElements.ForEach(func(element *uiElementInstance) bool {

				if element.drawable.highlightable() && element.wasDrawn {
					visibleHighlightableElements = append(visibleHighlightableElements, element)
				}
				return true

			})

		}

		sortByDistance := func(targetPos Vector2) {

			sort.Slice(visibleHighlightableElements, func(i, j int) bool {
				ir := visibleHighlightableElements[i].currentRect
				jr := visibleHighlightableElements[j].currentRect
				return ir.Center().DistanceTo(targetPos) < jr.Center().DistanceTo(targetPos)
			})

		}

		sortByXY := func() {

			sort.Slice(visibleHighlightableElements, func(i, j int) bool {
				ir := visibleHighlightableElements[i].currentRect
				jr := visibleHighlightableElements[j].currentRect
				return ir.X+(ir.Y*100000) < jr.X+(jr.Y*100000)
				// return ir.Center().X+(ir.Center().Y*100000) < jr.Center().X+(jr.Center().Y*100000)
			})

		}

		layout := highlightedElement.layout

		elementFound := false

		if len(layout.CustomHighlightingOrder) > 0 {

			targetID := ""

			for i, e := range layout.CustomHighlightingOrder {
				if e == highlightedElement.id {
					if queuedInput == queuedInputRight || queuedInput == queuedInputDown || queuedInput == queuedInputNext {
						if i < len(layout.CustomHighlightingOrder)-1 {
							targetID = layout.CustomHighlightingOrder[i+1]
						} else {
							targetID = layout.CustomHighlightingOrder[0]
						}
					} else if queuedInput == queuedInputLeft || queuedInput == queuedInputUp || queuedInput == queuedInputPrev {
						if i > 0 {
							targetID = layout.CustomHighlightingOrder[i-1]
						} else {
							targetID = layout.CustomHighlightingOrder[len(layout.CustomHighlightingOrder)-1]
						}
					}
					break
				}
			}

			if targetID != "" {
				for _, e := range visibleHighlightableElements {
					if e.id == targetID {
						highlightedElement = e
						elementFound = true
						break
					}
				}
			}

		}

		if !elementFound {

			var closest *uiElementInstance

			switch queuedInput {
			case queuedInputRight:
				next := currentElementRect.Center()
				next.X = currentElementRect.Right() + 1
				sortByDistance(next)

				for _, e := range visibleHighlightableElements {
					if e == highlightedElement {
						continue
					}
					if e.currentRect.Center().X > currentElementRect.Right() {
						closest = e
						break
					}
				}

				// If there's nothing to the right, sort by distance to the negative location on the X axis and go for that one
				if closest == nil {

					for _, e := range visibleHighlightableElements {
						if e == highlightedElement {
							continue
						}

						overlap := e.currentRect.overlappingAxisY(currentElementRect)

						if overlap > 0 && (closest == nil || e.currentRect.X < closest.currentRect.X) {
							closest = e
						}
					}

				}
			case queuedInputLeft:

				next := currentElementRect.Center()
				next.X = currentElementRect.X - 1
				sortByDistance(next)

				for _, e := range visibleHighlightableElements {
					if e == highlightedElement {
						continue
					}
					if e.currentRect.Center().X < currentElementRect.X {
						closest = e
						break
					}
				}

				if closest == nil {

					for _, e := range visibleHighlightableElements {
						if e == highlightedElement {
							continue
						}
						if e.currentRect.overlappingAxisY(currentElementRect) > 0 && (closest == nil || e.currentRect.X > closest.currentRect.X) {
							closest = e
						}
					}
				}

			case queuedInputUp:
				next := currentElementRect.Center()
				next.Y = currentElementRect.Y - 1
				sortByDistance(next)
				for _, e := range visibleHighlightableElements {
					if e == highlightedElement {
						continue
					}
					if e.currentRect.Center().Y < currentElementRect.Y {
						closest = e
						break
					}
				}

				if closest == nil {

					for _, e := range visibleHighlightableElements {
						if e == highlightedElement {
							continue
						}
						if e.currentRect.overlappingAxisX(currentElementRect) > 0 && (closest == nil || e.currentRect.Y > closest.currentRect.Y) {
							closest = e
						}
					}
				}

			case queuedInputDown:
				next := currentElementRect.Center()
				next.Y = currentElementRect.Bottom() + 1
				sortByDistance(next)
				for _, e := range visibleHighlightableElements {
					if e == highlightedElement {
						continue
					}
					if e.currentRect.Center().Y > currentElementRect.Bottom() {
						closest = e
						break
					}
				}

				if closest == nil {

					for _, e := range visibleHighlightableElements {
						if e == highlightedElement {
							continue
						}
						if e.currentRect.overlappingAxisX(currentElementRect) > 0 && (closest == nil || e.currentRect.Y < closest.currentRect.Y) {
							closest = e
						}
					}
				}

			case queuedInputNext:
				sortByXY()
				for index, e := range visibleHighlightableElements {
					if e == highlightedElement {
						if index < len(visibleHighlightableElements)-1 {
							closest = visibleHighlightableElements[index+1]
						} else {
							closest = visibleHighlightableElements[0]
						}
					}
				}

			case queuedInputPrev:
				sortByXY()
				for index, e := range visibleHighlightableElements {
					if e == highlightedElement {
						if index > 0 {
							closest = visibleHighlightableElements[index-1]
						} else {
							closest = visibleHighlightableElements[len(visibleHighlightableElements)-1]
						}
					}
				}
			}

			if closest != nil {
				highlightedElement = closest
			}

		}

	}

}

func HighlightedUIElement() *uiElementInstance {
	return highlightedElement
}

// func keyPressed(key ebiten.Key) bool {

// 	if inpututil.IsKeyJustPressed(key) || (ebiten.IsKeyPressed(key) && time.Since(repeatTimer) > time.Second/4) {

// 		if inpututil.IsKeyJustPressed(key) {
// 			repeatTimer = time.Now()
// 		}

// 		return true

// 	}

// 	return false

// }

// type Direction int

// const (
// 	DirectionUp = iota + 1
// 	DirectionRight
// 	DirectionNext
// )

// const (
// 	DirectionDown = -iota - 1
// 	DirectionLeft
// 	DirectionPrev
// )

// internalStateAccessOnce accesses a state associated with an ID.
// If the state was accessed before in the current frame, then
// this either panics or warns with a log print.
func internalStateAccessOnce(id any) {

	if _, ok := id.(UIElement); ok {
		panic("gooey: id is a UIElement, which is likely not intentional")
	}

	for _, i := range IDsInUse {
		if i == id {
			switch UIIDReusePolicy {
			case UIIDReusePolicyPanic:
				panic(fmt.Sprint("gooey: UI element ID [", id, "] is used multiple times. Each UI element should have a unique ID."))
			case UIIDReusePolicyWarn:
				log.Println("gooey: UI element ID", id, "is used multiple times. Each UI element should have a unique ID.")
			}
		}
	}

	IDsInUse = append(IDsInUse, id)

}
