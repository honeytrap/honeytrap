package geojson

import "github.com/tidwall/tile38/geojson/geohash"

// MultiPolygon is a geojson object with the type "MultiPolygon"
type MultiPolygon struct {
	Coordinates [][][]Position
	BBox        *BBox
}

func fillMultiPolygon(coordinates [][][]Position, bbox *BBox, err error) (MultiPolygon, error) {
	if err == nil {
	outer:
		for _, ps := range coordinates {
			if len(ps) == 0 {
				err = errMustBeALinearRing
				break
			}
			for _, ps := range ps {
				if !isLinearRing(ps) {
					err = errMustBeALinearRing
					break outer
				}
			}
		}
	}
	return MultiPolygon{
		Coordinates: coordinates,
		BBox:        bbox,
	}, err
}

// CalculatedBBox is exterior bbox containing the object.
func (g MultiPolygon) CalculatedBBox() BBox {
	return level4CalculatedBBox(g.Coordinates, g.BBox)
}

// CalculatedPoint is a point representation of the object.
func (g MultiPolygon) CalculatedPoint() Position {
	return g.CalculatedBBox().center()
}

// Geohash converts the object to a geohash value.
func (g MultiPolygon) Geohash(precision int) (string, error) {
	p := g.CalculatedPoint()
	return geohash.Encode(p.Y, p.X, precision)
}

// PositionCount return the number of coordinates.
func (g MultiPolygon) PositionCount() int {
	return level4PositionCount(g.Coordinates, g.BBox)
}

// Weight returns the in-memory size of the object.
func (g MultiPolygon) Weight() int {
	return level4Weight(g.Coordinates, g.BBox)
}

// MarshalJSON allows the object to be encoded in json.Marshal calls.
func (g MultiPolygon) MarshalJSON() ([]byte, error) {
	return []byte(g.JSON()), nil
}

// JSON is the json representation of the object. This might not be exactly the same as the original.
func (g MultiPolygon) JSON() string {
	return level4JSON("MultiPolygon", g.Coordinates, g.BBox)
}

// String returns a string representation of the object. This might be JSON or something else.
func (g MultiPolygon) String() string {
	return g.JSON()
}

func (g MultiPolygon) bboxPtr() *BBox {
	return g.BBox
}
func (g MultiPolygon) hasPositions() bool {
	if g.BBox != nil {
		return true
	}
	for _, c := range g.Coordinates {
		for _, c := range c {
			if len(c) > 0 {
				return true
			}
		}
	}
	return false
}

// WithinBBox detects if the object is fully contained inside a bbox.
func (g MultiPolygon) WithinBBox(bbox BBox) bool {
	if g.BBox != nil {
		return rectBBox(g.CalculatedBBox()).InsideRect(rectBBox(bbox))
	}
	if len(g.Coordinates) == 0 {
		return false
	}
	for _, p := range g.Coordinates {
		if !(Polygon{Coordinates: p}).WithinBBox(bbox) {
			return false
		}
	}
	return true
}

// IntersectsBBox detects if the object intersects a bbox.
func (g MultiPolygon) IntersectsBBox(bbox BBox) bool {
	if g.BBox != nil {
		return rectBBox(g.CalculatedBBox()).IntersectsRect(rectBBox(bbox))
	}
	for _, p := range g.Coordinates {
		if (Polygon{Coordinates: p}).IntersectsBBox(bbox) {
			return true
		}
	}
	return false
}

// Within detects if the object is fully contained inside another object.
func (g MultiPolygon) Within(o Object) bool {
	return withinObjectShared(g, o,
		func(v Polygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, p := range g.Coordinates {
				if len(p) > 0 {
					if !polyPositions(p[0]).Inside(polyExteriorHoles(v.Coordinates)) {
						return false
					}
				}
			}
			return true
		},
		func(v MultiPolygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, p := range g.Coordinates {
				if len(p) > 0 {
					for _, c := range v.Coordinates {
						if !polyPositions(p[0]).Inside(polyExteriorHoles(c)) {
							return false
						}
					}
				}
			}
			return true
		},
	)
}

// Intersects detects if the object intersects another object.
func (g MultiPolygon) Intersects(o Object) bool {
	return intersectsObjectShared(g, o,
		func(v Polygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, p := range g.Coordinates {
				if len(p) > 0 {
					if polyPositions(p[0]).Intersects(polyExteriorHoles(v.Coordinates)) {
						return true
					}
				}
			}
			return false
		},
		func(v MultiPolygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, p := range g.Coordinates {
				if len(p) > 0 {
					for _, c := range v.Coordinates {
						if polyPositions(p[0]).Intersects(polyExteriorHoles(c)) {
							return true
						}
					}
				}
			}
			return false
		},
	)
}

// Nearby detects if the object is nearby a position.
func (g MultiPolygon) Nearby(center Position, meters float64) bool {
	return nearbyObjectShared(g, center.X, center.Y, meters)
}

// IsBBoxDefined returns true if the object has a defined bbox.
func (g MultiPolygon) IsBBoxDefined() bool {
	return g.BBox != nil
}

// IsGeometry return true if the object is a geojson geometry object. false if it something else.
func (g MultiPolygon) IsGeometry() bool {
	return true
}
