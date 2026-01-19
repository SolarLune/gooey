package gooey

import (
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// UIColor presents a flat color to the screen.
type UIColor struct {
	ArrangerModifier ArrangeFunc // A customizeable modifier that alters the location where the UI element is going to render.
	Color            Color       // The color used to draw a colored square to the screen.
}

func NewUIColor() UIColor {
	return UIColor{}
}

func (w UIColor) WithColor(color Color) UIColor {
	w.Color = color
	return w
}

func (w UIColor) WithColorRGBA(r, g, b, a float32) UIColor {
	w.Color.R = r
	w.Color.G = g
	w.Color.B = b
	w.Color.A = a
	return w
}

func (f UIColor) highlightable() bool {
	return false
}

func (f UIColor) draw(dc *DrawCall) {

	if f.ArrangerModifier != nil {
		f.ArrangerModifier(dc)
	}

	vector.FillRect(dc.Instance.layout.subscreen(), dc.Rect.X, dc.Rect.Y, dc.Rect.W, dc.Rect.H, f.Color.Multiply(dc.Color).ToNRGBA64(), false)

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (f UIColor) AddTo(layout *Layout, id string) {
	layout.add(id, f, layout.newDefaultDrawcall())
}
