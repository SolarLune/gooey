package gooey

import "github.com/hajimehoshi/ebiten/v2"

type Rect struct {
	X, Y, W, H float32
}

func (r Rect) IsZero() bool {
	return r.X == 0 && r.Y == 0 && r.W == 0 && r.H == 0
}

func (r Rect) Center() Vector2 {
	return Vector2{
		X: r.X + r.W/2,
		Y: r.Y + r.H/2,
	}
}

func (r Rect) Right() float32 {
	return r.X + r.W
}

func (r Rect) Bottom() float32 {
	return r.Y + r.H
}

func (r Rect) Top() float32 {
	return r.Y
}

func (r Rect) Left() float32 {
	return r.X
}

func (r Rect) SetBottom(y float32) Rect {
	r.Y = y - r.H
	return r
}

func (r Rect) SetRight(x float32) Rect {
	r.X = x - r.W
	return r
}

func (r Rect) ScaleDownTo(y float32) Rect {
	r.H = y - r.Y
	return r
}

func (r Rect) ScaleRightTo(x float32) Rect {
	r.W = x - r.X
	return r
}

func (r Rect) ScaleUpTo(y float32) Rect {
	bottom := r.Y + r.H
	r.Y = y
	r.H = bottom - r.Y
	return r
}

func (r Rect) ScaleLeftTo(x float32) Rect {
	right := r.X + r.W
	r.X = x
	r.W = right - r.X
	return r
}

func (r Rect) Move(dx, dy float32) Rect {
	r.X += dx
	r.Y += dy
	return r
}

func (r Rect) MoveVec(delta Vector2) Rect {
	r.X += delta.X
	r.Y += delta.Y
	return r
}

func (r Rect) Split(percentage float32, verticalSplit bool) (leftTop, rightBottom Rect) {

	if verticalSplit {
		leftTop = r
		leftTop.W /= 2
		rightBottom = leftTop
		rightBottom.X += rightBottom.W
	} else {
		leftTop = r
		leftTop.H /= 2
		rightBottom = leftTop
		rightBottom.Y += rightBottom.H
	}

	return

}

func (r Rect) AlignToRect(other Rect, anchor AnchorPosition, padding float32) Rect {

	right := r.X + r.W
	bottom := r.Y + r.H

	switch anchor {
	case AnchorTopLeft:
		r.X = float32(other.X) + padding
		r.Y = float32(other.Y) + padding
	case AnchorTopCenter:
		r.X = float32(other.X) + float32(other.W)/2 - r.W/2
		r.Y = float32(other.Y) + padding
	case AnchorTopRight:
		r.X = float32(right) - r.W - padding
		r.Y = float32(other.Y) - padding
	case AnchorCenterLeft:
		r.X = float32(other.X) + padding
		r.Y = float32(other.Y) + float32(other.H)/2 - r.H/2
	case AnchorCenter:
		r.X = float32(other.X) + float32(other.W)/2 - r.W/2
		r.Y = float32(other.Y) + float32(other.H)/2 - r.H/2
	case AnchorCenterRight:
		r.X = float32(right) - r.W - padding
		r.Y = float32(other.Y) + float32(other.H)/2 - r.H/2
	case AnchorBottomLeft:
		r.X = float32(other.X) + padding
		r.Y = float32(bottom) - r.H - padding
	case AnchorBottomCenter:
		r.X = float32(other.X) + float32(other.W)/2 - r.W/2
		r.Y = float32(bottom) - r.H - padding
	case AnchorBottomRight:
		r.X = float32(right) - r.W - padding
		r.Y = float32(bottom) - r.H - padding
		// default:
		// 	return errors.New("can't align area to an image using an unsupported alignment type")
	}

	return r
}

func (r Rect) AlignToScreenbuffer(anchor AnchorPosition, padding float32) Rect {
	return r.AlignToImage(screenBuffer, anchor, padding)
}

func (r Rect) AlignToImage(img *ebiten.Image, anchor AnchorPosition, padding float32) Rect {
	bounds := img.Bounds()
	return r.AlignToRect(Rect{float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy())}, anchor, padding)
}

func (r Rect) overlappingAxisX(other Rect) float32 {
	left := max(r.X, other.X)
	right := min(r.Right(), other.Right())
	return right - left
}

func (r Rect) overlappingAxisY(other Rect) float32 {
	top := max(r.Y, other.Y)
	bottom := min(r.Bottom(), other.Bottom())
	return bottom - top
}

func (r Rect) Lerp(other Rect, percentage float32) Rect {
	r.X += (other.X - r.X) * percentage
	r.Y += (other.Y - r.Y) * percentage
	r.W += (other.W - r.W) * percentage
	r.H += (other.H - r.H) * percentage
	return r
}

func (r Rect) LerpXY(other Rect, percentage float32) Rect {
	r.X += (other.X - r.X) * percentage
	r.Y += (other.Y - r.Y) * percentage
	return r
}

func (r Rect) ContainsPoint(vec Vector2) bool {
	return vec.X >= r.X && vec.X <= r.X+r.W && vec.Y >= r.Y && vec.Y <= r.Y+r.H
}

// func (r Rect) Overlap(other Rect) Rect {

// }
