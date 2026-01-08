package gooey

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type StretchMode int

const (
	StretchModeFit              = iota // The element scales the smallest amount to fill the rectangle it draws in while keeping its aspect ratio
	StretchModeStretch                 // The element stretches to fill the entire rectangle it draws in, regardless of aspect ratio
	StretchModeFillHorizontally        // The element stretches to fill the entire rectangle it draws in horizontally, keeping its aspect ratio. If it's too tall, it simply overdraws.
	StretchModeFillVertically          // The element stretches to fill the entire rectangle it draws in vertically, keeping its aspect ratio. If it's too wide, it simply overdraws.
	StretchModeNinepatch               // The element is split into a ninepatch and the sides scale to fit the rectangle it draws in.
	StretchModeThreepatch              // The element is split into a horizontal or vertical threepatch and the sides scale to fit the rectangle it draws in.
)

type UIImage struct {
	Image          *ebiten.Image // The source of the image to use for drawing
	DrawOptions    *ebiten.DrawImageOptions
	LayoutModifier ArrangeFunc
	Stretch        StretchMode
	// ClipToRect     bool
}

func NewUIImage() UIImage {
	return UIImage{}
}

func (i UIImage) WithImage(img *ebiten.Image) UIImage {
	i.Image = img
	return i
}

func (i UIImage) WithDrawOptions(drawOpt *ebiten.DrawImageOptions) UIImage {
	i.DrawOptions = drawOpt
	return i
}

func (i UIImage) WithStretch(stretch StretchMode) UIImage {
	i.Stretch = stretch
	return i
}

func (i UIImage) WithLayoutModifier(layoutModifier ArrangeFunc) UIImage {
	i.LayoutModifier = layoutModifier
	return i
}

func (i UIImage) WithTransform(translation, scale Vector2, rotation float32) UIImage {

	if i.DrawOptions == nil {
		i.DrawOptions = &ebiten.DrawImageOptions{}
	}

	opt := i.DrawOptions
	opt.GeoM.Scale(float64(scale.X), float64(scale.Y))
	opt.GeoM.Rotate(float64(rotation))
	opt.GeoM.Translate(float64(translation.X), float64(translation.Y))

	return i
}

// func (i UIImage) WithClipToRect(clipToRect bool) UIImage {
// 	i.ClipToRect = clipToRect
// 	return i
// }

func (i UIImage) highlightable() bool {
	return false
}

