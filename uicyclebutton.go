package gooey

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// UICycleButton draws a button that cycles left to right or top to bottom between a set of choices.
type UICycleButton struct {
	Options []string // The choices to cycle between.

	BaseColor Color // The base color of the button - this color is used to draw the button normally.
	// The highlight color for the button - this color is used to draw the Button when the mouse hovers over
	// the button or the button is highlighted using keyboard / gamepad input)
	HighlightColor Color
	DisabledColor  Color // The disabled color of the button - this color is used when the button is disabled.

	GraphicsButtonBaseColor      Color   // The color to draw for the next and previous graphics buttons as a default
	GraphicsButtonHighlightColor Color   // The color to draw for the next and previous graphics buttons when the mouse hovers over
	GraphicsButtonPressedColor   Color   // The color to draw for the next and previous graphics buttons when pressing
	GraphicsButtonDisabledColor  Color   // The color to draw for the next and previous graphics buttons when they're disabled
	ClickZoneSize                float32 // The size of the buttons; if <= 0, it defaults to the minimum of the button's width or height

	ArrangerModifier ArrangeFunc // A customizeable layout-modifying function that alters where the UI element draws.

	Disabled bool // Whether the button is disabled or not; when disabled, it cannot be pressed or highlighted.

	// Whether the button is clickable to the right and left, or top and bottom,
	// and if you press right and left, or up and down, to cycle through its options.
	Vertical bool

	GraphicsBody           UIElement // The UI element used to represent the body of the button.
	GraphicsButtonPrevious UIElement // The UI element used to represent the previous button of the button.
	GraphicsButtonNext     UIElement // The UI element used to represent the next button of the button.
	Pointer                *int      // When set, whatever value the UICycleButton holds will be applied here
}

func NewCycleButton() UICycleButton {
	return UICycleButton{
		BaseColor:      NewColor(0.6, 0.6, 0.6, 1),
		HighlightColor: NewColor(1, 1, 1, 1),
		DisabledColor:  NewColor(0.2, 0.2, 0.2, 1),

		GraphicsButtonBaseColor:      NewColor(0.6, 0.6, 0.6, 1),
		GraphicsButtonHighlightColor: NewColor(1, 1, 1, 1),
		GraphicsButtonPressedColor:   NewColor(0.2, 0.2, 0.2, 1),
	}
}

func (b UICycleButton) WithBaseColor(color Color) UICycleButton {
	b.BaseColor = color
	return b
}

func (b UICycleButton) WithHighlightColor(color Color) UICycleButton {
	b.HighlightColor = color
	return b
}

func (b UICycleButton) WithDisabledColor(color Color) UICycleButton {
	b.DisabledColor = color
	return b
}

func (b UICycleButton) WithGraphicsButtonBaseColor(color Color) UICycleButton {
	b.GraphicsButtonBaseColor = color
	return b
}

func (b UICycleButton) WithGraphicsButtonHighlightColor(color Color) UICycleButton {
	b.GraphicsButtonHighlightColor = color
	return b
}

func (b UICycleButton) WithGraphicsButtonDisabledColor(color Color) UICycleButton {
	b.GraphicsButtonDisabledColor = color
	return b
}

func (b UICycleButton) WithDisabled(disabled bool) UICycleButton {
	b.Disabled = disabled
	return b
}

func (b UICycleButton) WithArrangerModifier(modifier ArrangeFunc) UICycleButton {
	b.ArrangerModifier = modifier
	return b
}

func (b UICycleButton) WithOptions(options ...string) UICycleButton {
	b.Options = options
	return b
}

func (b UICycleButton) WithGraphicsBody(gfx UIElement) UICycleButton {
	b.GraphicsBody = gfx
	return b
}

func (b UICycleButton) WithClickZoneSize(size float32) UICycleButton {
	b.ClickZoneSize = size
	return b
}

func (b UICycleButton) WithGraphicsButtonPrevious(gfx UIElement) UICycleButton {
	b.GraphicsButtonPrevious = gfx
	return b
}

func (b UICycleButton) WithGraphicsButtonNext(gfx UIElement) UICycleButton {
	b.GraphicsButtonNext = gfx
	return b
}

func (b UICycleButton) WithVertical(vertical bool) UICycleButton {
	b.Vertical = vertical
	return b
}

func (b UICycleButton) WithPointer(pointer *int) UICycleButton {
	b.Pointer = pointer
	return b
}

// Sets the text of any and all labels attached to the Button to the given string.
// func (b UICycleButton) WithText(txt string) UICycleButton {

// 	switch el := b.Graphics.(type) {
// 	case UICollection:
// 		el.ForAllElementsRecursive(func(e UIElement) (UIElement, bool) {
// 			if label, ok := e.(UILabel); ok {
// 				label.Text = txt
// 				return label, true
// 			}
// 			return e, true
// 		})
// 	case UILabel:
// 		el.Text = txt
// 	}

// 	return b
// }

