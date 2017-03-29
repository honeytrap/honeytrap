package geojson

import (
	"bytes"

	"github.com/tidwall/gjson"
	"github.com/tidwall/tile38/geojson/geohash"
)

// GeometryCollection is a geojson object with the type "GeometryCollection"
type GeometryCollection struct {
	Geometries []Object
	BBox       *BBox
}

func fillGeometryCollectionMap(json string) (GeometryCollection, error) {
	var g GeometryCollection
	res := gjson.Get(json, "geometries")
	switch res.Type {
	default:
		return g, errInvalidGeometries
	case gjson.Null:
		return g, errGeometriesRequired
	case gjson.JSON:
		if !resIsArray(res) {
			return g, errInvalidGeometries
		}
		v := res.Array()
		g.Geometries = make([]Object, len(v))
		for i, res := range v {
			if res.Type != gjson.JSON {
				return g, errInvalidGeometry
			}
			o, err := objectMap(res.Raw, gcoll)
			if err != nil {
				return g, err
			}
			g.Geometries[i] = o
		}
	}
	var err error
	g.BBox, err = fillBBox(json)
	return g, err
}

// CalculatedBBox is exterior bbox containing the object.
func (g GeometryCollection) CalculatedBBox() BBox {
	if g.BBox != nil {
		return *g.BBox
	}
	var bbox BBox
	for i, g := range g.Geometries {
		if i == 0 {
			bbox = g.CalculatedBBox()
		} else {
			bbox = bbox.union(g.CalculatedBBox())
		}
	}
	return bbox
}

// CalculatedPoint is a point representation of the object.
func (g GeometryCollection) CalculatedPoint() Position {
	return g.CalculatedBBox().center()
}

// Geohash converts the object to a geohash value.
func (g GeometryCollection) Geohash(precision int) (string, error) {
	p := g.CalculatedPoint()
	return geohash.Encode(p.Y, p.X, precision)
}

// PositionCount return the number of coordinates.
func (g GeometryCollection) PositionCount() int {
	var res int
	for _, g := range g.Geometries {
		res += g.PositionCount()
	}
	if g.BBox != nil {
		return 2 + res
	}
	return res
}

// Weight returns the in-memory size of the object.
func (g GeometryCollection) Weight() int {
	var res int
	for _, g := range g.Geometries {
		res += g.Weight()
	}
	return res
}

// MarshalJSON allows the object to be encoded in json.Marshal calls.
func (g GeometryCollection) MarshalJSON() ([]byte, error) {
	return []byte(g.JSON()), nil
}

// JSON is the json representation of the object. This might not be exactly the same as the original.
func (g GeometryCollection) JSON() string {
	var buf bytes.Buffer
	buf.WriteString(`{"type":"GeometryCollection","geometries":[`)
	for i, g := range g.Geometries {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(g.JSON())
	}
	buf.WriteByte(']')
	g.BBox.write(&buf)
	buf.WriteByte('}')
	return buf.String()
}

// String returns a string representation of the object. This might be JSON or something else.
func (g GeometryCollection) String() string {
	return g.JSON()
}

// Bytes is the bytes representation of the object.
func (g GeometryCollection) Bytes() []byte {
	return []byte(g.JSON())
}
func (g GeometryCollection) bboxPtr() *BBox {
	return g.BBox
}
func (g GeometryCollection) hasPositions() bool {
	if g.BBox != nil {
		return true
	}
	for _, g := range g.Geometries {
		if g.hasPositions() {
			return true
		}
	}
	return false
}

// WithinBBox detects if the object is fully contained inside a bbox.
func (g GeometryCollection) WithinBBox(bbox BBox) bool {
	if g.BBox != nil {
		return rectBBox(g.CalculatedBBox()).InsideRect(rectBBox(bbox))
	}
	if len(g.Geometries) == 0 {
		return false
	}
	for _, g := range g.Geometries {
		if !g.WithinBBox(bbox) {
			return false
		}
	}
	return true
}

// IntersectsBBox detects if the object intersects a bbox.
func (g GeometryCollection) IntersectsBBox(bbox BBox) bool {
	if g.BBox != nil {
		return rectBBox(g.CalculatedBBox()).IntersectsRect(rectBBox(bbox))
	}
	for _, g := range g.Geometries {
		if g.IntersectsBBox(bbox) {
			return false
		}
	}
	return true
}

// Within detects if the object is fully contained inside another object.
func (g GeometryCollection) Within(o Object) bool {
	return withinObjectShared(g, o,
		func(v Polygon) bool {
			if len(g.Geometries) == 0 {
				return false
			}
			for _, g := range g.Geometries {
				if !g.Within(o) {
					return false
				}
			}
			return true
		},
		func(v MultiPolygon) bool {
			if len(g.Geometries) == 0 {
				return false
			}
			for _, g := range g.Geometries {
				if !g.Within(o) {
					return false
				}
			}
			return true
		},
	)
}

// Intersects detects if the object intersects another object.
func (g GeometryCollection) Intersects(o Object) bool {
	return intersectsObjectShared(g, o,
		func(v Polygon) bool {
			if len(g.Geometries) == 0 {
				return false
			}
			for _, g := range g.Geometries {
				if g.Intersects(o) {
					return true
				}
			}
			return false
		},
		func(v MultiPolygon) bool {
			if len(g.Geometries) == 0 {
				return false
			}
			for _, g := range g.Geometries {
				if g.Intersects(o) {
					return true
				}
			}
			return false
		},
	)
}

// Nearby detects if the object is nearby a position.
func (g GeometryCollection) Nearby(center Position, meters float64) bool {
	return nearbyObjectShared(g, center.X, center.Y, meters)
}

// IsBBoxDefined returns true if the object has a defined bbox.
func (g GeometryCollection) IsBBoxDefined() bool {
	return g.BBox != nil
}

// IsGeometry return true if the object is a geojson geometry object. false if it something else.
func (g GeometryCollection) IsGeometry() bool {
	return true
}
