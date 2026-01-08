package gooey

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type UISlider struct {
	Background     UIElement
	SliderGraphics UIElement
	LayoutModifier ArrangeFunc

	BaseColor                Color
	HighlightColor           Color
	DisabledColor            Color
	ClickPadding             float32
	SliderHeadLerpPercentage float32

	StepSize float32

	Disabled bool
}

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

func (s UISlider) highlightable() bool {
	return !s.Disabled
}

type SliderState struct {
	Percentage       float32
	visualPercentage float32

	held               bool
	sliderHeadPosition Vector2
	disabled           bool
}

func (s *SliderState) SliderHeadPosition() Vector2 {
	return s.sliderHeadPosition
}

func (s UISlider) draw(dc DrawCall) {

	if dc.Instance.state == nil {
		dc.Instance.state = &SliderState{}
	}

	state := dc.Instance.state.(*SliderState)
	state.disabled = s.Disabled

	if s.LayoutModifier != nil {
		dc = s.LayoutModifier(dc)
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
		// percY := (float32(clickY) - (rect.Y + options.ClickPadding)) / (rect.H - (options.ClickPadding / 2))

		if horizontal {
			state.Percentage = percX
		} else {
			state.Percentage = percY
		}

		state.Percentage = float32(math.Round(float64(state.Percentage)*(1.0/float64(stepSize))) * float64(stepSize))

	}

	state.Percentage = clamp(state.Percentage, 0, 1)

	dc.Color = dc.Color.MultiplyRGBA(baseColor.ToFloat32s())

	if s.Background != nil {
		dc.Instance.layout.add(dc.Instance.id+"__bg", s.Background, dc)
		dc.Instance.layout.Advance(-1)
	}

	if s.SliderGraphics != nil {

		if s.SliderHeadLerpPercentage <= 0 {
			state.visualPercentage = state.Percentage
		} else {
			state.visualPercentage += (state.Percentage - state.visualPercentage) * s.SliderHeadLerpPercentage
		}

		sliderRect := dc.Rect
		sliderRect.W = min(dc.Rect.W, dc.Rect.H)
		sliderRect.H = sliderRect.W

		if horizontal {
			sliderRect.X = dc.Rect.X + (state.visualPercentage * (dc.Rect.W - sliderRect.W))
		} else {
			sliderRect.Y = dc.Rect.Y + (state.visualPercentage * (dc.Rect.H - sliderRect.H))
		}

		state.sliderHeadPosition = sliderRect.Center()

		dc.Rect = sliderRect

		dc.Instance.layout.add(dc.Instance.id+"__sliderobj", s.SliderGraphics, dc)
		dc.Instance.layout.Advance(-1)

	}

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (s UISlider) AddTo(layout *Layout, id string) float32 {
	dc := layout.add(id, s, layout.newDefaultDrawcall())
	return dc.Instance.state.(*SliderState).Percentage
}
