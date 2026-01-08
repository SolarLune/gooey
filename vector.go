package gooey

import (
	"fmt"
	"math"
	"strings"
)

// Vector2 represents a 2D Vector2, which can be used for usual 2D applications (position, direction, velocity, etc).
// Any Vector2 functions that modify the calling Vector2 return copies of the modified Vector2, meaning you can do method-chaining easily.
type Vector2 struct {
	X float32 // The X (1st) component of the Vector
	Y float32 // The Y (2nd) component of the Vector
}

// NewVector2 creates a new Vector with the specified x and y components.
func NewVector2(x, y float32) Vector2 {
	return Vector2{X: x, Y: y}
}

// This clones the Vector3 - this doesn't do anything by itself, but is useful for consistency between other API elements and when used with scripting languages like goja.
func (vec Vector2) Clone() Vector2 {
	return vec
}

// String returns a string representation of the Vector.
func (vec Vector2) String() string {
	return fmt.Sprintf("{%.2f, %.2f}", vec.X, vec.Y)
}

// Add returns a copy of the calling vector, added together with the other Vector provided.
func (vec Vector2) Add(other Vector2) Vector2 {
	vec.X += other.X
	vec.Y += other.Y
	return vec
}

// AddX returns a copy of the calling vector with an added value to the X axis.
func (vec Vector2) AddX(x float32) Vector2 {
	vec.X += x
	return vec
}

// AddY returns a copy of the calling vector with an added value to the Y axis.
func (vec Vector2) AddY(y float32) Vector2 {
	vec.Y += y
	return vec
}

// Sub returns a copy of the calling Vector, with the other Vector subtracted from it.
func (vec Vector2) Sub(other Vector2) Vector2 {
	vec.X -= other.X
	vec.Y -= other.Y
	return vec
}

// SubX returns a copy of the calling vector with an added value to the X axis.
func (vec Vector2) SubX(x float32) Vector2 {
	vec.X -= x
	return vec
}

// SubY returns a copy of the calling vector with an added value to the Y axis.
func (vec Vector2) SubY(y float32) Vector2 {
	vec.Y -= y
	return vec
}

// Expand expands the Vector by the margin specified, in absolute units, if each component is over the minimum argument.
// To illustrate: Given a Vector of {1, 0.1, -0.3}, Vector.Expand(0.5, 0.2) would give you a Vector of {1.5, 0.1, -0.8}.
// This function returns a copy of the Vector with the result.
func (vec Vector2) Expand(margin, min float32) Vector2 {
	if vec.X > min || vec.X < -min {
		vec.X += float32(math.Copysign(float64(margin), float64(vec.X)))
	}
	if vec.Y > min || vec.Y < -min {
		vec.Y += float32(math.Copysign(float64(margin), float64(vec.Y)))
	}
	return vec
}

// Invert returns a copy of the Vector with all components inverted.
func (vec Vector2) Invert() Vector2 {
	vec.X = -vec.X
	vec.Y = -vec.Y
	return vec
}

// Magnitude returns the length of the Vector.
func (vec Vector2) Magnitude() float32 {
	return float32(math.Sqrt(float64(vec.X*vec.X + vec.Y*vec.Y)))
}

// MagnitudeSquared returns the squared length of the Vector; this is faster than Length() as it avoids using math.math.Sqrt().
func (vec Vector2) MagnitudeSquared() float32 {
	return vec.X*vec.X + vec.Y*vec.Y
}

// ClampMagnitude clamps the overall magnitude of the Vector to the maximum magnitude specified, returning a copy with the result.
func (vec Vector2) ClampMagnitude(maxMag float32) Vector2 {
	if vec.Magnitude() > maxMag {
		vec = vec.Unit().Scale(maxMag)
	}
	return vec
}

// Clamp clamps the Vector2 to the maximum values provided.
func (vec Vector2) Clamp(x, y float32) Vector2 {
	vec.X = clamp(vec.X, -x, x)
	vec.Y = clamp(vec.Y, -y, y)
	return vec
}

// ClampToVec clamps the Vector2 to the maximum values in the Vector2 provided.
func (vec Vector2) ClampToVec(extents Vector2) Vector2 {
	vec.X = clamp(vec.X, -extents.X, extents.X)
	vec.Y = clamp(vec.Y, -extents.Y, extents.Y)
	return vec
}

// AddMagnitude adds magnitude to the Vector in the direction it's already pointing.
func (vec Vector2) AddMagnitude(mag float32) Vector2 {
	return vec.Add(vec.Unit().Scale(mag))
}

// SubMagnitude subtracts the given magnitude from the Vector's existing magnitude.
// If the vector's magnitude is less than the given magnitude to subtract, a zero-length Vector will be returned.
func (vec Vector2) SubMagnitude(mag float32) Vector2 {
	if vec.Magnitude() > mag {
		return vec.Sub(vec.Unit().Scale(mag))
	}
	return Vector2{0, 0}

}