func (i UIImage) draw(dc DrawCall) {

	if i.LayoutModifier != nil {
		dc = i.LayoutModifier(dc)
	}

	var drawOpt ebiten.DrawImageOptions
	if i.DrawOptions != nil {
		drawOpt = *i.DrawOptions
	}

	ogGeom := drawOpt.GeoM
	drawOpt.GeoM.Reset()

	drawOpt.ColorScale.Scale(dc.Color.ToFloat32s())

	srcDx := float64(i.Image.Bounds().Dx())
	srcDy := float64(i.Image.Bounds().Dy())

	if i.Stretch != StretchModeNinepatch && i.Stretch != StretchModeThreepatch {

		drawOpt.GeoM.Translate(
			float64(-srcDx/2),
			float64(-srcDy/2),
		)

		switch i.Stretch {
		case StretchModeStretch:
			drawOpt.GeoM.Scale(float64(dc.Rect.W)/srcDx, float64(dc.Rect.H)/srcDy)
		case StretchModeFit:
			smallestFactor := min(float64(dc.Rect.W)/srcDx, float64(dc.Rect.H)/srcDy)
			drawOpt.GeoM.Scale(smallestFactor, smallestFactor)
		case StretchModeFillHorizontally:
			drawOpt.GeoM.Scale(float64(dc.Rect.W)/srcDx, float64(dc.Rect.W)/srcDx)
		case StretchModeFillVertically:
			drawOpt.GeoM.Scale(float64(dc.Rect.H)/srcDy, float64(dc.Rect.H)/srcDy)
		}

		drawOpt.GeoM.Concat(ogGeom)

		drawOpt.GeoM.Translate(
			float64(dc.Rect.Center().X),
			float64(dc.Rect.Center().Y),
		)

		dest := dc.Instance.layout.subscreen()

		// if i.ClipToRect {
		// 	dest = dest.SubImage(image.Rect(int(dc.Rect.X), int(dc.Rect.Y), int(dc.Rect.X+dc.Rect.W), int(dc.Rect.Y+dc.Rect.H))).(*ebiten.Image)
		// }

		dest.DrawImage(i.Image, &drawOpt)

	} else if i.Stretch == StretchModeNinepatch {

		drawOpt.GeoM.Concat(ogGeom)

		DrawNinepatch(dc.Instance.layout.subscreen(), i.Image, dc.Rect.X, dc.Rect.Y, dc.Rect.W, dc.Rect.H, colorm.ColorM{}, &colorm.DrawImageOptions{
			GeoM:       drawOpt.GeoM,
			ColorScale: drawOpt.ColorScale,
			Blend:      drawOpt.Blend,
			Filter:     drawOpt.Filter,
		})

	} else if i.Stretch == StretchModeThreepatch {

		drawOpt.GeoM.Concat(ogGeom)

		horizontal := dc.Rect.W > dc.Rect.H

		DrawThreepatch(dc.Instance.layout.subscreen(), i.Image, dc.Rect.X, dc.Rect.Y, dc.Rect.W, dc.Rect.H, horizontal, colorm.ColorM{}, &colorm.DrawImageOptions{
			GeoM:       drawOpt.GeoM,
			ColorScale: drawOpt.ColorScale,
			Blend:      drawOpt.Blend,
			Filter:     drawOpt.Filter,
		})

	}

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (i UIImage) AddTo(layout *Layout, id string) {
	layout.add(id, i, layout.newDefaultDrawcall())
}

var verts = []ebiten.Vertex{
	{}, {}, {}, {},
}
var indices = []uint16{0, 1, 2, 1, 2, 3}

type UIImageLooping struct {
	OffsetX        float32
	OffsetY        float32
	ScaleX         float32
	ScaleY         float32
	Rotation       float32
	Image          *ebiten.Image // The source of the image to use for drawing
	DrawOptions    *ebiten.DrawTrianglesOptions
	LayoutModifier ArrangeFunc
}

func (i UIImageLooping) WithImage(img *ebiten.Image) UIImageLooping {
	i.Image = img
	return i
}

func (i UIImageLooping) WithDrawOptions(drawOpt *ebiten.DrawTrianglesOptions) UIImageLooping {
	i.DrawOptions = drawOpt
	return i
}

func (i UIImageLooping) WithLayoutModifier(layoutModifier ArrangeFunc) UIImageLooping {
	i.LayoutModifier = layoutModifier
	return i
}

func (i UIImageLooping) highlightable() bool {
	return false
}

func (i UIImageLooping) draw(dc DrawCall) {

	if i.LayoutModifier != nil {
		dc = i.LayoutModifier(dc)
	}

	var drawOpt ebiten.DrawTrianglesOptions
	if i.DrawOptions != nil {
		drawOpt = *i.DrawOptions
	} else {
		drawOpt = ebiten.DrawTrianglesOptions{}
	}

	drawOpt.Address = ebiten.AddressRepeat

	verts[0].DstX = dc.Rect.X
	verts[0].DstY = dc.Rect.Y

	verts[1].DstX = dc.Rect.X + dc.Rect.W
	verts[1].DstY = dc.Rect.Y

	verts[2].DstX = dc.Rect.X
	verts[2].DstY = dc.Rect.Y + dc.Rect.H

	verts[3].DstX = dc.Rect.X + dc.Rect.W
	verts[3].DstY = dc.Rect.Y + dc.Rect.H

	for index := 0; index < 4; index++ {

		pos := Vector2{
			(verts[index].DstX / i.ScaleX) + i.OffsetX,
			(verts[index].DstY / i.ScaleY) + i.OffsetY,
		}.Rotate(i.Rotation)

		verts[index].SrcX = float32(math.Round(float64(pos.X)))
		verts[index].SrcY = float32(math.Round(float64(pos.Y)))

		verts[index].ColorR = dc.Color.R
		verts[index].ColorG = dc.Color.G
		verts[index].ColorB = dc.Color.B
		verts[index].ColorA = dc.Color.A
	}

	dc.Instance.layout.subscreen().DrawTriangles(verts, indices, i.Image, &drawOpt)

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (i UIImageLooping) AddTo(layout *Layout, id string) {
	layout.add(id, i, layout.newDefaultDrawcall())
}