func (b UICycleButton) draw(dc *DrawCall) {

	if b.ArrangerModifier != nil {
		b.ArrangerModifier(dc)
	}

	if dc.Instance.state == nil {
		s := &CycleButtonState{}
		dc.Instance.state = s
		if b.Pointer != nil {
			s.selected = *b.Pointer
		}
	}

	state := dc.Instance.state.(*CycleButtonState)

	color := b.BaseColor

	mouseX, mouseY := ebiten.CursorPosition()

	hovering := Vector2{float32(mouseX), float32(mouseY)}.Inside(dc.Rect)

	prevZone := dc.Rect
	prevZone.W = b.ClickZoneSize
	if prevZone.W == 0 {
		prevZone.W = min(dc.Rect.W, dc.Rect.H)
	}
	nextZone := prevZone.SetRight(dc.Rect.Right())

	if b.Vertical {
		nextZone = dc.Rect
		nextZone.H = b.ClickZoneSize
		if nextZone.H == 0 {
			nextZone.H = min(dc.Rect.W, dc.Rect.H)
		}
		prevZone = nextZone.SetBottom(dc.Rect.Bottom())
	}

	if dc.isHighlighted || (usingMouse && hovering) {
		color = b.HighlightColor
	}

	if b.Disabled {
		color = b.DisabledColor
	} else if dc.isHighlighted {

		if !b.Vertical {

			if queuedInput == queuedInputLeft {
				state.selected--
				queuedInput = queuedInputNone
			} else if queuedInput == queuedInputRight {
				state.selected++
				queuedInput = queuedInputNone
			}

		} else {

			if queuedInput == queuedInputUp {
				state.selected--
				queuedInput = queuedInputNone
			} else if queuedInput == queuedInputDown {
				state.selected++
				queuedInput = queuedInputNone
			}
		}

	}

	if state.selected < 0 {
		state.selected = len(b.Options) - 1
	} else if state.selected >= len(b.Options) {
		state.selected = 0
	}

	if b.Pointer != nil {
		(*b.Pointer) = state.selected
	}

	txt := ""
	if len(b.Options) > 0 {
		txt = b.Options[state.selected]
	}

	dc.Color = dc.Color.MultiplyRGBA(color.ToFloat32s())
	if b.GraphicsBody != nil {
		setTextForAllLabelsInGraphic(b.GraphicsBody, txt)
		dc.Instance.layout.add(dc.Instance.id+"__gfx_body", b.GraphicsBody, dc.Clone())
		dc.Instance.layout.Advance(-1)
	}

	if b.GraphicsButtonPrevious != nil {
		setTextForAllLabelsInGraphic(b.GraphicsButtonPrevious, txt)

		newDrawcall := dc.Clone()

		newDrawcall.Rect.W = min(newDrawcall.Rect.W, newDrawcall.Rect.H)
		newDrawcall.Rect.H = newDrawcall.Rect.W

		if b.ClickZoneSize > 0 {

			if !b.Vertical {
				newDrawcall.Rect.W = b.ClickZoneSize
			} else {
				newDrawcall.Rect.H = b.ClickZoneSize
			}

		}

		zoneColor := b.GraphicsButtonBaseColor

		if b.Disabled {
			zoneColor = b.GraphicsButtonDisabledColor
		} else if usingMouse && hovering {

			// Left Zone
			if prevZone.ContainsPoint(Vector2{X: float32(mouseX), Y: float32(mouseY)}) {

				zoneColor = b.GraphicsButtonHighlightColor

				if repeatingMouseClick {
					state.selected--
				}
				if updateSettings.LeftMouseClick {
					zoneColor = b.GraphicsButtonPressedColor
				}
			}

		}

		if dc.isHighlighted {
			zoneColor = b.GraphicsButtonHighlightColor
		}

		newDrawcall.Color = zoneColor

		dc.Instance.layout.add(dc.Instance.id+"__gfx_button_left", b.GraphicsButtonPrevious, newDrawcall)
		dc.Instance.layout.Advance(-1)
	}

	if b.GraphicsButtonNext != nil {
		setTextForAllLabelsInGraphic(b.GraphicsButtonPrevious, txt)

		newDrawcall := dc.Clone()

		newDrawcall.Rect.W = min(newDrawcall.Rect.W, newDrawcall.Rect.H)
		newDrawcall.Rect.H = newDrawcall.Rect.W

		if b.ClickZoneSize > 0 {

			if !b.Vertical {
				newDrawcall.Rect.W = b.ClickZoneSize
			} else {
				newDrawcall.Rect.H = b.ClickZoneSize
			}

		}

		if !b.Vertical {
			newDrawcall.Rect = newDrawcall.Rect.SetRight(dc.Rect.Right())
		} else {
			newDrawcall.Rect = newDrawcall.Rect.SetBottom(dc.Rect.Bottom())
		}

		zoneColor := b.GraphicsButtonBaseColor

		if b.Disabled {
			zoneColor = b.GraphicsButtonDisabledColor
		} else if usingMouse && hovering {

			// Left Zone
			if nextZone.ContainsPoint(Vector2{X: float32(mouseX), Y: float32(mouseY)}) {

				zoneColor = b.GraphicsButtonHighlightColor

				if repeatingMouseClick {
					state.selected++
				}
				if updateSettings.LeftMouseClick {
					zoneColor = b.GraphicsButtonPressedColor
				}
			}

		}

		if dc.isHighlighted {
			zoneColor = b.GraphicsButtonHighlightColor
		}

		newDrawcall.Color = zoneColor

		dc.Instance.layout.add(dc.Instance.id+"__gfx_button_right", b.GraphicsButtonNext, newDrawcall)
		dc.Instance.layout.Advance(-1)
	}

	// state.selected = 0

}

func (b UICycleButton) highlightable() bool {
	return !b.Disabled
}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (b UICycleButton) AddTo(layout *Layout, id string) int {

	dc := layout.newDefaultDrawcall()
	layout.add(id, b, dc)
	return dc.Instance.state.(*CycleButtonState).selected
}

type CycleButtonState struct {
	selected int
}
