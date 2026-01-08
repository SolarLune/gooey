package gooey

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// UIButton represents a pressable / clickable UI element. You can add graphics to it by specifying its Graphics property.
type UIButton struct {
	BaseColor      Color // The base color for the button.
	HighlightColor Color // The highlight color for the Button. This color is used to draw the Button when the mouse hovers over the button or the button is highlighted using keyboard / gamepad input)
	PressedColor   Color // The click color for the button. This color is used to draw the button when the mouse is clicking on the button
	DisabledColor  Color
	Toggleable     bool // When enabled, buttons are toggleable.
	Disabled       bool
	LayoutModifier ArrangeFunc

	Graphics UIElement
	Pointer  *bool
}

// NewUIButton creates a new UIButton with sensible default values for colors.
func NewUIButton() UIButton {
	return UIButton{
		BaseColor:      NewColor(0.6, 0.6, 0.6, 1),
		HighlightColor: NewColor(1, 1, 1, 1),
		DisabledColor:  NewColor(0.2, 0.2, 0.2, 1),
		PressedColor:   NewColor(0.2, 0.2, 0.2, 1),
	}
}

func (b UIButton) WithBaseColor(color Color) UIButton {
	b.BaseColor = color
	return b
}

func (b UIButton) WithHighlightColor(color Color) UIButton {
	b.HighlightColor = color
	return b
}

func (b UIButton) WithPressedColor(color Color) UIButton {
	b.PressedColor = color
	return b
}

func (b UIButton) WithDisabledColor(color Color) UIButton {
	b.DisabledColor = color
	return b
}

func (b UIButton) WithToggleable(toggleable bool) UIButton {
	b.Toggleable = toggleable
	return b
}

func (b UIButton) WithDisabled(disabled bool) UIButton {
	b.Disabled = disabled
	return b
}

func (b UIButton) WithLayoutModifier(modifier ArrangeFunc) UIButton {
	b.LayoutModifier = modifier
	return b
}

func (b UIButton) WithGraphics(gfx UIElement) UIButton {
	b.Graphics = gfx
	return b
}

// Sets the text of any and all labels attached to the Button to the given string.
func (b UIButton) WithText(txt string) UIButton {

	setTextForAllLabelsInGraphic(b.Graphics, txt)

	return b
}

func (b UIButton) WithPointer(pointer *bool) UIButton {
	b.Pointer = pointer
	return b
}

func (b UIButton) highlightable() bool {
	return !b.Disabled
}

func (b UIButton) draw(dc DrawCall) {

	if dc.Instance.state == nil {
		dc.Instance.state = &ButtonState{}
	}
	state := dc.Instance.state.(*ButtonState)

	state.toggleable = b.Toggleable

	if b.LayoutModifier != nil {
		dc = b.LayoutModifier(dc)
	}

	buttonColor := b.BaseColor

	mouseX, mouseY := ebiten.CursorPosition()

	hovering := Vector2{float32(mouseX), float32(mouseY)}.Inside(dc.Rect)

	if b.Disabled {

		if b.DisabledColor.IsZero() {
			buttonColor = buttonColor.SubRGBA(0.4, 0.4, 0.4, 0)
		} else {
			buttonColor = b.DisabledColor
		}

	} else {

		if dc.isHighlighted || (usingMouse && hovering) {

			buttonColor = b.HighlightColor

			if updateSettings.AcceptInput || (updateSettings.UseMouse && hovering && updateSettings.LeftMouseClick) {
				if b.PressedColor.IsZero() {
					buttonColor = buttonColor.SubRGBA(0.4, 0.4, 0.4, 0)
				} else {
					buttonColor = b.PressedColor
				}
				if state.pressedState == 0 {
					// Initial click
					state.pressedState = 1
				}
				// Held otherwise
			} else if state.pressedState == 1 {
				// Released
				state.pressedState = 2
				if b.PressedColor.IsZero() {
					buttonColor = buttonColor.SubRGBA(0.4, 0.4, 0.4, 0)
				} else {
					buttonColor = b.PressedColor
				}

			} else if state.pressedState == 2 {
				state.pressedState = 0
			}

		} else {
			state.pressedState = 0
		}

	}

	if b.Toggleable {
		if state.pressedState == 2 {
			state.toggled = !state.toggled
		}
		if state.toggled {
			buttonColor = b.PressedColor
		}
	}

	dc.Color = dc.Color.MultiplyRGBA(buttonColor.ToFloat32s())
	if b.Graphics != nil {
		dc.Instance.layout.add(dc.Instance.id+"__gfx", b.Graphics, dc)
		dc.Instance.layout.Advance(-1)
	}

	if b.Pointer != nil {
		if state.toggleable {
			(*b.Pointer) = state.toggled
		} else {
			(*b.Pointer) = state.Pressed()
		}
	}

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
// The function returns whether the button was pressed, or if toggleable, is currently pressed.
func (b UIButton) AddTo(layout *Layout, id string) bool {
	dc := layout.add(id, b, layout.newDefaultDrawcall())
	return dc.Instance.state.(*ButtonState).Pressed()
}

type ButtonState struct {
	pressedState int
	disabled     bool
	toggled      bool
	toggleable   bool
}

func (b *ButtonState) Pressed() bool {
	if b.toggleable {
		return b.toggled
	}
	return b.pressedState == 2
}

func (b *ButtonState) Disabled() bool {
	return b.disabled
}
