package gooey

import (
	"image"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	chunkW := bounds.Dx() / 3
	chunkH := bounds.Dy() / 3

	origin := bounds.Min

	if horizontal {

		x += float32(chunkW / 2)
		y += float32(h/2) - float32(srcH/2)
		w -= float32(chunkW)

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

	} else {

		y += float32(chunkH / 2)
		x += float32(w/2) - float32(srcW/2)
		h -= float32(chunkH)

		// midHeight := (float64(h) - float64(chunkH*2)) / float64(chunkH)
		top := threepatchImg.SubImage(image.Rect(origin.X, origin.Y, origin.X+srcW, origin.Y+chunkH)).(*ebiten.Image)
		opt.GeoM.Reset()
		// opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x), float64(y))
		colorm.DrawImage(target, top, colorMat, opt)

		if ch := (float64(h) - float64(chunkH*2)) / float64(chunkH); ch > 0 {
			mid := threepatchImg.SubImage(image.Rect(origin.X, origin.Y+chunkH, origin.X+srcW, origin.Y+chunkH*2)).(*ebiten.Image)
			opt.GeoM.Reset()
			opt.GeoM.Scale(1, ch)
			opt.GeoM.Translate(float64(x), float64(y)+float64(chunkH))
			colorm.DrawImage(target, mid, colorMat, opt)
		}

		bottom := threepatchImg.SubImage(image.Rect(origin.X, origin.Y+chunkH*2, origin.X+srcW, origin.Y+chunkH*3)).(*ebiten.Image)
		opt.GeoM.Reset()
		// opt.GeoM.Scale(1, midHeight)
		opt.GeoM.Translate(float64(x), float64(y+h)-float64(chunkH))
		colorm.DrawImage(target, bottom, colorMat, opt)

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

// SubImage is a helper function to easily return a subimage of an *ebiten.Image.
func SubImage(img *ebiten.Image, x, y, w, h float32) *ebiten.Image {
	return img.SubImage(image.Rect(int(x), int(y), int(x+w), int(y+h))).(*ebiten.Image)
}

type sortedElementInstanceMap struct {
	Data  map[string]*uiElementInstance
	Order []string
}

func newSortedElementInstanceMap() *sortedElementInstanceMap {
	return &sortedElementInstanceMap{
		Data: map[string]*uiElementInstance{},
	}
}

func (s *sortedElementInstanceMap) Clone() *sortedElementInstanceMap {
	newMap := newSortedElementInstanceMap()
	for k, v := range s.Data {
		newMap.Data[k] = v.Clone()
	}
	newMap.Order = append(newMap.Order, s.Order...)
	return newMap
}

func (s *sortedElementInstanceMap) Add(id string) *uiElementInstance {

	if inst, exists := s.Data[id]; !exists {
		inst = &uiElementInstance{
			id: id,
		}
		s.Data[id] = inst
		s.Order = append(s.Order, id)
		return inst
	} else {
		return inst
	}

}

func (s *sortedElementInstanceMap) Contains(id string) bool {
	_, exists := s.Data[id]
	return exists
}

func (s *sortedElementInstanceMap) ForEach(forEach func(instance *uiElementInstance) bool) {
	for _, id := range s.Order {
		if !forEach(s.Data[id]) {
			break
		}
	}
}

var glyphs []text.Glyph
var glyphsWithPadding = map[*ebiten.Image]*ebiten.Image{}

// var glyphsPool = sync.Pool{
// 	New: func() any {
// 		return &[]text.Glyph{}
// 	},
// }

func drawTextWithShader(dst *ebiten.Image, txt string, face text.Face, options *text.DrawOptions, uniforms map[string]any, shader *ebiten.Shader) {
	var layoutOp text.LayoutOptions
	var drawOp ebiten.DrawImageOptions

	if options != nil {
		layoutOp = options.LayoutOptions
		drawOp = options.DrawImageOptions
	}

	geoM := drawOp.GeoM

	paddingAmount := 64

	if len(glyphs) == 0 {
		glyphs = text.AppendGlyphs(glyphs, txt, face, &layoutOp)
		for _, g := range glyphs {
			if g.Image == nil {
				continue
			}
			if _, exists := glyphsWithPadding[g.Image]; !exists {
				withPadding := ebiten.NewImage(g.Image.Bounds().Dx()+paddingAmount, g.Image.Bounds().Dy()+paddingAmount)
				opt := &ebiten.DrawImageOptions{}
				opt.GeoM.Translate(float64(paddingAmount)/2, float64(paddingAmount)/2)
				withPadding.DrawImage(g.Image, opt)
				glyphsWithPadding[g.Image] = withPadding
			}
		}
	}

	// glyphs := glyphsPool.Get().(*[]text.Glyph)
	// defer func() {
	// 	// Clear the content to avoid memory leaks.
	// 	// The capacity is kept so that the next call to Draw can reuse it.
	// 	*glyphs = slices.Delete(*glyphs, 0, len(*glyphs))
	// 	glyphsPool.Put(glyphs)
	// }()
	// *glyphs = text.AppendGlyphs((*glyphs)[:0], txt, face, &layoutOp)

	iterationStart := 0
	iterationDirection := 1
	iterationEnd := len(glyphs) - 1

	if textStyle.ShadowDirectionX > 0 || textStyle.ShadowDirectionY > 0 {
		iterationStart = len(glyphs) - 1
		iterationDirection = -1
		iterationEnd = 0
	}

	for i := iterationStart; i != iterationEnd+iterationDirection; i += iterationDirection {
		if i < 0 || i >= len(glyphs) {
			break
		}
		g := glyphs[i]
		if g.Image == nil {
			continue
		}
		drawOp.GeoM.Reset()
		drawOp.GeoM.Translate(g.X-float64(paddingAmount/2), g.Y-float64(paddingAmount/2))
		drawOp.GeoM.Concat(geoM)
		// dst.DrawImage(g.Image, &drawOp)

		glyphImage := glyphsWithPadding[g.Image]

		dst.DrawRectShader(glyphImage.Bounds().Dx(), glyphImage.Bounds().Dy(), shader, &ebiten.DrawRectShaderOptions{
			GeoM:       drawOp.GeoM,
			ColorScale: drawOp.ColorScale,
			Blend:      drawOp.Blend,
			Uniforms:   uniforms,
			Images: [4]*ebiten.Image{
				glyphImage,
			},
		})
	}

	glyphs = glyphs[:0]
}

func setTextForAllLabelsInGraphic(graphic UIElement, txt string) {

	// We do this dynamically because the selected option can change
	switch el := graphic.(type) {
	case UICollection:
		el.ForAllElementsRecursive(func(e UIElement) (UIElement, bool) {
			if label, ok := e.(UILabel); ok {
				label.Text = txt
				return label, true
			}
			return e, true
		})
	case UILabel:
		el.Text = txt
	}

}
