package geojson

import (
	"github.com/tidwall/tile38/geojson/geo"
	"github.com/tidwall/tile38/geojson/geohash"
	"github.com/tidwall/tile38/geojson/poly"
)

// SimplePoint is a geojson object with the type "Point" and where there coordinate is 2D and there is no bbox.
type SimplePoint struct {
	X, Y float64
}

// New2DPoint creates a SimplePoint
func New2DPoint(x, y float64) SimplePoint {
	return SimplePoint{x, y}
}

func fillSimplePoint(coordinates Position, bbox *BBox, err error) (SimplePoint, error) {
	return SimplePoint{X: coordinates.X, Y: coordinates.Y}, err
}

// CalculatedBBox is exterior bbox containing the object.
func (g SimplePoint) CalculatedBBox() BBox {
	return BBox{
		Min: Position{X: g.X, Y: g.Y, Z: 0},
		Max: Position{X: g.X, Y: g.Y, Z: 0},
	}
}

// CalculatedPoint is a point representation of the object.
func (g SimplePoint) CalculatedPoint() Position {
	return Position{X: g.X, Y: g.Y, Z: 0}
}

// Geohash converts the object to a geohash value.
func (g SimplePoint) Geohash(precision int) (string, error) {
	p := g.CalculatedPoint()
	return geohash.Encode(p.Y, p.X, precision)
}

// PositionCount return the number of coordinates.
func (g SimplePoint) PositionCount() int {
	return 1
}

// Weight returns the in-memory size of the object.
func (g SimplePoint) Weight() int {
	return 2 * 8
}

// MarshalJSON allows the object to be encoded in json.Marshal calls.
func (g SimplePoint) MarshalJSON() ([]byte, error) {
	return []byte(g.JSON()), nil
}

// JSON is the json representation of the object. This might not be exactly the same as the original.
func (g SimplePoint) JSON() string {
	return level1JSON("Point", Position{X: g.X, Y: g.Y, Z: 0}, nil)
}

// String returns a string representation of the object. This might be JSON or something else.
func (g SimplePoint) String() string {
	return g.JSON()
}

func (g SimplePoint) bboxPtr() *BBox {
	return nil
}
func (g SimplePoint) hasPositions() bool {
	return true
}

// WithinBBox detects if the object is fully contained inside a bbox.
func (g SimplePoint) WithinBBox(bbox BBox) bool {
	return poly.Point(Position{X: g.X, Y: g.Y, Z: 0}).InsideRect(rectBBox(bbox))
}

// IntersectsBBox detects if the object intersects a bbox.
func (g SimplePoint) IntersectsBBox(bbox BBox) bool {
	return poly.Point(Position{X: g.X, Y: g.Y, Z: 0}).InsideRect(rectBBox(bbox))
}

// Within detects if the object is fully contained inside another object.
func (g SimplePoint) Within(o Object) bool {
	return withinObjectShared(g, o,
		func(v Polygon) bool {
			return poly.Point(Position{X: g.X, Y: g.Y, Z: 0}).Inside(polyExteriorHoles(v.Coordinates))
		},
		func(v MultiPolygon) bool {
			for _, c := range v.Coordinates {
				if !poly.Point(Position{X: g.X, Y: g.Y, Z: 0}).Inside(polyExteriorHoles(c)) {
					return false
				}
			}
			return true
		},
	)
}

// Intersects detects if the object intersects another object.
func (g SimplePoint) Intersects(o Object) bool {
	return intersectsObjectShared(g, o,
		func(v Polygon) bool {
			return poly.Point(Position{X: g.X, Y: g.Y, Z: 0}).Intersects(polyExteriorHoles(v.Coordinates))
		},
		func(v MultiPolygon) bool {
			for _, c := range v.Coordinates {
				if poly.Point(Position{X: g.X, Y: g.Y, Z: 0}).Intersects(polyExteriorHoles(c)) {
					return true
				}
			}
			return false
		},
	)
}

// Nearby detects if the object is nearby a position.
func (g SimplePoint) Nearby(center Position, meters float64) bool {
	return geo.DistanceTo(center.Y, center.X, g.Y, g.X) <= meters
}

// IsBBoxDefined returns true if the object has a defined bbox.
func (g SimplePoint) IsBBoxDefined() bool {
	return false
}

// IsGeometry return true if the object is a geojson geometry object. false if it something else.
func (g SimplePoint) IsGeometry() bool {
	return true
}
