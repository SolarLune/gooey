package gooey

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// This is just cribbing the Color struct from Tetra3D.

// Color represents a color, containing R, G, B, and A components, each expected to range from 0 to 1.
type Color struct {
	R, G, B, A float32
}

// NewColor returns a new Color, with the provided R, G, B, and A components expected to range from 0 to 1.
func NewColor(r, g, b, a float32) Color {
	return Color{r, g, b, a}
}

// NewColorFromColor returns a new Color based on a provided color.Color.
func NewColorFromColor(color color.Color) Color {

	r, g, b, a := color.RGBA()

	return NewColor(
		float32(r)/65535,
		float32(g)/65535,
		float32(b)/65535,
		float32(a)/65535,
	)

}

// NewColorRandom creates a randomized color, with each component lying between the minimum and maximum values.
func NewColorRandom(min, max float32, grayscale bool) Color {
	color := NewColor(1, 1, 1, 1)
	diff := max - min
	if grayscale {
		r := min + (diff * rand.Float32())
		color.R = r
		color.G = r
		color.B = r
	} else {
		color.R = min + (diff * rand.Float32())
		color.G = min + (diff * rand.Float32())
		color.B = min + (diff * rand.Float32())
	}
	return color
}

// AddRGBA adds the provided R, G, B, and A values to the color as provided. The components are expected to range from 0 to 1.
func (color Color) AddRGBA(r, g, b, a float32) Color {
	color.R += r
	color.G += g
	color.B += b
	color.A += a
	return color
}

// Add adds the provided Color to the existing Color.
func (color Color) Add(other Color) Color {
	return color.AddRGBA(other.ToFloat32s())
}

// MultiplyRGBA multiplies the color's RGBA channels by the provided R, G, B, and A scalar values.
func (color Color) MultiplyRGBA(scalarR, scalarG, scalarB, scalarA float32) Color {
	color.R *= scalarR
	color.G *= scalarG
	color.B *= scalarB
	color.A *= scalarA
	return color
}

func (color Color) MultiplyScalarRGB(scalar float32) Color {
	return color.MultiplyRGBA(scalar, scalar, scalar, 1)
}

func (color Color) MultiplyScalarRGBA(scalar float32) Color {
	return color.MultiplyRGBA(scalar, scalar, scalar, scalar)
}

// Multiply multiplies the existing Color by the provided Color.
func (color Color) Multiply(other Color) Color {
	return color.MultiplyRGBA(other.ToFloat32s())
}

// Sub subtracts the other Color from the calling Color instance.
func (color Color) SubRGBA(r, g, b, a float32) Color {
	color.R -= r
	color.G -= g
	color.B -= b
	color.A -= a
	return color
}

// Sub subtracts the other Color from the calling Color instance.
func (color Color) Sub(other Color) Color {
	return color.SubRGBA(other.ToFloat32s())
}

// Mix mixes the calling Color with the other Color, mixed to the percentage given (ranging from 0 - 1).
func (color Color) Mix(other Color, percentage float32) Color {

	p := clamp(float64(percentage), 0, 1)
	percentage = float32(p)

	color.R += (other.R - color.R) * percentage
	color.G += (other.G - color.G) * percentage
	color.B += (other.B - color.B) * percentage
	color.A += (other.A - color.A) * percentage
	return color

}

// AddAlpha adds the provided alpha amount to the Color
func (c Color) AddAlpha(alpha float32) Color {
	c.A += alpha
	return c
}

// SubAlpha adds the provided alpha amount to the Color
func (c Color) SubAlpha(alpha float32) Color {
	c.A -= alpha
	return c
}

// SetAlpha returns a copy of the the Color with the alpha set to the provided alpha value.
func (color Color) SetAlpha(alpha float32) Color {
	color.A = alpha
	return color
}

// ToFloat32s returns the Color as four float32 in the order R, G, B, and A.
func (color Color) ToFloat32s() (float32, float32, float32, float32) {
	return color.R, color.G, color.B, color.A
}

// ToFloat64s returns four float64 values for each channel in the Color in the order R, G, B, and A.
func (color Color) ToFloat64s() (float64, float64, float64, float64) {
	return float64(color.R), float64(color.G), float64(color.B), float64(color.A)
}

