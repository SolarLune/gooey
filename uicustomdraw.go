package gooey

import "github.com/hajimehoshi/ebiten/v2"

// UICustomDraw represents a UI element that uses a custom draw function to draw within the space of a Layout.
type UICustomDraw struct {
	DrawFunc       func(screen *ebiten.Image, dc DrawCall)
	LayoutModifier ArrangeFunc
	Highlightable  bool
}

func NewUICustomDraw(drawFunc func(screen *ebiten.Image, dc DrawCall)) UICustomDraw {
	return UICustomDraw{
		DrawFunc: drawFunc,
	}
}

func (d UICustomDraw) WithLayoutModifier(modifier ArrangeFunc) UICustomDraw {
	d.LayoutModifier = modifier
	return d
}

func (d UICustomDraw) WithHighlightable(highlightable bool) UICustomDraw {
	d.Highlightable = highlightable
	return d
}

func (d UICustomDraw) draw(dc DrawCall) {

	if d.LayoutModifier != nil {
		dc = d.LayoutModifier(dc)
	}

	if d.DrawFunc != nil {
		d.DrawFunc(dc.Instance.Layout().subscreen(), dc)
	}

}

func (d UICustomDraw) highlightable() bool {
	return d.Highlightable
}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (d UICustomDraw) AddTo(layout *Layout, id string) {
	layout.add(id, d, layout.newDefaultDrawcall())
}
