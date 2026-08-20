package model

import (
	"fmt"
	"math"
)

type Vector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func NewVector(x, y, z float64) Vector { return Vector{X: x, Y: y, Z: z} }

func (v Vector) Length() float64 { return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z) }

func (v Vector) HorizontalLength() float64 { return math.Hypot(v.X, v.Y) }

func (v Vector) Declination() float64 {
	angle := math.Atan2(v.Y, v.X) * 180 / math.Pi
	if angle < 0 {
		angle += 360
	}
	return angle
}

func (v Vector) Inclination() float64 {
	return math.Atan2(v.Z, v.HorizontalLength()) * 180 / math.Pi
}

func (v Vector) Normalize() (Vector, error) {
	length := v.Length()
	if length == 0 || math.IsNaN(length) || math.IsInf(length, 0) {
		return Vector{}, fmt.Errorf("%w: zero vector cannot be normalized", ErrInvalid)
	}
	return Vector{X: v.X / length, Y: v.Y / length, Z: v.Z / length}, nil
}

func (v Vector) Dot(other Vector) float64 { return v.X*other.X + v.Y*other.Y + v.Z*other.Z }

func (v Vector) Cross(other Vector) Vector {
	return Vector{X: v.Y*other.Z - v.Z*other.Y, Y: v.Z*other.X - v.X*other.Z, Z: v.X*other.Y - v.Y*other.X}
}

func (v Vector) Add(other Vector) Vector {
	return Vector{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

func (v Vector) Sub(other Vector) Vector {
	return Vector{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

func (v Vector) Scale(factor float64) Vector {
	return Vector{X: v.X * factor, Y: v.Y * factor, Z: v.Z * factor}
}

func (v Vector) Distance(other Vector) float64 { return v.Sub(other).Length() }

func (v Vector) IsFinite() bool {
	return !math.IsNaN(v.X) && !math.IsNaN(v.Y) && !math.IsNaN(v.Z) && !math.IsInf(v.X, 0) && !math.IsInf(v.Y, 0) && !math.IsInf(v.Z, 0)
}

func (v Vector) Validate() error {
	if !v.IsFinite() {
		return fmt.Errorf("%w: vector contains non-finite value", ErrInvalid)
	}
	if v.Length() == 0 {
		return fmt.Errorf("%w: vector must not be zero", ErrInvalid)
	}
	return nil
}

func VectorFromMeasurement(item Measurement) Vector { return Vector{X: item.X, Y: item.Y, Z: item.Z} }

func AverageVectors(items []Vector) (Vector, error) {
	if len(items) == 0 {
		return Vector{}, fmt.Errorf("%w: no vectors", ErrInvalid)
	}
	var sum Vector
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return Vector{}, err
		}
		sum = sum.Add(item)
	}
	return sum.Scale(1 / float64(len(items))), nil
}

func AngularDistance(left, right Vector) (float64, error) {
	a, err := left.Normalize()
	if err != nil {
		return 0, err
	}
	b, err := right.Normalize()
	if err != nil {
		return 0, err
	}
	dot := a.Dot(b)
	if dot > 1 {
		dot = 1
	}
	if dot < -1 {
		dot = -1
	}
	return math.Acos(dot) * 180 / math.Pi, nil
}

func RotateAroundZ(v Vector, degrees float64) Vector {
	radians := degrees * math.Pi / 180
	cosine, sine := math.Cos(radians), math.Sin(radians)
	return Vector{X: v.X*cosine - v.Y*sine, Y: v.X*sine + v.Y*cosine, Z: v.Z}
}
