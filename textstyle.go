package gooey

import "github.com/hajimehoshi/ebiten/v2/text/v2"

var textStyle = NewTextStyle()

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

// NewTextStyle returns a TextStyle set to default values.
func NewTextStyle() TextStyle {
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
	if style.Font == nil {
		style.Font = defaultFont
	}
	textStyle.lineHeight = style.Font.Metrics().HAscent + style.Font.Metrics().HDescent
}

// CurrentTextStyle returns the currently active TextStyle.
func CurrentTextStyle() TextStyle {
	return textStyle
}
