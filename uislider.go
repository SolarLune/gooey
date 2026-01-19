package gooey

import (
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

type UISlider struct {
	Background       UIElement   // A UI element to use for drawing the background of the slider.
	SliderGraphics   UIElement   // A UI element to use for drawing the head of the slider.
	ArrangerModifier ArrangeFunc // A customizeable modifier that alters the location where the UI element is going to render.

	BaseColor                Color   // The color to use for the slider by default.
	HighlightColor           Color   // The color to use for the slider when it is highlighted.
	DisabledColor            Color   // The color to use for the slider when it is disabled.
	SliderHeadLerpPercentage float32 // What percentage to lerp the slider between.
	StepSize                 float32 // How coarse in percentages the slider is. Defaults to 0.1 (10%).

	Disabled bool // If the slider is disabled.

	Pointer *float32 // A pointer to a variable to set for the slider to represent.
}

// Creates a new UISlider with sensible default values.
func NewUISlider() UISlider {
	return UISlider{
		SliderHeadLerpPercentage: 0.1,
	}
}

func (s UISlider) WithBackground(bg UIElement) UISlider {
	s.Background = bg
	return s
}

func (s UISlider) WithSliderObject(obj UIElement) UISlider {
	s.SliderGraphics = obj
	return s
}

func (s UISlider) WithBaseColor(color Color) UISlider {
	s.BaseColor = color
	return s
}

func (s UISlider) WithHighlightColor(color Color) UISlider {
	s.HighlightColor = color
	return s
}

func (s UISlider) WithDisabledColor(color Color) UISlider {
	s.DisabledColor = color
	return s
}

func (s UISlider) WithStepSize(stepSize float32) UISlider {
	s.StepSize = stepSize
	return s
}

func (s UISlider) WithDisabled(disabled bool) UISlider {
	s.Disabled = disabled
	return s
}

func (s UISlider) WithPointer(pointer *float32) UISlider {
	s.Pointer = pointer
	return s
}

func (s UISlider) highlightable() bool {
	return !s.Disabled
}

// Sets the text of any and all labels already attached to the Slider's graphics or background to the given string.
func (s UISlider) WithText(txt string) UISlider {
	setTextForAllLabelsInGraphic(s.SliderGraphics, txt)
	setTextForAllLabelsInGraphic(s.Background, txt)
	return s
}

type SliderState struct {
	Percentage       float32
	visualPercentage float32

	held               bool
	sliderHeadPosition Vector2
	disabled           bool
}

// Returns the percentage of the slider as a string.
// start and end are the starting and ending value of the percentage (so while it's 0 to 1 for the slider itself,
// it ranges from, say, 1 to 1000 in the string).
// precision is the number of spaces the value will have in the returned string.
func (s *SliderState) PercentageAsString(start, end float32, precision int) string {
	v := start + (s.Percentage * (end - start))
	return strconv.FormatFloat(float64(v), 'f', precision, 32)
}

func (s *SliderState) SliderHeadPosition() Vector2 {
	return s.sliderHeadPosition
}

func (s UISlider) draw(dc *DrawCall) {

	if dc.Instance.state == nil {
		state := &SliderState{}
		if s.Pointer != nil {
			state.Percentage = *s.Pointer
			state.visualPercentage = state.Percentage
		}
		dc.Instance.state = state
	}

	state := dc.Instance.state.(*SliderState)
	state.disabled = s.Disabled

	if s.ArrangerModifier != nil {
		s.ArrangerModifier(dc)
	}

	stepSize := s.StepSize
	if stepSize == 0 {
		stepSize = 0.1
	}

	baseColor := s.BaseColor
	if baseColor.IsZero() {
		baseColor = NewColor(0.8, 0.8, 0.8, 1)
	}

	mouseX, mouseY := ebiten.CursorPosition()

	hovering := Vector2{float32(mouseX), float32(mouseY)}.Inside(dc.Rect)

	horizontal := dc.Rect.H <= dc.Rect.W

	if s.Disabled {

		if s.DisabledColor.IsZero() {
			baseColor = baseColor.SubRGBA(0.4, 0.4, 0.4, 0)
		} else {
			baseColor = s.DisabledColor
		}

	} else {

		if highlightedElement == dc.Instance || (usingMouse && ((hovering && !updateSettings.LeftMouseClick) || state.held)) {

			if s.HighlightColor.IsZero() {
				baseColor = baseColor.AddRGBA(0.2, 0.2, 0.2, 1)
			} else {
				baseColor = s.HighlightColor
			}

		}

		if usingMouse {

			if hovering && justClicked {
				state.held = true
			} else if !updateSettings.LeftMouseClick {
				state.held = false
			}

		} else if highlightedElement == dc.Instance {

			if horizontal {

				if queuedInput == queuedInputRight {
					state.Percentage += stepSize
					queuedInput = queuedInputNone
				} else if queuedInput == queuedInputLeft {
					state.Percentage -= stepSize
					queuedInput = queuedInputNone
				}

			} else {

				if queuedInput == queuedInputUp {
					state.Percentage -= stepSize
					queuedInput = queuedInputNone
				} else if queuedInput == queuedInputDown {
					state.Percentage += stepSize
					queuedInput = queuedInputNone
				}

			}

		}

	}

	if state.held {
		clickX, clickY := ebiten.CursorPosition()
		percX := (float32(clickX) - dc.Rect.X) / dc.Rect.W
		percY := (float32(clickY) - dc.Rect.Y) / dc.Rect.H

		if horizontal {
			state.Percentage = percX
		} else {
			state.Percentage = percY
		}

		state.Percentage = float32(math.Round(float64(state.Percentage)*(1.0/float64(stepSize))) * float64(stepSize))

	}

	state.Percentage = clamp(state.Percentage, 0, 1)

	if s.Pointer != nil {
		*s.Pointer = state.Percentage
	}

	dc.Color = dc.Color.MultiplyRGBA(baseColor.ToFloat32s())

	if s.Background != nil {
		dc.Instance.layout.add(dc.Instance.id+"__bg", s.Background, dc.Clone())
		dc.Instance.layout.Advance(-1)
	}

	if s.SliderGraphics != nil {

		newDC := dc.Clone()

		if s.SliderHeadLerpPercentage <= 0 {
			state.visualPercentage = state.Percentage
		} else {
			state.visualPercentage += (state.Percentage - state.visualPercentage) * s.SliderHeadLerpPercentage
		}

		sliderRect := newDC.Rect
		sliderRect.W = min(newDC.Rect.W, newDC.Rect.H)
		sliderRect.H = sliderRect.W

		if horizontal {
			sliderRect.X = newDC.Rect.X + (state.visualPercentage * (newDC.Rect.W - sliderRect.W))
		} else {
			sliderRect.Y = newDC.Rect.Y + (state.visualPercentage * (newDC.Rect.H - sliderRect.H))
		}

		state.sliderHeadPosition = sliderRect.Center()

		newDC.Rect = sliderRect

		newDC.Instance.layout.add(newDC.Instance.id+"__sliderobj", s.SliderGraphics, newDC)
		newDC.Instance.layout.Advance(-1)

	}

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (s UISlider) AddTo(layout *Layout, id string) float32 {
	dc := layout.newDefaultDrawcall()
	layout.add(id, s, dc)
	return dc.Instance.state.(*SliderState).Percentage
}
