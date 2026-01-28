package gooey

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// UIButton represents a pressable / clickable UI element. You can add graphics to it by specifying its Graphics property.
type UIButton struct {
	BaseColor             Color       // The base color for the button.
	HighlightColor        Color       // The highlight color for the Button. This color is used to draw the Button when the mouse hovers over the button or the button is highlighted using keyboard / gamepad input).
	PressedColor          Color       // The click color for the button. This color is used to draw the button when the mouse is clicking on the button.
	ToggledHighlightColor Color       // The highlight color for the button while toggled. This color is used to draw the button when the mouse is clicking on the button.
	DisabledColor         Color       // The disabled color for the button. Used to draw the button when disabled.
	Toggleable            bool        // When enabled, buttons are toggleable.
	Disabled              bool        // When enabled, buttons cannot be highlighted or pressed / toggled.
	ArrangerModifier      ArrangeFunc // A customizeable modifier that alters the location where the UI element is going to render.

	Graphics UIElement // A UIElement object to use as a graphic.
	Pointer  *bool     // A pointer to a variable to set when the button is pressed or toggled.
}

// NewUIButton creates a new UIButton with sensible default values for colors.
func NewUIButton() UIButton {
	return UIButton{
		BaseColor:             NewColor(0.6, 0.6, 0.6, 1),
		HighlightColor:        NewColor(1, 1, 1, 1),
		DisabledColor:         NewColor(0.2, 0.2, 0.2, 1),
		PressedColor:          NewColor(0.2, 0.2, 0.2, 1),
		ToggledHighlightColor: NewColor(0.4, 0.4, 0.4, 1),
	}
}

