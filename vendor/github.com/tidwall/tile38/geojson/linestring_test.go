package geojson

import "testing"

func TestLineString(t *testing.T) {
	testJSON(t, `{"type":"LineString","coordinates":[[100.1,5.1],[101.1,51.1]]}`)
	testJSON(t, `{"type":"LineString","coordinates":[[100.1,5.1],[101.1,51.1]],"bbox":[10,20,30,40]}`)
	testJSON(t, `{"type":"LineString","coordinates":[[100.1,5.1,15.5],[101.1,51.1,20],[10001.1,71.1,10]],"bbox":[10,20,12,30,40,15]}`)
	testJSON(t, `{
    "type": "LineString",
    "coordinates": [
        [-101.744384765625,39.32155002466662],
        [-101.5521240234375,39.330048552942415],
        [-101.40380859375,39.330048552942415],
        [-101.33239746093749,39.364032338047984],
        [-101.041259765625,39.36827914916011],
        [-100.975341796875,39.30454987014581],
        [-100.9149169921875,39.24501680713314],
        [-100.843505859375,39.16414104768742],
        [-100.8050537109375,39.104488809440475],
        [-100.491943359375,39.10022600175347],
        [-100.43701171875,39.095962936305476],
        [-100.338134765625,39.095962936305476],
        [-100.1953125,39.027718840211605],
        [-100.008544921875,39.01064750994083],
        [-99.86572265625,39.00211029922512],
        [-99.6844482421875,38.97222194853654],
        [-99.51416015625,38.929502416386605],
        [-99.38232421875,38.92095542046727],
        [-99.3218994140625,38.89530825492018],
        [-99.1131591796875,38.86965182408357],
        [-99.0802001953125,38.85682013474361],
        [-98.82202148437499,38.85682013474361],
        [-98.44848632812499,38.84826438869913],
        [-98.20678710937499,38.84826438869913],
        [-98.02001953125,38.8782049970615],
        [-97.635498046875,38.87392853923629]
    ]
}`)
}

func TestLineStringWithinBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSON(t, `{"type":"LineString","coordinates":[[10,10],[20,20]],"bbox":[0,0,100,100]}`).(LineString)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"LineString","coordinates":[[10,10],[20,20]]}`).(LineString)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"LineString","coordinates":[[10,10],[20,20]],"bbox":[-10,-10,100,100]}`).(LineString)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"LineString","coordinates":[[-10,-10],[-20,-20]]}`).(LineString)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
}

func TestLineStringIntersectsBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSON(t, `{"type":"LineString","coordinates":[[10,10],[20,20]],"bbox":[0,0,100,100]}`).(LineString)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"LineString","coordinates":[[-1,3],[3,-1]]}`).(LineString)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"LineString","coordinates":[[-1,1],[1,-1]]}`).(LineString)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"LineString","coordinates":[[-2,1],[1,-1]]}`).(LineString)
	if p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
}