// ToFloat32Slice returns a [4]float32 array for each channel in the Color in the order of R, G, B, and A.
func (color Color) ToFloat32Slice() [4]float32 {
	return [4]float32{float32(color.R), float32(color.G), float32(color.B), float32(color.A)}
}

// ToRGBA64 converts a color to a color.RGBA64 instance.
func (c Color) ToRGBA64() color.RGBA64 {
	return color.RGBA64{
		c.capRGBA64(c.R),
		c.capRGBA64(c.G),
		c.capRGBA64(c.B),
		c.capRGBA64(c.A),
	}
}

// ToNRGBA64 converts a color to a color.NRGBA64 (non-alpha color multiplied) color instance.
func (c Color) ToNRGBA64() color.NRGBA64 {
	return color.NRGBA64{
		c.capRGBA64(c.R),
		c.capRGBA64(c.G),
		c.capRGBA64(c.B),
		c.capRGBA64(c.A),
	}
}

func (color Color) capRGBA64(value float32) uint16 {
	if value > 1 {
		value = 1
	} else if value < 0 {
		value = 0
	}
	return uint16(value * math.MaxUint16)
}

// ConvertTosRGB() converts the color's R, G, and B components to the sRGB color space. This is used to convert
// colors from their values in GLTF to how they should appear on the screen. See: https://en.wikipedia.org/wiki/SRGB
func (color Color) ConvertTosRGB() Color {

	if color.R <= 0.0031308 {
		color.R *= 12.92
	} else {
		color.R = float32(1.055*math.Pow(float64(color.R), 1/2.4) - 0.055)
	}

	if color.G <= 0.0031308 {
		color.G *= 12.92
	} else {
		color.G = float32(1.055*math.Pow(float64(color.G), 1/2.4) - 0.055)
	}

	if color.B <= 0.0031308 {
		color.B *= 12.92
	} else {
		color.B = float32(1.055*math.Pow(float64(color.B), 1/2.4) - 0.055)
	}

	return color

}

// Lerp linearly interpolates the color from the starting color to the target by the percentage given.
func (c Color) Lerp(other Color, percentage float64) Color {

	percentage = clamp(percentage, 0, 1)
	c.R += (other.R - c.R) * float32(percentage)
	c.G += (other.G - c.G) * float32(percentage)
	c.B += (other.B - c.B) * float32(percentage)
	c.A += (other.A - c.A) * float32(percentage)

	return c
}

// Lerp linearly interpolates the color from the starting color to the target by the percentage given.
func (c Color) LerpRGBA(r, g, b, a, percentage float64) Color {

	percentage = clamp(percentage, 0, 1)
	c.R += (float32(r) - c.R) * float32(percentage)
	c.G += (float32(g) - c.G) * float32(percentage)
	c.B += (float32(b) - c.B) * float32(percentage)
	c.A += (float32(a) - c.A) * float32(percentage)

	return c
}

func (color Color) String() string {
	return fmt.Sprintf("<%0.2f, %0.2f, %0.2f, %0.2f>", color.R, color.G, color.B, color.A)
}

// NewColorFromHSV returns a new color, using hue, saturation, and value numbers, each ranging from 0 to 1. A hue of
// 0 is red, while 1 is also red, but on the other end of the spectrum.
// Cribbed from: https://github.com/lucasb-eyer/go-colorful/blob/master/colors.go
func NewColorFromHSV(h, s, v float64) Color {

	for h > 1 {
		h--
	}
	for h < 0 {
		h++
	}

	h *= 360

	if s > 1 {
		s = 1
	} else if s < 0 {
		s = 0
	}

	if v > 1 {
		v = 1
	} else if v < 0 {
		v = 0
	}

	Hp := h / 60.0
	C := v * s
	X := C * (1.0 - math.Abs(math.Mod(Hp, 2.0)-1.0))

	m := v - C
	r, g, b := 0.0, 0.0, 0.0

	switch {
	case 0.0 <= Hp && Hp < 1.0:
		r = C
		g = X
	case 1.0 <= Hp && Hp < 2.0:
		r = X
		g = C
	case 2.0 <= Hp && Hp < 3.0:
		g = C
		b = X
	case 3.0 <= Hp && Hp < 4.0:
		g = X
		b = C
	case 4.0 <= Hp && Hp < 5.0:
		r = X
		b = C
	case 5.0 <= Hp && Hp < 6.0:
		r = C
		b = X
	}

	return Color{float32(m + r), float32(m + g), float32(m + b), 1}
}

