package geojson

import "testing"

func TestCirclePolygon(t *testing.T) {
	circle := CirclePolygon(-115, 33, 10000, 20)
	point := Point{Coordinates: Position{-115, 33, 0}}
	if !point.Intersects(circle) {
		t.Fatal("should intersect")
	}
	circle2 := CirclePolygon(-115, 33, 20000, 20)
	if !circle2.Intersects(circle) {
		t.Fatal("should intersect")
	}
	if !circle.Intersects(circle2) {
		t.Fatal("should intersect")
	}
	rect := Polygon{
		Coordinates: [][]Position{
			{
				{X: -120, Y: 20, Z: 0},
				{X: -120, Y: 40, Z: 0},
				{X: -100, Y: 40, Z: 0},
				{X: -100, Y: 40, Z: 0},
				{X: -120, Y: 20, Z: 0},
			},
		},
	}
	if !circle.Intersects(rect) {
		t.Fatal("should intersect")
	}
	if !rect.Intersects(circle) {
		t.Fatal("should intersect")
	}
	line := LineString{
		Coordinates: []Position{
			{X: -116, Y: 23, Z: 0},
			{X: -114, Y: 43, Z: 0},
		},
	}
	if !line.Intersects(circle) {
		t.Fatal("should intersect")
	}
}