func (b UIButton) IsZero() bool {
	return b.Disabled == false &&
		b.BaseColor.IsZero() &&
		b.HighlightColor.IsZero() &&
		b.PressedColor.IsZero() &&
		b.DisabledColor.IsZero() &&
		!b.Toggleable &&
		!b.Disabled &&
		b.ArrangerModifier == nil &&
		b.Graphics == nil &&
		b.Pointer == nil
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

func (b UIButton) WithToggledHighlightColor(color Color) UIButton {
	b.ToggledHighlightColor = color
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

func (b UIButton) WithArrangerModifier(modifier ArrangeFunc) UIButton {
	b.ArrangerModifier = modifier
	return b
}

func (b UIButton) WithGraphics(gfx UIElement) UIButton {
	b.Graphics = gfx
	return b
}

// Sets the text of any and all labels already attached to the Button's graphics to the given string.
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

func (b UIButton) draw(dc *DrawCall) {

	if dc.Instance.state == nil {
		dc.Instance.state = &ButtonState{}
	}
	state := dc.Instance.state.(*ButtonState)

	state.toggleable = b.Toggleable

	if b.ArrangerModifier != nil {
		b.ArrangerModifier(dc)
	}

	mouseX, mouseY := ebiten.CursorPosition()

	hovering := Vector2{float32(mouseX), float32(mouseY)}.Inside(dc.Rect)

	isHighlighted := usingMouse && hovering || dc.isHighlighted

	if dc.isHighlighted || (usingMouse && hovering) {

		if updateSettings.AcceptInput || (updateSettings.UseMouse && hovering && updateSettings.LeftMouseClick) {
			// Initial click
			if !b.Disabled && state.pressedState == 0 {
				state.pressedState = 1
			}
		} else if state.pressedState == 1 {
			// Released
			state.pressedState = 2

		} else if state.pressedState == 2 {
			// Not held
			state.pressedState = 0
		}

	} else {
		state.pressedState = 0
	}

	if b.Toggleable && state.pressedState == 2 {
		state.toggled = !state.toggled
	}

	buttonColor := b.BaseColor

	if b.Disabled {
		buttonColor = b.DisabledColor
	} else {

		if state.toggled {
			if isHighlighted {
				buttonColor = b.ToggledHighlightColor
			} else {
				buttonColor = b.PressedColor
			}
		} else {

			if state.pressedState == 1 {
				buttonColor = b.PressedColor
			} else if isHighlighted {
				buttonColor = b.HighlightColor
			}
		}
	}

	dc.Color = dc.Color.MultiplyRGBA(buttonColor.ToFloat32s())
	if b.Graphics != nil {
		dc.Instance.layout.add(dc.Instance.id+"__gfx", b.Graphics, dc.Clone())
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
	dc := layout.newDefaultDrawcall()
	layout.add(id, b, dc)
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

type UIButtonGroup struct {
	MinimumToggled     int
	MaximumToggled     int
	DisallowUntoggling bool
	BaseButton         UIButton
	ArrangerModifier   ArrangeFunc
	Options            []string
	Pointer            int
}

func NewUIButtonGroup(baseStyle UIButton, options ...string) UIButtonGroup {
	return UIButtonGroup{
		BaseButton: baseStyle.WithToggleable(true),
		Options:    options,
	}
}

func (b UIButtonGroup) WithMinimumToggled(count int) UIButtonGroup {
	b.MinimumToggled = count
	return b
}

func (b UIButtonGroup) WithMaximumToggled(count int) UIButtonGroup {
	b.MaximumToggled = count
	return b
}

func (b UIButtonGroup) WithOptions(options ...string) UIButtonGroup {
	b.Options = options
	return b
}

func (b UIButtonGroup) WithBaseButton(baseButton UIButton) UIButtonGroup {
	b.BaseButton = baseButton
	return b
}

func (b UIButtonGroup) WithArrangerModifier(modifier ArrangeFunc) UIButtonGroup {
	b.ArrangerModifier = modifier
	return b
}

type ButtonGroupState struct {
	selectionMade bool
	selected      []bool
	options       []string
	drawnButtons  []*ButtonState
}

// func (b *ButtonGroupState) Choice() int {
// 	return b.choice
// }

// func (b *ButtonGroupState) ChoiceAsString() string {
// 	return b.choiceString
// }

func (b *ButtonGroupState) Options() []string {
	return append([]string{}, b.options...)
}

func (b *ButtonGroupState) SelectionMade() bool {
	return b.selectionMade
}

func (b *ButtonGroupState) Selected() []bool {
	return b.selected
}

func (b *ButtonGroupState) FirstSelected() int {
	for i, s := range b.selected {
		if s {
			return i
		}
	}
	return -1
}

func (b *ButtonGroupState) SetSelected(index int, selection bool) {

	index = clamp(index, 0, len(b.selected)-1)
	b.selected[index] = selection

	b.updateButtonStates()

}

func (b *ButtonGroupState) SetAllSelected(selected bool) {

	for i := range b.selected {
		b.selected[i] = selected
	}

	b.updateButtonStates()

}

func (b *ButtonGroupState) updateButtonStates() {
	for i, s := range b.drawnButtons {
		s.toggled = b.selected[i]
	}
}

func (b UIButtonGroup) draw(dc *DrawCall) {

	if dc.Instance.state == nil {
		state := &ButtonGroupState{}
		dc.Instance.state = state
		// state.choice = 0
	}
	state := dc.Instance.state.(*ButtonGroupState)

	if len(state.selected) < len(b.Options) {
		state.selected = make([]bool, len(b.Options))
		state.options = b.Options
	}

	if b.ArrangerModifier != nil {
		b.ArrangerModifier(dc)
	}

	// for _, o := range b.Options {
	// 	b.ButtonStyle.WithText(o).AddTo()
	// }

	state.drawnButtons = state.drawnButtons[:0]

	for index, option := range b.Options {
		opt := b.BaseButton.WithText(option).WithToggleable(true)
		newDC := dc.Instance.layout.newDefaultDrawcall()
		dc.Instance.layout.add(dc.Instance.id+"__"+strconv.Itoa(index), opt, newDC)
		state.drawnButtons = append(state.drawnButtons, newDC.Instance.state.(*ButtonState))
	}

	toggleCount := 0
	state.selectionMade = false

	if b.DisallowUntoggling {
		for _, s := range state.drawnButtons {
			if s.pressedState == 2 && !s.toggled {
				state.selectionMade = true
				s.toggled = true
				s.pressedState = 0
			}
		}
	}

	for _, b := range state.drawnButtons {
		if b.toggled {
			toggleCount++
		}
	}

	if b.MinimumToggled > 0 {

		if toggleCount < b.MinimumToggled {
			toToggle := b.MinimumToggled - toggleCount
			for _, button := range state.drawnButtons {
				if !button.toggled {
					button.toggled = true
					toToggle--
				}
				if toToggle <= 0 {
					break
				}
			}
		}

	}

	if b.MaximumToggled > 0 {

		prioritized := -1

		for i, button := range state.drawnButtons {
			if button.pressedState == 2 {
				prioritized = i
				break
			}
		}

		if toggleCount > b.MaximumToggled {
			toToggle := toggleCount - b.MaximumToggled
			for i, button := range state.drawnButtons {
				if i != prioritized && button.toggled {
					button.toggled = false
					toToggle--
				}
				if toToggle <= 0 {
					break
				}
			}

			if toToggle > 0 {
				state.drawnButtons[prioritized].toggled = false
			}
		}

	}

	for i, s := range state.drawnButtons {
		if s.pressedState == 2 {
			state.selectionMade = true
		}
		state.selected[i] = s.toggled
	}

}

func (b UIButtonGroup) highlightable() bool {
	return false
}

func (b UIButtonGroup) AddTo(layout *Layout, id string) *ButtonGroupState {
	dc := layout.newDefaultDrawcall()
	layout.add(id, b, dc)
	return dc.Instance.state.(*ButtonGroupState)
}
