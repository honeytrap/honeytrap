package geojson

import (
	"github.com/tidwall/tile38/geojson/geohash"
	"github.com/tidwall/tile38/geojson/poly"
)

// MultiLineString is a geojson object with the type "MultiLineString"
type MultiLineString struct {
	Coordinates [][]Position
	BBox        *BBox
}

func fillMultiLineString(coordinates [][]Position, bbox *BBox, err error) (MultiLineString, error) {
	if err == nil {
		for _, coordinates := range coordinates {
			if len(coordinates) < 2 {
				err = errLineStringInvalidCoordinates
				break
			}
		}
	}
	return MultiLineString{
		Coordinates: coordinates,
		BBox:        bbox,
	}, err
}

// CalculatedBBox is exterior bbox containing the object.
func (g MultiLineString) CalculatedBBox() BBox {
	return level3CalculatedBBox(g.Coordinates, g.BBox, false)
}

// CalculatedPoint is a point representation of the object.
func (g MultiLineString) CalculatedPoint() Position {
	return g.CalculatedBBox().center()
}

// Geohash converts the object to a geohash value.
func (g MultiLineString) Geohash(precision int) (string, error) {
	p := g.CalculatedPoint()
	return geohash.Encode(p.Y, p.X, precision)
}

// PositionCount return the number of coordinates.
func (g MultiLineString) PositionCount() int {
	return level3PositionCount(g.Coordinates, g.BBox)
}

// Weight returns the in-memory size of the object.
func (g MultiLineString) Weight() int {
	return level3Weight(g.Coordinates, g.BBox)
}

// MarshalJSON allows the object to be encoded in json.Marshal calls.
func (g MultiLineString) MarshalJSON() ([]byte, error) {
	return []byte(g.JSON()), nil
}

// JSON is the json representation of the object. This might not be exactly the same as the original.
func (g MultiLineString) JSON() string {
	return level3JSON("MultiLineString", g.Coordinates, g.BBox)
}

// String returns a string representation of the object. This might be JSON or something else.
func (g MultiLineString) String() string {
	return g.JSON()
}

func (g MultiLineString) bboxPtr() *BBox {
	return g.BBox
}
func (g MultiLineString) hasPositions() bool {
	if g.BBox != nil {
		return true
	}
	for _, c := range g.Coordinates {
		if len(c) > 0 {
			return true
		}
	}
	return false
}

// WithinBBox detects if the object is fully contained inside a bbox.
func (g MultiLineString) WithinBBox(bbox BBox) bool {
	if g.BBox != nil {
		return rectBBox(g.CalculatedBBox()).InsideRect(rectBBox(bbox))
	}
	if len(g.Coordinates) == 0 {
		return false
	}
	for _, ls := range g.Coordinates {
		if len(ls) == 0 {
			return false
		}
		for _, p := range ls {
			if !poly.Point(p).InsideRect(rectBBox(bbox)) {
				return false
			}
		}
	}
	return true
}

// IntersectsBBox detects if the object intersects a bbox.
func (g MultiLineString) IntersectsBBox(bbox BBox) bool {
	if g.BBox != nil {
		return rectBBox(g.CalculatedBBox()).IntersectsRect(rectBBox(bbox))
	}
	for _, ls := range g.Coordinates {
		if polyPositions(ls).IntersectsRect(rectBBox(bbox)) {
			return true
		}
	}
	return false
}

// Within detects if the object is fully contained inside another object.
func (g MultiLineString) Within(o Object) bool {
	return withinObjectShared(g, o,
		func(v Polygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, ls := range g.Coordinates {
				if !polyPositions(ls).Inside(polyExteriorHoles(v.Coordinates)) {
					return false
				}
			}
			return true
		},
		func(v MultiPolygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, ls := range g.Coordinates {
				for _, c := range v.Coordinates {
					if !polyPositions(ls).Inside(polyExteriorHoles(c)) {
						return false
					}
				}
			}
			return true
		},
	)
}

// Intersects detects if the object intersects another object.
func (g MultiLineString) Intersects(o Object) bool {
	return intersectsObjectShared(g, o,
		func(v Polygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, ls := range g.Coordinates {
				if polyPositions(ls).Intersects(polyExteriorHoles(v.Coordinates)) {
					return true
				}
			}
			return false
		},
		func(v MultiPolygon) bool {
			if len(g.Coordinates) == 0 {
				return false
			}
			for _, ls := range g.Coordinates {
				for _, c := range v.Coordinates {
					if polyPositions(ls).Intersects(polyExteriorHoles(c)) {
						return true
					}
				}
			}
			return false
		},
	)
}

// Nearby detects if the object is nearby a position.
func (g MultiLineString) Nearby(center Position, meters float64) bool {
	return nearbyObjectShared(g, center.X, center.Y, meters)
}

// IsBBoxDefined returns true if the object has a defined bbox.
func (g MultiLineString) IsBBoxDefined() bool {
	return g.BBox != nil
}

// IsGeometry return true if the object is a geojson geometry object. false if it something else.
func (g MultiLineString) IsGeometry() bool {
	return true
}
