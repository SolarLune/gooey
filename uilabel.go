package gooey

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Alignment sets where an element (like an icon) should be placed within another (say, a button).
type Alignment int

const (
	AlignmentTopLeft Alignment = iota
	AlignmentTopCenter
	AlignmentTopRight
	AlignmentCenterLeft
	AlignmentCenterCenter
	AlignmentCenterRight
	AlignmentBottomLeft
	AlignmentBottomCenter
	AlignmentBottomRight
)

type UILabel struct {
	Text        string // Text to display in the Textbox.
	Alignment   Alignment
	DrawOptions *text.DrawOptions

	TypewriterIndex int // What index of the text to type out - <= 0 = all characters in Text, >0 = Text[ 0 : TypewriterIndex - 1 ].

	LineSpacing  float32
	MaxCharCount int

	PaddingTop    float32
	PaddingLeft   float32
	PaddingRight  float32
	PaddingBottom float32

	LayoutModifier    ArrangeFunc
	OverrideTextStyle TextStyle

	// When editable is true, you can click on the Textbox to begin editing and change the textbox's Text string.
	// A known issue is that you can't manually change the text of an editable textbox after creation, so try not to do that.
	// editable bool
	// allowedCharacters string // Regex string of allowed characters

	// CaretDrawFunc func(screen *ebiten.Image, x, y, lineHeight float32) // A function to draw the caret; otherwise, the caret is just a black vertical bar
}

func NewUILabel() UILabel {
	return UILabel{}
}

func (l UILabel) WithText(text string) UILabel {
	l.Text = text
	return l
}

func (l UILabel) WithLineSpacing(lineSpacing float32) UILabel {
	l.LineSpacing = lineSpacing
	return l
}

func (l UILabel) WithAlignment(alignment Alignment) UILabel {
	l.Alignment = alignment
	return l
}

func (l UILabel) WithPadding(padding float32) UILabel {
	l.PaddingLeft = padding
	l.PaddingRight = padding
	l.PaddingTop = padding
	l.PaddingBottom = padding
	return l
}

func (l UILabel) WithTextStyle(textStyle TextStyle) UILabel {
	l.OverrideTextStyle = textStyle
	return l
}

func (l UILabel) WithMaxCharCount(maxCharCount int) UILabel {
	l.MaxCharCount = maxCharCount
	return l
}

func (l UILabel) WithTypewriterIndex(typewriterIndex int) UILabel {
	l.TypewriterIndex = typewriterIndex
	return l
}

func (l UILabel) WithLayoutModifier(modifier ArrangeFunc) UILabel {
	l.LayoutModifier = modifier
	return l
}

func (l UILabel) highlightable() bool {
	return false
}

type LabelState struct {
	targetText []rune
}

