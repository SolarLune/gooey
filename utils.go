package gooey

import (
	"image"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

func min[V float64 | float32 | int](a, b V) V {
	if a < b {
		return a
	}
	return b
}

func max[V float64 | float32 | int](a, b V) V {
	if a > b {
		return a
	}
	return b
}

func clamp[V float64 | float32 | int](value, min, max V) V {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

func pow(value float64, power int) float64 {
	x := value
	for i := 0; i < power; i++ {
		x += x
	}
	return x
}

func round(value float64) float64 {

	iv := float64(int(value))

	if value > iv+0.5 {
		return iv + 1
	} else if value < iv-0.5 {
		return iv - 1
	}

	return iv

}

func splitWithSeparator(str string, seps string) []string {

	output := []string{}
	start := 0

	index := strings.IndexAny(str, seps)

	if index < 0 {
		return []string{str}
	}

	for index >= 0 {
		end := start + index + 1
		if end > len(str) {
			end = len(str)
		}
		output = append(output, str[start:end])
		start += index + 1
		index = strings.IndexAny(str[start:], seps)
		if index < 0 {
			output = append(output, str[start:])
		}
	}

	return output

}

func DrawNinepatch(target, ninepatchImg *ebiten.Image, x, y, w, h float32, colorMat colorm.ColorM, drawOptions *colorm.DrawImageOptions) {

	d := *drawOptions
	opt := &d // Make a copy so we don't modify the original draw image options

	bounds := ninepatchImg.Bounds()
	chunkW := bounds.Dx() / 3
	chunkH := bounds.Dy() / 3

	origin := bounds.Min

	topLeft := ninepatchImg.SubImage(image.Rect(origin.X, origin.Y, origin.X+chunkW, origin.Y+chunkH)).(*ebiten.Image)
	opt.GeoM.Translate(float64(x), float64(y))

	colorm.DrawImage(target, topLeft, colorMat, opt)

	if cw := (float64(w) - float64(chunkW*2)) / float64(chunkW); cw > 0 {
		top := ninepatchImg.SubImage(image.Rect(origin.X+chunkW, origin.Y, origin.X+chunkW*2, origin.Y+chunkH)).(*ebiten.Image)
		opt.GeoM.Reset()
		opt.GeoM.Scale(cw, 1)
		opt.GeoM.Translate(float64(x)+float64(chunkW), float64(y))
		colorm.DrawImage(target, top, colorMat, opt)
	}

	topRight := ninepatchImg.SubImage(image.Rect(origin.X+chunkW*2, origin.Y, origin.X+chunkW*3, origin.Y+chunkH)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Translate(float64(x+w)-float64(chunkW), float64(y))
	colorm.DrawImage(target, topRight, colorMat, opt)

	midHeight := (float64(h) - float64(chunkH*2)) / float64(chunkH)
	// fmt.Println(textHeight)
	if midHeight > 0 {
		midLeft := ninepatchImg.SubImage(image.Rect(origin.X, origin.Y+chunkH, origin.X+chunkW, origin.Y+chunkH*2)).(*ebiten.Image)
		opt.GeoM.Reset()
		opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x), float64(y)+float64(chunkH))
		colorm.DrawImage(target, midLeft, colorMat, opt)

		if cw := (float64(w) - float64(chunkW*2)) / float64(chunkW); cw > 0 {
			mid := ninepatchImg.SubImage(image.Rect(origin.X+chunkW, origin.Y+chunkH, origin.X+chunkW*2, origin.Y+chunkH*2)).(*ebiten.Image)
			opt.GeoM.Reset()
			opt.GeoM.Scale(cw, midHeight)
			opt.GeoM.Translate(float64(x)+float64(chunkW), float64(y)+float64(chunkH))
			colorm.DrawImage(target, mid, colorMat, opt)
		}

		midRight := ninepatchImg.SubImage(image.Rect(origin.X+chunkW*2, origin.Y+chunkH, origin.X+chunkW*3, origin.Y+chunkH*2)).(*ebiten.Image)
		opt.GeoM.Reset()
		opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x+w)-float64(chunkW), float64(y)+float64(chunkH))
		colorm.DrawImage(target, midRight, colorMat, opt)
	} else {
		midHeight = 0
	}

	bottomLeft := ninepatchImg.SubImage(image.Rect(origin.X, origin.Y+chunkH*2, origin.X+chunkW, origin.Y+chunkH*3)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Translate(float64(x), float64(y+h)-float64(chunkH))
	colorm.DrawImage(target, bottomLeft, colorMat, opt)

	if cw := (float64(w) - float64(chunkW*2)) / float64(chunkW); cw > 0 {
		bottom := ninepatchImg.SubImage(image.Rect(origin.X+chunkW, origin.Y+chunkH*2, origin.X+chunkW*2, origin.Y+chunkH*3)).(*ebiten.Image)
		opt.GeoM.Reset()
		opt.GeoM.Scale(cw, 1)
		opt.GeoM.Translate(float64(x)+float64(chunkW), float64(y+h)-float64(chunkH))
		colorm.DrawImage(target, bottom, colorMat, opt)
	}

	bottomRight := ninepatchImg.SubImage(image.Rect(origin.X+chunkW*2, origin.Y+chunkH*2, origin.X+chunkW*3, origin.Y+chunkH*3)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Translate(float64(x+w)-float64(chunkW), float64(y+h)-float64(chunkH))
	colorm.DrawImage(target, bottomRight, colorMat, opt)

}

func DrawThreepatch(target, threepatchImg *ebiten.Image, x, y, w, h float32, horizontal bool, colorMat colorm.ColorM, drawOptions *colorm.DrawImageOptions) {

	d := *drawOptions
	opt := &d // Make a copy so we don't modify the original draw image options

	bounds := threepatchImg.Bounds()
	// srcW := bounds.Dx()
	srcH := bounds.Dy()
	chunkW := bounds.Dx() / 3
	// chunkH := bounds.Dy() / 3

	x += float32(chunkW / 2)
	y += float32(h/2) - float32(srcH/2)
	w -= float32(chunkW)

	origin := bounds.Min

	if horizontal {

		// midHeight := (float64(h) - float64(chunkH*2)) / float64(chunkH)
		left := threepatchImg.SubImage(image.Rect(origin.X, origin.Y, origin.X+chunkW, origin.Y+srcH)).(*ebiten.Image)
		opt.GeoM.Reset()
		// opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x), float64(y))
		colorm.DrawImage(target, left, colorMat, opt)

		if cw := (float64(w) - float64(chunkW*2)) / float64(chunkW); cw > 0 {
			mid := threepatchImg.SubImage(image.Rect(origin.X+chunkW, origin.Y, origin.X+chunkW*2, origin.Y+srcH)).(*ebiten.Image)
			opt.GeoM.Reset()
			opt.GeoM.Scale(cw, 1)
			opt.GeoM.Translate(float64(x)+float64(chunkW), float64(y))
			colorm.DrawImage(target, mid, colorMat, opt)
		}

		right := threepatchImg.SubImage(image.Rect(origin.X+chunkW*2, origin.Y, origin.X+chunkW*3, origin.Y+srcH)).(*ebiten.Image)
		opt.GeoM.Reset()
		// opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x+w)-float64(chunkW), float64(y))
		colorm.DrawImage(target, right, colorMat, opt)

	}

}

// Magnitude returns the length of the Vector (ignoring the Vector's W component).
func vecLength(x, y float32) float32 {
	return float32(math.Sqrt(float64(x*x + y*y)))
}

func vecNormalized(x, y float32) (float32, float32) {
	l := vecLength(x, y)
	return x / l, y / l
}
