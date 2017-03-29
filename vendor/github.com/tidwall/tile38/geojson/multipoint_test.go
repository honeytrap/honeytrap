package geojson

import "testing"

func TestMultiPointJSON(t *testing.T) {
	testJSON(t, `{"type":"MultiPoint","coordinates":[[100.1,5.1,10],[101.1,51.1,10]],"bbox":[0.1,0.1,15.1,100.1,100.1,19.1]}`)
	testJSON(t, `{"type":"MultiPoint","coordinates":[[100.1,5.1],[101.1,51.1]]}`)
	testJSON(t, `{"type":"MultiPoint","coordinates":[[100.1,5.1],[101.1,51.1]],"bbox":[0.1,0.1,100.1,100.1]}`)
	testJSON(t, `{"type":"MultiPoint","coordinates":[[100.1,5.1,20],[101.1,51.1,50]]}`)
}
func TestMultiPointWithinBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]],"bbox":[0,0,100,100]}`).(MultiPoint)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]]}`).(MultiPoint)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]],"bbox":[-10,-10,100,100]}`).(MultiPoint)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[-10,-10],[-20,-20]]}`).(MultiPoint)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
}
func TestMultiPointIntersectsBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]],"bbox":[0,0,100,100]}`).(MultiPoint)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]]}`).(MultiPoint)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]],"bbox":[-10,-10,100,100]}`).(MultiPoint)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]],"bbox":[-10,-10,0,0]}`).(MultiPoint)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[20,20]],"bbox":[-10,-10,-1,-1]}`).(MultiPoint)
	if p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[-10,-10],[-20,-20]]}`).(MultiPoint)
	if p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiPoint","coordinates":[[10,10],[-20,-20]]}`).(MultiPoint)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}

}