// MoveTowards moves a Vector2 towards another Vector2 given a specific magnitude. If the distance is less than that magnitude, it returns the target vector.
func (vec Vector2) MoveTowards(target Vector2, magnitude float32) Vector2 {
	diff := target.Sub(vec)
	if diff.Magnitude() > magnitude {
		return vec.Add(diff.Unit().Scale(magnitude))
	}
	return target
}

// Distance returns the distance from the calling Vector to the other Vector provided.
func (vec Vector2) DistanceTo(other Vector2) float32 {
	vec.X -= other.X
	vec.Y -= other.Y
	return float32(math.Sqrt(float64(vec.X*vec.X + vec.Y*vec.Y)))
}

// Distance returns the squared distance from the calling Vector to the other Vector provided. This is faster than Distance(), as it avoids using math.math.Sqrt().
func (vec Vector2) DistanceSquaredTo(other Vector2) float32 {
	vec.X -= other.X
	vec.Y -= other.Y
	return vec.X*vec.X + vec.Y*vec.Y
}

// Mult performs Hadamard (component-wise) multiplication on the calling Vector with the other Vector provided, returning a copy with the result.
func (vec Vector2) Mult(other Vector2) Vector2 {
	vec.X *= other.X
	vec.Y *= other.Y
	return vec
}

// Unit returns a copy of the Vector, normalized (set to be of unit length).
func (vec Vector2) Unit() Vector2 {
	l := vec.Magnitude()
	if l < 1e-8 || l == 1 {
		// If it's 0, then don't modify the vector
		return vec
	}
	vec.X, vec.Y = vec.X/l, vec.Y/l
	return vec
}

// Swizzle2 swizzles the Vector3 using the string provided, returning the swizzled copy as a Vector2.
// The string should be composed of the axes of a vector, i.e. 'x', 'y', 'z', or 'w'.
// If the string is shorter than 3 values, the remaining values are left at 0.
// Example: `vec := Vector{1, 2, 3}.Swizzle2("-z-x") // Returns a Vector2 of {-3, -1}.`
func (vec Vector2) Swizzle2(swizzleString string) Vector2 {

	out := Vector2{}

	swizzleString = strings.ToLower(swizzleString)

	ogX := vec.X
	ogY := vec.Y
	targetValue := float32(0)
	negating := false

	for i, v := range swizzleString {

		switch v {
		case 'x':
			targetValue = ogX
		case 'y':
			targetValue = ogY
		case '-':
			negating = true
		default:
			continue
		}

		if negating {
			targetValue *= -1
		}

		switch i {
		case 0:
			out.X = targetValue
		case 1:
			out.Y = targetValue
		}

		negating = false

	}

	return out

}

// SetX sets the X component in the vector to the value provided.
func (vec Vector2) SetX(x float32) Vector2 {
	vec.X = x
	return vec
}

// SetY sets the Y component in the vector to the value provided.
func (vec Vector2) SetY(y float32) Vector2 {
	vec.Y = y
	return vec
}

// SetXY sets the values in the Vector to the x, y, and z values provided.
func (vec Vector2) SetXY(x, y float32) Vector2 {
	vec.X = x
	vec.Y = y
	return vec
}

// Set sets the values in the Vector3 to the values in the other Vector2.
func (vec *Vector2) Set(other Vector2) {
	vec.X = other.X
	vec.Y = other.Y
}

// Reflect reflects the vector against the given surface normal.
func (vec Vector2) Reflect(normal Vector2) Vector2 {
	n := normal.Unit()
	return vec.Sub(n.Scale(2 * n.Dot(vec)))
}

// Perp returns the right-handed perpendicular of the vector (i.e. the vector rotated 90 degrees to the right, clockwise).
func (vec Vector2) Perp() Vector2 {
	return Vector2{-vec.Y, vec.X}
}

// Floats returns a [2]float32 array consisting of the Vector's contents.
func (vec Vector2) Floats() [2]float32 {
	return [2]float32{vec.X, vec.Y}
}

// Equals returns true if the two Vectors are close enough in all values (excluding W).
func (vec Vector2) Equals(other Vector2) bool {

	eps := float64(1e-4)

	if math.Abs(float64(vec.X-other.X)) > eps || math.Abs(float64(vec.Y-other.Y)) > eps {
		return false
	}

	return true

}

// IsZero returns true if the values in the Vector are extremely close to 0 (excluding W).
func (vec Vector2) IsZero() bool {

	eps := float64(1e-4)

	if math.Abs(float64(vec.X)) > eps || math.Abs(float64(vec.Y)) > eps {
		return false
	}

	// if !onlyXYZ && math.math.Abs(vec.W-other.W) > eps {
	// 	return false
	// }

	return true

}

// IsNaN returns if any of the values are infinite.
func (vec Vector2) IsNaN() bool {
	return math.IsNaN(float64(vec.X)) || math.IsNaN(float64(vec.Y))
}

// IsInf returns if any of the values are infinite.
func (vec Vector2) IsInf() bool {
	return math.IsInf(float64(vec.X), 0) || math.IsInf(float64(vec.Y), 0)
}

