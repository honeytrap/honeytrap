package geojson

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"unsafe"

	"github.com/tidwall/gjson"
	"github.com/tidwall/tile38/geojson/geo"
	"github.com/tidwall/tile38/geojson/poly"
)

const sizeofPosition = 24 // (X,Y,Z) * 8

// Position is a simple point
type Position poly.Point

func pointPositions(positions []Position) []poly.Point {
	return *(*[]poly.Point)(unsafe.Pointer(&positions))
}
func polyPositions(positions []Position) poly.Polygon {
	return *(*poly.Polygon)(unsafe.Pointer(&positions))
}
func polyMultiPositions(positions [][]Position) []poly.Polygon {
	return *(*[]poly.Polygon)(unsafe.Pointer(&positions))
}
func polyExteriorHoles(positions [][]Position) (exterior poly.Polygon, holes []poly.Polygon) {
	switch len(positions) {
	case 0:
	case 1:
		exterior = polyPositions(positions[0])
	default:
		exterior = polyPositions(positions[0])
		holes = polyMultiPositions(positions[1:])
	}
	return
}

func (p Position) writeJSON(buf *bytes.Buffer, isCordZ bool) {
	buf.WriteString(strconv.FormatFloat(p.X, 'f', -1, 64))
	buf.WriteByte(',')
	buf.WriteString(strconv.FormatFloat(p.Y, 'f', -1, 64))
	if isCordZ {
		buf.WriteByte(',')
		buf.WriteString(strconv.FormatFloat(p.Z, 'f', -1, 64))
	}
}

func (p Position) writeBytes(buf *bytes.Buffer, isCordZ bool) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(p.X))
	buf.Write(b)
	binary.LittleEndian.PutUint64(b, math.Float64bits(p.Y))
	buf.Write(b)
	if isCordZ {
		binary.LittleEndian.PutUint64(b, math.Float64bits(p.Z))
		buf.Write(b)
	}
}

const earthRadius = 6371e3

func toRadians(deg float64) float64 { return deg * math.Pi / 180 }
func toDegrees(rad float64) float64 { return rad * 180 / math.Pi }

// DistanceTo calculates the distance to a position
func (p Position) DistanceTo(position Position) float64 {
	return geo.DistanceTo(p.Y, p.X, position.Y, position.X)
}

// Destination calculates a new position based on the distance and bearing.
func (p Position) Destination(meters, bearingDegrees float64) Position {
	lat, lon := geo.DestinationPoint(p.Y, p.X, meters, bearingDegrees)
	return Position{X: lon, Y: lat, Z: 0}
}

func fillPosition(coords gjson.Result) (Position, error) {
	var p Position
	v := coords.Array()
	switch len(v) {
	case 0:
		return p, errInvalidNumberOfPositionValues
	case 1:
		if v[0].Type != gjson.Number {
			return p, errInvalidPositionValue
		}
		return p, errInvalidNumberOfPositionValues
	}
	for i := 0; i < len(v); i++ {
		if v[i].Type != gjson.Number {
			return p, errInvalidPositionValue
		}
	}
	p.X = v[0].Float()
	p.Y = v[1].Float()
	if len(v) > 2 {
		p.Z = v[2].Float()
	} else {
		p.Z = nilz
	}
	return p, nil
}

func fillPositionBytes(b []byte, isCordZ bool) (Position, []byte, error) {
	var p Position
	if len(b) < 8 {
		return p, b, errNotEnoughData
	}
	p.X = math.Float64frombits(binary.LittleEndian.Uint64(b))
	b = b[8:]
	if len(b) < 8 {
		return p, b, errNotEnoughData
	}
	p.Y = math.Float64frombits(binary.LittleEndian.Uint64(b))
	b = b[8:]
	if isCordZ {
		if len(b) < 8 {
			return p, b, errNotEnoughData
		}
		p.Z = math.Float64frombits(binary.LittleEndian.Uint64(b))
		b = b[8:]
	} else {
		p.Z = nilz
	}
	return p, b, nil
}

// ExternalJSON is the simple json representation of the position used for external applications.
func (p Position) ExternalJSON() string {
	if p.Z != 0 {
		return `{"lat":` + strconv.FormatFloat(p.Y, 'f', -1, 64) + `,"lon":` + strconv.FormatFloat(p.X, 'f', -1, 64) + `,"z":` + strconv.FormatFloat(p.Z, 'f', -1, 64) + `}`
	}
	return `{"lat":` + strconv.FormatFloat(p.Y, 'f', -1, 64) + `,"lon":` + strconv.FormatFloat(p.X, 'f', -1, 64) + `}`
}
