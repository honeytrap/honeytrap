package geojson

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func P(x, y float64) Position {
	return Position{x, y, 0}
}

func P3(x, y, z float64) Position {
	return Position{x, y, z}
}

const testPolyHoles = `
{"type":"Polygon","coordinates":[
[[0,0],[0,6],[12,-6],[12,0],[0,0]],
[[1,1],[1,2],[2,2],[2,1],[1,1]],
[[11,-1],[11,-3],[9,-1],[11,-1]]
]}`

func tPoint(x, y float64) Point {
	o, err := ObjectJSON(fmt.Sprintf(`{"type":"Point","coordinates":[%f,%f]}`, x, y))
	if err != nil {
		log.Fatal(err)
	}
	return testConvertToPoint(o)
}

func doesJSONMatch(js1, js2 string) bool {
	var m1, m2 map[string]interface{}
	json.Unmarshal([]byte(js1), &m1)
	json.Unmarshal([]byte(js2), &m2)
	b1, _ := json.Marshal(m1)
	b2, _ := json.Marshal(m2)
	return string(b1) == string(b2)
}

func testJSON(t *testing.T, jstr string) Object {
	o, err := ObjectJSON(jstr)
	if err != nil {
		t.Fatal(err)
	}
	if !doesJSONMatch(o.JSON(), jstr) {
		t.Fatalf("expected '%v', got '%v'", o.JSON(), jstr)
	}

	return o
}

func testInvalidJSON(t *testing.T, js string, expecting error) {
	_, err := ObjectJSON(js)
	if err == nil {
		t.Fatalf("expecting an error for json '%s'", js)
	}
	if err.Error() != expecting.Error() {
		t.Fatalf("\nInvalid error for json:\n'%s'\ngot '%s'\nexpected '%s'", js, err.Error(), expecting.Error())
	}
}

func TestInvalidJSON(t *testing.T) {
	testInvalidJSON(t, `{}`, errInvalidTypeMember)
	testInvalidJSON(t, `{"type":"Poin"}`, fmt.Errorf(fmtErrTypeIsUnknown, "Poin"))
	testInvalidJSON(t, `{"type":"Point"}`, errCoordinatesRequired)
	testInvalidJSON(t, `{"type":"Point","coordinates":[]}`, errInvalidNumberOfPositionValues)
	testInvalidJSON(t, `{"type":"Point","coordinates":[1]}`, errInvalidNumberOfPositionValues)
	testInvalidJSON(t, `{"type":"Point","coordinates":[1,2,"asd"]}`, errInvalidPositionValue)
	testInvalidJSON(t, `{"type":"Point","coordinates":[[1,2]]}`, errInvalidPositionValue)
	testInvalidJSON(t, `{"type":"Point","coordinates":[[1,2],[1,3]]}`, errInvalidPositionValue)
	testInvalidJSON(t, `{"type":"MultiPoint","coordinates":[1,2]}`, errInvalidCoordinates)
	testInvalidJSON(t, `{"type":"MultiPoint","coordinates":[[]]}`, errInvalidNumberOfPositionValues)
	testInvalidJSON(t, `{"type":"MultiPoint","coordinates":[[1]]}`, errInvalidNumberOfPositionValues)
	testInvalidJSON(t, `{"type":"MultiPoint","coordinates":[[1,2,"asd"]]}`, errInvalidPositionValue)
	testInvalidJSON(t, `{"type":"LineString","coordinates":[]}`, errLineStringInvalidCoordinates)
	testInvalidJSON(t, `{"type":"MultiLineString","coordinates":[[]]}`, errLineStringInvalidCoordinates)
	testInvalidJSON(t, `{"type":"MultiLineString","coordinates":[[[]]]}`, errInvalidNumberOfPositionValues)
	testInvalidJSON(t, `{"type":"MultiLineString","coordinates":[[[]]]}`, errInvalidNumberOfPositionValues)
	testInvalidJSON(t, `{"type":"Polygon","coordinates":[[1,1],[2,2],[3,3],[4,4]]}`, errInvalidCoordinates)
	testInvalidJSON(t, `{"type":"Polygon","coordinates":[[[1,1],[2,2],[3,3],[4,4]]]}`, errMustBeALinearRing)
	testInvalidJSON(t, `{"type":"Polygon","coordinates":[[[1,1],[2,2],[3,3],[1,1]],[[1,1],[2,2],[3,3],[4,4]]]}`, errMustBeALinearRing)
	testInvalidJSON(t, `{"type":"Point","coordinates":[1,2,3],"bbox":123}`, errBBoxInvalidType)
	testInvalidJSON(t, `{"type":"Point","coordinates":[1,2,3],"bbox":[]}`, errBBoxInvalidNumberOfValues)
	testInvalidJSON(t, `{"type":"Point","coordinates":[1,2,3],"bbox":[1,2,3]}`, errBBoxInvalidNumberOfValues)
	testInvalidJSON(t, `{"type":"Point","coordinates":[1,2,3],"bbox":[1,2,3,"a"]}`, errBBoxInvalidValue)
}

// func TestJunk(t *testing.T) {
// 	type ThreePoint struct{ X, Y, Z float64 }

// 	var s1 = ThreePoint{50, 50, 50}
// 	var s2 = &s1
// 	var o1 interface{} = s1
// 	var o2 interface{} = s2

// 	t1 := reflect.TypeOf(s1)
// 	t2 := reflect.TypeOf(s2)
// 	t3 := reflect.TypeOf(o1)
// 	t4 := reflect.TypeOf(o2)

// 	fmt.Printf("typeof: %s\n", t1)
// 	fmt.Printf("typeof: %s\n", t2)
// 	fmt.Printf("typeof: %s\n", t3)
// 	fmt.Printf("typeof: %s\n", t4)

// 	z1 := unsafe.Sizeof(s1)
// 	z2 := unsafe.Sizeof(s2)
// 	z3 := unsafe.Sizeof(o1)
// 	z4 := unsafe.Sizeof(o2.(*ThreePoint))

// 	fmt.Printf("sizeof: %d\n", z1)
// 	fmt.Printf("sizeof: %d\n", z2)
// 	fmt.Printf("sizeof: %d\n", z3)
// 	fmt.Printf("sizeof: %d\n", z4)

// }