// Rotate returns a copy of the Vector, rotated around an axis that pierces the screen by the angle
// provided (in radians), counter-clockwise.
func (vec Vector2) Rotate(angle float32) Vector2 {
	x := vec.X
	y := vec.Y
	vec.X = x*float32(math.Cos(float64(angle))) - y*float32(math.Sin(float64(angle)))
	vec.Y = x*float32(math.Sin(float64(angle))) + y*float32(math.Cos(float64(angle)))
	return vec
}

// Angle returns the signed angle in radians between the calling Vector and the provided other Vector.
func (vec Vector2) Angle(other Vector2) float32 {
	angle := math.Atan2(float64(other.Y), float64(other.X)) - math.Atan2(float64(vec.Y), float64(vec.X))
	if angle > math.Pi {
		angle -= 2 * math.Pi
	} else if angle <= -math.Pi {
		angle += 2 * math.Pi
	}
	return float32(angle)
}

// Floor floors the Vector's components off, returning a new Vector2.
// For example, Vector2{0.1, 1.87}.Floor() will return Vector2{0, 1}.
func (vec Vector2) Floor() Vector2 {
	vec.X = float32(math.Floor(float64(vec.X)))
	vec.Y = float32(math.Floor(float64(vec.Y)))
	return vec
}

// Ceil ceils the Vector's components off, returning a new Vector2.
// For example, Vector2{0.1, 1.27}.Ceil() will return Vector2{1, 2}.
func (vec Vector2) Ceil() Vector2 {
	vec.X = float32(math.Ceil(float64(vec.X)))
	vec.Y = float32(math.Ceil(float64(vec.Y)))
	return vec
}

// AngleRotation returns the angle in radians between this Vector and world right (1, 0).
func (vec Vector2) AngleRotation() float32 {
	return vec.Angle(Vector2{1, 0})
}

// Scale scales a Vector by the given scalar, returning a copy with the result.
func (vec Vector2) Scale(scalar float32) Vector2 {
	vec.X *= scalar
	vec.Y *= scalar
	return vec
}

// Divide divides a Vector by the given scalar, returning a copy with the result.
func (vec Vector2) Divide(scalar float32) Vector2 {
	vec.X /= scalar
	vec.Y /= scalar
	return vec
}

// Dot returns the dot product of a Vector and another Vector.
func (vec Vector2) Dot(other Vector2) float32 {
	return vec.X*other.X + vec.Y*other.Y
}

// Round rounds off the Vector's components to the given space in world unit increments, returning a clone
// (e.g. Vector{0.1, 1.27, 3.33}.Snap(0.25) will return Vector{0, 1.25, 3.25}).
func (vec Vector2) Round(snapToUnits float32) Vector2 {
	vec.X = float32(math.Round(float64(vec.X/snapToUnits))) * snapToUnits
	vec.Y = float32(math.Round(float64(vec.Y/snapToUnits))) * snapToUnits
	return vec
}

// ClampAngle clamps the Vector such that it doesn't exceed the angle specified (in radians).
// This function returns a normalized (unit) Vector.
func (vec Vector2) ClampAngle(baselineVec Vector2, maxAngle float32) Vector2 {

	mag := vec.Magnitude()

	angle := vec.Angle(baselineVec)

	if angle > maxAngle {
		vec = baselineVec.Slerp(vec, maxAngle/angle).Unit()
	}

	return vec.Scale(mag)

}

// Lerp performs a linear interpolation between the starting Vector and the provided
// other Vector, to the given percentage (ranging from 0 to 1).
func (vec Vector2) Lerp(other Vector2, percentage float32) Vector2 {
	// percentage = math.Clamp(percentage, 0, 1)
	vec.X = vec.X + ((other.X - vec.X) * percentage)
	vec.Y = vec.Y + ((other.Y - vec.Y) * percentage)
	return vec
}

// Slerp performs a spherical linear interpolation between the starting Vector and the provided
// ending Vector, to the given percentage (ranging from 0 to 1).
// This should be done with directions, usually, rather than positions.
// This being the case, this normalizes both Vectors.
func (vec Vector2) Slerp(targetDirection Vector2, percentage float32) Vector2 {

	vec = vec.Unit()
	targetDirection = targetDirection.Unit()

	// Thank you StackOverflow, once again! : https://stackoverflow.com/questions/67919193/how-does-unity-implements-vector3-slerp-exactly
	percentage = clamp(percentage, 0, 1)

	dot := vec.Dot(targetDirection)

	dot = clamp(dot, -1, 1)

	theta := float32(math.Acos(float64(dot))) * percentage
	relative := targetDirection.Sub(vec.Scale(dot)).Unit()

	return (vec.Scale(float32(math.Cos(float64(theta)))).Add(relative.Scale(float32(math.Sin(float64(theta)))))).Unit()

}

func (v Vector2) Inside(rect Rect) bool {
	return v.X >= rect.X && v.X < rect.X+rect.W && v.Y >= rect.Y && v.Y < rect.Y+rect.H
}
