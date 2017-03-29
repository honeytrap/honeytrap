package geojson

import "testing"

func testJSONPoint(t *testing.T, js string) Point {
	g := testJSON(t, js)
	switch v := g.(type) {
	case Point:
		return v
	case SimplePoint:
		return Point{Coordinates: Position{X: v.X, Y: v.Y, Z: 0}}
	}
	t.Fatalf("not a point: %v", g)
	return Point{}
}
func testConvertToPoint(g Object) Point {
	switch v := g.(type) {
	default:
		panic("not a point")
	case Point:
		return v
	case SimplePoint:
		return Point{Coordinates: Position{X: v.X, Y: v.Y, Z: 0}}
	}
}
func TestPointJSON(t *testing.T) {
	testJSON(t, `{"type":"Point","coordinates":[100.1,5.1],"bbox":[0.1,0.1,100.1,100.1]}`)
	testJSON(t, `{"type":"Point","coordinates":[100.1,5.1]}`)
	testJSON(t, `{"type":"Point","coordinates":[100.1,5.1,10.5],"bbox":[0.1,0.1,20,100.1,100.1,30]}`)
	testJSON(t, `{"type":"Point","coordinates":[100.1,5.1,10.5]}`)
}
func TestPointCreation2D(t *testing.T) {
	p := P(100.5, 200.1)
	g1 := Point{Coordinates: p}
	jstr := g1.JSON()
	g2, err := ObjectJSON(jstr)
	if err != nil {
		t.Fatal(err)
	}
	jstr2 := g2.JSON()
	if jstr2 != jstr {
		t.Fatalf("%v != %v", jstr2, jstr)
	}
	if testConvertToPoint(g2).Coordinates != p {
		t.Fatalf("%v != %v", testConvertToPoint(g2).Coordinates, p)
	}
}
func TestPointCreation3D(t *testing.T) {
	p := P3(100.5, 200.1, 1029.3)
	g1 := Point{Coordinates: p}
	jstr := g1.JSON()
	g2, err := ObjectJSON(jstr)
	if err != nil {
		t.Fatal(err)
	}
	jstr2 := g2.JSON()
	if jstr2 != jstr {
		t.Fatalf("%v != %v", jstr2, jstr)
	}
	if testConvertToPoint(g2).Coordinates != p {
		t.Fatalf("%v != %v", testConvertToPoint(g2).Coordinates, p)
	}
}
func TestPointWithinBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[0,0,100,100]}`)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[10,10]}`)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[-10,-10,100,100]}`)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[-10,-10]}`)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
}
func TestPointIntersectsBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[0,0,100,100]}`)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[10,10]}`)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[-10,-10,100,100]}`)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[-10,-10,0,0]}`)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[-10,-10,-1,-1]}`)
	if p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSONPoint(t, `{"type":"Point","coordinates":[-10,-10]}`)
	if p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}

}

func TestPointWithinObject(t *testing.T) {
	p := testJSONPoint(t, `{"type":"Point","coordinates":[10,10]}`)
	if p.Within(testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[1,1,2,2]}`)) {
		t.Fatal("!")
	}
	if !p.Within(testJSONPoint(t, `{"type":"Point","coordinates":[10,10],"bbox":[0,0,100,100]}`)) {
		t.Fatal("!")
	}
	poly := testJSON(t, testPolyHoles)
	ps := []Position{P(.5, 3), P(3.5, .5), P(6, 0), P(11, -1), P(11.5, -4.5)}
	expect := true
	for _, p := range ps {
		got := tPoint(p.X, p.Y).Within(poly)
		if got != expect {
			t.Fatalf("%v within = %t, expect %t", p, got, expect)
		}
	}
	ps = []Position{P(-2, 0), P(0, -2), P(1.5, 1.5), P(8, 1), P(10.5, -1.5), P(14, -1), P(8, -3)}
	expect = false
	for _, p := range ps {
		got := tPoint(p.X, p.Y).Within(poly)
		if got != expect {
			t.Fatalf("%v within = %t, expect %t", p, got, expect)
		}
	}

}
