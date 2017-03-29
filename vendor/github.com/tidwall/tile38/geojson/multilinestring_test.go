package geojson

import "testing"

func TestMultiLineString(t *testing.T) {
	testJSON(t, `{"type":"MultiLineString","coordinates":[[[100.1,5.1],[101.1,6.1]],[[102.1,7.1],[103.1,8.1]]]}`)
	testJSON(t, `{
    "type": "MultiLineString",
    "coordinates": [
        [
            [-105.0214433670044,39.57805759162015],
            [-105.02150774002075,39.57780951131517],
            [-105.02157211303711,39.57749527498758],
            [-105.02157211303711,39.57716449836683],
            [-105.02157211303711,39.57703218727656],
            [-105.02152919769287,39.57678410330158]
        ],
        [
            [-105.01989841461182,39.574997872470774],
            [-105.01959800720215,39.57489863607502],
            [-105.01906156539916,39.57478286010041]
        ],
        [
            [-105.01717329025269,39.5744024519653],
            [-105.01698017120361,39.574385912433804],
            [-105.0166368484497,39.574385912433804],
            [-105.01650810241699,39.5744024519653],
            [-105.0159502029419,39.574270135602866]
        ],
        [
            [-105.0142765045166,39.57397242286402],
            [-105.01412630081175,39.57403858136094],
            [-105.0138258934021,39.57417089816531],
            [-105.01331090927124,39.57445207053608]
        ]
    ]
}`)
}

func TestMultiLineStringWithinBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSON(t, `{"type":"MultiLineString","coordinates":[[[10,10],[20,20]]],"bbox":[0,0,100,100]}`).(MultiLineString)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiLineString","coordinates":[[[10,10],[20,20]],[[30,30],[40,40]]]}`).(MultiLineString)
	if !p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiLineString","coordinates":[[[10,10],[20,20]]],"bbox":[-10,-10,100,100]}`).(MultiLineString)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiLineString","coordinates":[[[-10,-10],[-20,-20]]]}`).(MultiLineString)
	if p.WithinBBox(bbox) {
		t.Fatal("!")
	}
}

func TestMultiLineStringIntersectsBBox(t *testing.T) {
	bbox := BBox{Min: Position{0, 0, 0}, Max: Position{100, 100, 0}}
	p := testJSON(t, `{"type":"MultiLineString","coordinates":[[[10,10],[20,20]]],"bbox":[0,0,100,100]}`).(MultiLineString)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiLineString","coordinates":[[[-1,3],[3,-1]],[[-1000,-1000],[-1020,-1020]]]}`).(MultiLineString)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiLineString","coordinates":[[[-1,1],[1,-1]]]}`).(MultiLineString)
	if !p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
	p = testJSON(t, `{"type":"MultiLineString","coordinates":[[[-2,1],[1,-1]]]}`).(MultiLineString)
	if p.IntersectsBBox(bbox) {
		t.Fatal("!")
	}
}
