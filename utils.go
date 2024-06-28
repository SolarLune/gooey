package gooey

import (
	"image"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
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

func DrawNinepatch(target, ninepatchImg *ebiten.Image, x, y, w, h float32, opt *ebiten.DrawImageOptions) {

	bounds := ninepatchImg.Bounds()
	chunkW := bounds.Dx() / 3
	chunkH := bounds.Dy() / 3

	origin := bounds.Min

	topLeft := ninepatchImg.SubImage(image.Rect(origin.X, origin.Y, origin.X+chunkW, origin.Y+chunkH)).(*ebiten.Image)
	opt.GeoM.Translate(float64(x), float64(y))
	target.DrawImage(topLeft, opt)

	top := ninepatchImg.SubImage(image.Rect(origin.X+chunkW, origin.Y, origin.X+chunkW*2, origin.Y+chunkH)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Scale((float64(w)-float64(chunkW*2))/float64(chunkW), 1)
	opt.GeoM.Translate(float64(x)+float64(chunkW), float64(y))
	target.DrawImage(top, opt)

	topRight := ninepatchImg.SubImage(image.Rect(origin.X+chunkW*2, origin.Y, origin.X+chunkW*3, origin.Y+chunkH)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Translate(float64(x+w)-float64(chunkW), float64(y))
	target.DrawImage(topRight, opt)

	midHeight := (float64(h) - float64(chunkH*2)) / float64(chunkH)
	// fmt.Println(textHeight)
	if midHeight > 0 {
		midLeft := ninepatchImg.SubImage(image.Rect(origin.X, origin.Y+chunkH, origin.X+chunkW, origin.Y+chunkH*2)).(*ebiten.Image)
		opt.GeoM.Reset()
		opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x), float64(y)+float64(chunkH))
		target.DrawImage(midLeft, opt)

		mid := ninepatchImg.SubImage(image.Rect(origin.X+chunkW, origin.Y+chunkH, origin.X+chunkW*2, origin.Y+chunkH*2)).(*ebiten.Image)
		opt.GeoM.Reset()
		opt.GeoM.Scale((float64(w)-float64(chunkW*2))/float64(chunkW), midHeight)
		opt.GeoM.Translate(float64(x)+float64(chunkW), float64(y)+float64(chunkH))
		target.DrawImage(mid, opt)

		midRight := ninepatchImg.SubImage(image.Rect(origin.X+chunkW*2, origin.Y+chunkH, origin.X+chunkW*3, origin.Y+chunkH*2)).(*ebiten.Image)
		opt.GeoM.Reset()
		opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x+w)-float64(chunkW), float64(y)+float64(chunkH))
		target.DrawImage(midRight, opt)
	} else {
		midHeight = 0
	}

	bottomLeft := ninepatchImg.SubImage(image.Rect(origin.X, origin.Y+chunkH*2, origin.X+chunkW, origin.Y+chunkH*3)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Translate(float64(x), float64(y+h)-float64(chunkH))
	target.DrawImage(bottomLeft, opt)

	bottom := ninepatchImg.SubImage(image.Rect(origin.X+chunkW, origin.Y+chunkH*2, origin.X+chunkW*2, origin.Y+chunkH*3)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Scale((float64(w)-float64(chunkW*2))/float64(chunkW), 1)
	opt.GeoM.Translate(float64(x)+float64(chunkW), float64(y+h)-float64(chunkH))
	target.DrawImage(bottom, opt)

	bottomRight := ninepatchImg.SubImage(image.Rect(origin.X+chunkH*2, origin.Y+chunkH*2, origin.X+chunkW*3, origin.Y+chunkH*3)).(*ebiten.Image)
	opt.GeoM.Reset()
	opt.GeoM.Translate(float64(x+w)-float64(chunkW), float64(y+h)-float64(chunkH))
	target.DrawImage(bottomRight, opt)

}

// Magnitude returns the length of the Vector (ignoring the Vector's W component).
func vecLength(x, y float32) float32 {
	return float32(math.Sqrt(float64(x*x + y*y)))
}

func vecNormalized(x, y float32) (float32, float32) {
	l := vecLength(x, y)
	return x / l, y * l
}