func (l UILabel) draw(dc DrawCall) {

	if dc.Instance.state == nil {
		dc.Instance.state = &LabelState{}
	}

	state := dc.Instance.state.(*LabelState)

	state.targetText = []rune(l.Text)

	if l.LayoutModifier != nil {
		dc = l.LayoutModifier(dc)
	}

	ogTextStyle := textStyle

	setStyle := l.OverrideTextStyle

	if !setStyle.IsZero() {
		SetDefaultTextStyle(setStyle)
	}

	textStyle.TextColor = textStyle.TextColor.MultiplyRGBA(dc.Color.ToFloat32s())
	textStyle.OutlineColor = textStyle.OutlineColor.MultiplyRGBA(dc.Color.ToFloat32s())

	var opt *text.DrawOptions

	if l.DrawOptions != nil {
		o := *l.DrawOptions
		opt = &o
	} else {
		opt = &text.DrawOptions{}
	}

	ogGeom := opt.GeoM

	parsedText := []string{}

	allTextWidth, _ := text.Measure(l.Text, textStyle.Font, float64(l.LineSpacing))

	if allTextWidth > float64(dc.Rect.W-l.PaddingLeft-l.PaddingRight) || strings.ContainsRune(string(state.targetText), '\n') {

		for _, s := range strings.Split(string(state.targetText), "\n") {

			out := []string{""}
			lineWidth := 0.0

			res := splitWithSeparator(s, " -")
			if len(res) == 1 {
				for _, letter := range res[0] {
					wordWidth, _ := text.Measure(string(letter), textStyle.Font, textStyle.lineHeight)
					if lineWidth+wordWidth > float64(dc.Rect.W)-float64(l.PaddingLeft+l.PaddingRight) {
						out[len(out)-1] = strings.TrimRight(out[len(out)-1], " ")
						out = append(out, "")
						lineWidth = 0
					}
					out[len(out)-1] += string(letter)
					lineWidth += wordWidth
				}
			} else {

				for _, word := range res {
					wordWidth, _ := text.Measure(word, textStyle.Font, textStyle.lineHeight)
					if lineWidth+wordWidth > float64(dc.Rect.W)-float64(l.PaddingLeft+l.PaddingRight) {
						out[len(out)-1] = strings.TrimRight(out[len(out)-1], " ")
						out = append(out, "")
						lineWidth = 0
					}
					out[len(out)-1] += word
					lineWidth += wordWidth
				}

			}

			parsedText = append(parsedText, out...)

		}

	} else {
		parsedText = append(parsedText, string(state.targetText))
	}

	lineSpacing := textStyle.lineHeight
	if l.LineSpacing != 0 {
		lineSpacing = float64(l.LineSpacing)
	}

	t := l.TypewriterIndex

	// if options.TypewriterIndex <= 0 || options.editable {
	if l.TypewriterIndex <= 0 {
		t = len(state.targetText)
	}

	t = clamp(t, 0, len(state.targetText))

	// Height of the previous line + every other line
	totalTextHeight := textStyle.lineHeight + float64((len(parsedText)-1)*int(lineSpacing))

	for lineIndex, line := range parsedText {

		cut := false

		if t > len(line) {
			t -= len(line)
		} else {
			line = line[:t]
			cut = true
		}

		lineWidth, _ := text.Measure(line, textStyle.Font, lineSpacing)

		lx := float64(0)
		ly := float64(0)

		if l.Alignment == AlignmentTopCenter ||
			l.Alignment == AlignmentCenterCenter ||
			l.Alignment == AlignmentBottomCenter {
			lx = float64(dc.Rect.W/2) - (lineWidth / 2)
		} else if l.Alignment == AlignmentTopRight ||
			l.Alignment == AlignmentCenterRight ||
			l.Alignment == AlignmentBottomRight {
			lx = float64(dc.Rect.W) - lineWidth - float64(l.PaddingRight)
		} else {
			lx = float64(l.PaddingLeft)
		}

		if l.Alignment == AlignmentCenterLeft ||
			l.Alignment == AlignmentCenterCenter ||
			l.Alignment == AlignmentCenterRight {
			ly = float64(dc.Rect.H/2) - (totalTextHeight / 2)
		} else if l.Alignment == AlignmentBottomLeft ||
			l.Alignment == AlignmentBottomCenter ||
			l.Alignment == AlignmentBottomRight {
			ly = float64(dc.Rect.H) - totalTextHeight - float64(l.PaddingBottom)
		} else {
			ly = float64(l.PaddingTop)
		}

		opt.GeoM.Reset()

		opt.GeoM.Concat(ogGeom)

		opt.GeoM.Translate(float64(dc.Rect.X)+lx, float64(dc.Rect.Y)+ly+float64(lineIndex*int(lineSpacing)))

		font := textStyle.Font

		if font == nil {
			font = defaultFont
		}

		rounded := float32(0)
		if textStyle.OutlineRounded {
			rounded = 1
		}

		shadX := textStyle.ShadowDirectionX
		shadY := textStyle.ShadowDirectionY

		shadX, shadY = vecNormalized(shadX, shadY)

		uniformMap := map[string]any{
			"OutlineThickness": float32(textStyle.OutlineThickness),
			"OutlineRounded":   rounded,
			"ShadowVector":     [2]float32{shadX, shadY},
			"ShadowLength":     float32(textStyle.ShadowLength),
			"TextColor":        textStyle.TextColor.MultiplyRGBA(opt.ColorScale.R(), opt.ColorScale.G(), opt.ColorScale.B(), opt.ColorScale.A()).ToFloat32Slice(),
			"OutlineColor":     textStyle.OutlineColor.ToFloat32Slice(),
			"ShadowColorNear":  textStyle.ShadowColorNear.ToFloat32Slice(),
			"ShadowColorFar":   textStyle.ShadowColorFar.ToFloat32Slice(),
		}

		drawTextWithShader(dc.Instance.layout.subscreen(), line, textStyle.Font, opt, uniformMap, textShader)

		if cut {
			break
		}

	}

	SetDefaultTextStyle(ogTextStyle)

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (l UILabel) AddTo(layout *Layout, id string) {
	layout.add(id, l, layout.newDefaultDrawcall())
	// return dc.Instance.State.(*LabelState)
}

// func (a *Area) UILabel(id any, options LabelOptions) *LabelState {

// 	internalStateAccessOnce(id)

// 	ogTextStyle := textStyle

// 	setStyle := options.OverrideTextStyle

// 	if states[id] == nil {
// 		states[id] = &LabelState{uiBase: uiBase{id: id}}
// 	}

// 	state := states[id].(*LabelState)

// 	state.draw = func(area *Area, elementIndex int, rect Rect, color Color) {

// 		state.targetText = []rune(options.Text)

// 		if options.LayoutModifier != nil {
// 			rect = options.LayoutModifier(elementIndex, rect)
// 		}

// 		state.area = area
// 		state.rect = rect

// 		if !setStyle.IsZero() {
// 			SetDefaultTextStyle(setStyle)
// 		}

// 		textStyle.TextColor = textStyle.TextColor.MultiplyRGBA(color.ToFloat32s())
// 		textStyle.OutlineColor = textStyle.OutlineColor.MultiplyRGBA(color.ToFloat32s())

// 		var opt *text.DrawOptions

// 		if options.DrawOptions != nil {
// 			o := *options.DrawOptions
// 			opt = &o
// 		} else {
// 			opt = &text.DrawOptions{}
// 		}

// 		ogGeom := opt.GeoM

// 		parsedText := []string{}

// 		allTextWidth, _ := text.Measure(options.Text, textStyle.Font, float64(options.LineSpacing))

// 		if allTextWidth > float64(rect.W-options.PaddingLeft-options.PaddingRight) || strings.ContainsRune(string(state.targetText), '\n') {

// 			for _, s := range strings.Split(string(state.targetText), "\n") {

// 				out := []string{""}
// 				lineWidth := 0.0

// 				res := splitWithSeparator(s, " -")
// 				if len(res) == 1 {
// 					for _, letter := range res[0] {
// 						wordWidth, _ := text.Measure(string(letter), textStyle.Font, textStyle.lineHeight)
// 						if lineWidth+wordWidth > float64(rect.W)-float64(options.PaddingLeft+options.PaddingRight) {
// 							out[len(out)-1] = strings.TrimRight(out[len(out)-1], " ")
// 							out = append(out, "")
// 							lineWidth = 0
// 						}
// 						out[len(out)-1] += string(letter)
// 						lineWidth += wordWidth
// 					}
// 				} else {

// 					for _, word := range res {
// 						wordWidth, _ := text.Measure(word, textStyle.Font, textStyle.lineHeight)
// 						if lineWidth+wordWidth > float64(rect.W)-float64(options.PaddingLeft+options.PaddingRight) {
// 							out[len(out)-1] = strings.TrimRight(out[len(out)-1], " ")
// 							out = append(out, "")
// 							lineWidth = 0
// 						}
// 						out[len(out)-1] += word
// 						lineWidth += wordWidth
// 					}

// 				}

// 				parsedText = append(parsedText, out...)

// 			}

// 		} else {
// 			parsedText = append(parsedText, string(state.targetText))
// 		}

// 		lineSpacing := textStyle.lineHeight
// 		if options.LineSpacing != 0 {
// 			lineSpacing = float64(options.LineSpacing)
// 		}

// 		t := options.TypewriterIndex

// 		// if options.TypewriterIndex <= 0 || options.editable {
// 		if options.TypewriterIndex <= 0 {
// 			t = len(state.targetText)
// 		}

// 		t = clamp(t, 0, len(state.targetText))

// 		// Height of the previous line + every other line
// 		totalTextHeight := textStyle.lineHeight + float64((len(parsedText)-1)*int(lineSpacing))

// 		for lineIndex, line := range parsedText {

// 			cut := false

// 			if t > len(line) {
// 				t -= len(line)
// 			} else {
// 				line = line[:t]
// 				cut = true
// 			}

// 			lineWidth, _ := text.Measure(line, textStyle.Font, lineSpacing)

// 			lx := float64(0)
// 			ly := float64(0)

// 			if options.TextAnchor == AnchorTopCenter ||
// 				options.TextAnchor == AnchorCenter ||
// 				options.TextAnchor == AnchorBottomCenter {
// 				lx = float64(rect.W/2) - (lineWidth / 2)
// 			} else if options.TextAnchor == AnchorTopRight ||
// 				options.TextAnchor == AnchorCenterRight ||
// 				options.TextAnchor == AnchorBottomRight {
// 				lx = float64(rect.W) - lineWidth - float64(options.PaddingRight)
// 			} else {
// 				lx = float64(options.PaddingLeft)
// 			}

// 			if options.TextAnchor == AnchorCenterLeft ||
// 				options.TextAnchor == AnchorCenter ||
// 				options.TextAnchor == AnchorCenterRight {
// 				ly = float64(rect.H/2) - (totalTextHeight / 2)
// 			} else if options.TextAnchor == AnchorBottomLeft ||
// 				options.TextAnchor == AnchorBottomCenter ||
// 				options.TextAnchor == AnchorBottomRight {
// 				ly = float64(rect.H) - totalTextHeight - float64(options.PaddingBottom)
// 			} else {
// 				ly = float64(options.PaddingTop)
// 			}

// 			opt.GeoM.Reset()

// 			opt.GeoM.Concat(ogGeom)

// 			opt.GeoM.Translate(float64(rect.X)+lx, float64(rect.Y)+ly+float64(lineIndex*int(lineSpacing)))

// 			a.drawText(line, opt)

// 			if cut {
// 				break
// 			}

// 		}

// 		SetDefaultTextStyle(ogTextStyle)

// 	}

// 	a.uiDrawables = append(a.uiDrawables, state)

// 	return state
// }
