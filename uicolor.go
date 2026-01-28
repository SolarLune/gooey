package gooey

import (
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// UIColor presents a flat color to the screen.
type UIColor struct {
	ArrangerModifier ArrangeFunc // A customizeable modifier that alters the location where the UI element is going to render.
	FillColor        Color       // The color used to draw a colored square to the screen.
	OutlineThickness float32     // How thick of an outline to use; <= 0 = no outline drawn.
	OutlineColor     Color       // The color of the outline to draw
}

func NewUIColor() UIColor {
	return UIColor{}
}

func (w UIColor) WithFillColor(color Color) UIColor {
	w.FillColor = color
	return w
}

func (w UIColor) WithFillColorRGBA(r, g, b, a float32) UIColor {
	w.FillColor.R = r
	w.FillColor.G = g
	w.FillColor.B = b
	w.FillColor.A = a
	return w
}

func (w UIColor) WithOutlineColor(color Color) UIColor {
	w.OutlineColor = color
	return w
}

func (w UIColor) WithOutlineColorRGBA(r, g, b, a float32) UIColor {
	w.OutlineColor.R = r
	w.OutlineColor.G = g
	w.OutlineColor.B = b
	w.OutlineColor.A = a
	return w
}

func (w UIColor) WithOutlineThickness(thickness float32) UIColor {
	w.OutlineThickness = thickness
	return w
}

func (f UIColor) highlightable() bool {
	return false
}

func (f UIColor) draw(dc *DrawCall) {

	if f.ArrangerModifier != nil {
		f.ArrangerModifier(dc)
	}

	vector.FillRect(dc.Instance.layout.subscreen(), dc.Rect.X, dc.Rect.Y, dc.Rect.W, dc.Rect.H, f.FillColor.Multiply(dc.Color).ToNRGBA64(), false)

	if f.OutlineThickness > 0 {
		vector.StrokeRect(dc.Instance.layout.subscreen(), dc.Rect.X+f.OutlineThickness, dc.Rect.Y+f.OutlineThickness, dc.Rect.W-f.OutlineThickness-1, dc.Rect.H-f.OutlineThickness-1, f.OutlineThickness, f.OutlineColor.Multiply(dc.Color).ToNRGBA64(), false)
	}

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (f UIColor) AddTo(layout *Layout, id string) {
	layout.add(id, f, layout.newDefaultDrawcall())
}