func NewColorFromHexString(hex string) Color {

	c := NewColor(0, 0, 0, 1)

	hex = strings.TrimPrefix(hex, "#")

	if len(hex) >= 2 {
		v, _ := strconv.ParseInt(hex[:2], 16, 32)
		c.R = float32(v) / 256.0

		if len(hex) >= 4 {
			v, _ := strconv.ParseInt(hex[2:4], 16, 32)
			c.G = float32(v) / 256.0
		} else {
			c.G = 1
		}

		if len(hex) >= 6 {
			v, _ := strconv.ParseInt(hex[4:6], 16, 32)
			c.B = float32(v) / 256.0
		} else {
			c.B = 1
		}

		if len(hex) >= 8 {
			v, _ := strconv.ParseInt(hex[6:8], 16, 32)
			c.A = float32(v) / 256.0
		} else {
			c.A = 1
		}

	}

	return c

}

// Hue returns the hue of the color as a value ranging from 0 to 1.
func (color Color) Hue() float64 {
	// Function cribbed from: https://github.com/lucasb-eyer/go-colorful/blob/master/colors.go

	r := float64(color.R)
	g := float64(color.G)
	b := float64(color.B)

	min := math.Min(math.Min(r, g), b)
	v := math.Max(math.Max(r, g), b)
	C := v - min

	h := 0.0
	if min != v {
		if v == r {
			h = math.Mod((g-b)/C, 6.0)
		}
		if v == g {
			h = (b-r)/C + 2.0
		}
		if v == b {
			h = (r-g)/C + 4.0
		}
		h *= 60.0
		if h < 0.0 {
			h += 360.0
		}
	}
	return h / 360
}

// Saturation returns the saturation of the color as a value ranging from 0 to 1.
func (color Color) Saturation() float64 {

	r := float64(color.R)
	g := float64(color.G)
	b := float64(color.B)

	min := math.Min(math.Min(r, g), b)
	v := math.Max(math.Max(r, g), b)
	C := v - min

	s := 0.0
	if v != 0.0 {
		s = C / v
	}

	return s
}

// Value returns the value of the color as a value, ranging from 0 to 1.
func (color Color) Value() float64 {

	r := float64(color.R)
	g := float64(color.G)
	b := float64(color.B)

	return math.Max(math.Max(r, g), b)
}

func (color Color) Inverted() Color {
	color.R = 1 - color.R
	color.G = 1 - color.G
	color.B = 1 - color.B
	return color
}

// func (color Color) HSV() (float64, float64, float64) {

// 	r := float64(color.R)
// 	g := float64(color.G)
// 	b := float64(color.B)

// 	min := math.Min(math.Min(r, g), b)
// 	v := math.Max(math.Max(r, g), b)
// 	C := v - min

// 	s := 0.0
// 	if v != 0.0 {
// 		s = C / v
// 	}

// 	h := 0.0
// 	if min != v {
// 		if v == r {
// 			h = math.Mod((g-b)/C, 6.0)
// 		}
// 		if v == g {
// 			h = (b-r)/C + 2.0
// 		}
// 		if v == b {
// 			h = (r-g)/C + 4.0
// 		}
// 		h *= 60.0
// 		if h < 0.0 {
// 			h += 360.0
// 		}
// 	}
// 	return h / 360, s, v
// }

// SetHue returns a copy of the color with the hue set to the specified value.
// Hue goes through the rainbow, starting with red, and ranges from 0 to 1.
func (color Color) SetHue(h float64) Color {
	return NewColorFromHSV(h, color.Saturation(), color.Value())
}

// SetSaturation returns a copy of the color with the saturation of the color set to the specified value.
// Saturation ranges from 0 to 1.
func (color Color) SetSaturation(s float64) Color {
	return NewColorFromHSV(color.Hue(), s, color.Value())
}

// SetValue returns a copy of the color with the value of the color set to the specified value.
// Value ranges from 0 to 1.
func (color Color) SetValue(v float64) Color {
	return NewColorFromHSV(color.Hue(), color.Saturation(), v)
}

func (c Color) IsZero() bool {
	return c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0
}
