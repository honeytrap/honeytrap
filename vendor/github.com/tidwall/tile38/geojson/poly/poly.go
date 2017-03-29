// Package poly provides polygon detection methods.
package poly

import "fmt"

// Point is simple 2D point
type Point struct {
	X, Y, Z float64
}

// InsideRect detects point is inside of another rect
func (p Point) InsideRect(rect Rect) bool {
	if p.X < rect.Min.X || p.X > rect.Max.X {
		return false
	}
	if p.Y < rect.Min.Y || p.Y > rect.Max.Y {
		return false
	}
	return true
}

// Polygon is series of points that make up a polygon
type Polygon []Point

// InsideRect detects polygon is inside of another rect
func (p Polygon) InsideRect(rect Rect) bool {
	if len(p) == 0 {
		return false
	}
	for _, p := range p {
		if !p.InsideRect(rect) {
			return false
		}
	}
	return true
}

// IntersectsRect detects polygon is inside of another rect
func (p Polygon) IntersectsRect(rect Rect) bool {
	if len(p) == 0 {
		return false
	}
	rectPoly := Polygon{rect.Min, {rect.Min.X, rect.Max.Y, 0}, rect.Max, {rect.Max.X, rect.Min.Y, 0}, rect.Min}
	return p.Intersects(rectPoly, nil)
}

// String returns a string representation of the polygon.
func (p Polygon) String() string {
	s := "{"
	for i, p := range p {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("{%v, %v}", p.X, p.Y)
	}
	s += "}"
	return s
}

// Rect is rectangle
type Rect struct {
	Min, Max Point
}

// Rect returns the bounding box rectangle for the polygon
func (p Polygon) Rect() Rect {
	var bbox Rect
	for i, p := range p {
		if i == 0 {
			bbox.Min = p
			bbox.Max = p
		} else {
			if p.X < bbox.Min.X {
				bbox.Min.X = p.X
			} else if p.X > bbox.Max.X {
				bbox.Max.X = p.X
			}
			if p.Y < bbox.Min.Y {
				bbox.Min.Y = p.Y
			} else if p.Y > bbox.Max.Y {
				bbox.Max.Y = p.Y
			}
		}
	}
	return bbox
}

// IntersectsRect detects if two bboxes intersect.
func (r Rect) IntersectsRect(rect Rect) bool {
	if r.Min.Y > rect.Max.Y || r.Max.Y < rect.Min.Y {
		return false
	}
	if r.Min.X > rect.Max.X || r.Max.X < rect.Min.X {
		return false
	}
	return true
}

// InsideRect detects rect is inside of another rect
func (r Rect) InsideRect(rect Rect) bool {
	if r.Min.X < rect.Min.X || r.Max.X > rect.Max.X {
		return false
	}
	if r.Min.Y < rect.Min.Y || r.Max.Y > rect.Max.Y {
		return false
	}
	return true
}
