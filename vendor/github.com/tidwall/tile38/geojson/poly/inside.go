package poly

// Inside returns true if point is inside of exterior and not in a hole.
// The validity of the exterior and holes must be done elsewhere and are assumed valid.
//   A valid exterior is a near-linear ring.
//   A valid hole is one that is full contained inside the exterior.
//   A valid hole may not share the same segment line as the exterior.
func (p Point) Inside(exterior Polygon, holes []Polygon) bool {
	if !insideshpext(p, exterior, true) {
		return false
	}
	for i := 0; i < len(holes); i++ {
		if insideshpext(p, holes[i], false) {
			return false
		}
	}
	return true
}

// Inside returns true if shape is inside of exterior and not in a hole.
func (shape Polygon) Inside(exterior Polygon, holes []Polygon) bool {
	var ok bool
	for _, p := range shape {
		ok = p.Inside(exterior, holes)
		if !ok {
			return false
		}
	}
	ok = true
	for _, hole := range holes {
		if hole.Inside(shape, nil) {
			return false
		}
	}
	return ok
}

func insideshpext(p Point, shape Polygon, exterior bool) bool {
	// if len(shape) < 3 {
	// 	return false
	// }
	in := false
	for i := 0; i < len(shape); i++ {
		res := raycast(p, shape[i], shape[(i+1)%len(shape)])
		if res.on {
			return exterior
		}
		if res.in {
			in = !in
		}
	}
	return in
}
