package collection

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/tidwall/tile38/geojson"
)

func TestCollection(t *testing.T) {
	const numItems = 10000
	objs := make(map[string]geojson.Object)
	c := New()
	for i := 0; i < numItems; i++ {
		id := strconv.FormatInt(int64(i), 10)
		var obj geojson.Object
		p := geojson.Position{X: rand.Float64()*360 - 180, Y: rand.Float64()*180 - 90, Z: 0}
		if rand.Int()%2 == 0 {
			obj = geojson.Point{Coordinates: p}
		} else {
			minX, minY := rand.Float64()*360-180, rand.Float64()*180-90
			obj = geojson.Point{Coordinates: p, BBox: &geojson.BBox{
				Min: geojson.Position{X: minX, Y: minY, Z: 0},
				Max: geojson.Position{X: minX + 100, Y: minY + 100, Z: 0},
			}}
		}
		objs[id] = obj
		c.ReplaceOrInsert(id, obj, nil, nil)
	}
	count := 0
	bbox := geojson.BBox{Min: geojson.Position{X: -180, Y: -90, Z: 0}, Max: geojson.Position{X: 180, Y: 90, Z: 0}}
	c.geoSearch(0, bbox, func(id string, obj geojson.Object, field []float64) bool {
		count++
		return true
	})
	if count != len(objs) {
		t.Fatalf("count = %d, expect %d", count, len(objs))
	}
	count = c.Count()
	if count != len(objs) {
		t.Fatalf("c.Count() = %d, expect %d", count, len(objs))
	}
	testCollectionVerifyContents(t, c, objs)
}

func testCollectionVerifyContents(t *testing.T, c *Collection, objs map[string]geojson.Object) {
	for id, o2 := range objs {
		o1, _, ok := c.Get(id)
		if !ok {
			t.Fatalf("ok[%s] = false, expect true", id)
		}
		j1 := o1.JSON()
		j2 := o2.JSON()
		if j1 != j2 {
			t.Fatalf("j1 == %s, expect %s", j1, j2)
		}
	}
}

func TestManyCollections(t *testing.T) {
	colsM := make(map[string]*Collection)
	cols := 100
	objs := 1000
	k := 0
	for i := 0; i < cols; i++ {
		key := strconv.FormatInt(int64(i), 10)
		for j := 0; j < objs; j++ {
			id := strconv.FormatInt(int64(j), 10)
			p := geojson.Position{X: rand.Float64()*360 - 180, Y: rand.Float64()*180 - 90, Z: 0}
			obj := geojson.Object(geojson.Point{Coordinates: p})
			col, ok := colsM[key]
			if !ok {
				col = New()
				colsM[key] = col
			}
			col.ReplaceOrInsert(id, obj, nil, nil)
			k++
		}
	}

	col := colsM["13"]
	//println(col.Count())
	bbox := geojson.BBox{Min: geojson.Position{X: -180, Y: 30, Z: 0}, Max: geojson.Position{X: 34, Y: 100, Z: 0}}
	col.geoSearch(0, bbox, func(id string, obj geojson.Object, fields []float64) bool {
		//println(id)
		return true
	})
}

type testPointItem struct {
	id     string
	object geojson.Object
}

func BenchmarkInsert(t *testing.B) {
	rand.Seed(time.Now().UnixNano())
	items := make([]testPointItem, t.N)
	for i := 0; i < t.N; i++ {
		items[i] = testPointItem{
			fmt.Sprintf("%d", i),
			geojson.SimplePoint{
				Y: rand.Float64()*180 - 90,
				X: rand.Float64()*360 - 180,
			},
		}
	}
	col := New()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		col.ReplaceOrInsert(items[i].id, items[i].object, nil, nil)
	}
}

func BenchmarkReplace(t *testing.B) {
	rand.Seed(time.Now().UnixNano())
	items := make([]testPointItem, t.N)
	for i := 0; i < t.N; i++ {
		items[i] = testPointItem{
			fmt.Sprintf("%d", i),
			geojson.SimplePoint{
				Y: rand.Float64()*180 - 90,
				X: rand.Float64()*360 - 180,
			},
		}
	}
	col := New()
	for i := 0; i < t.N; i++ {
		col.ReplaceOrInsert(items[i].id, items[i].object, nil, nil)
	}
	t.ResetTimer()
	for _, i := range rand.Perm(t.N) {
		o, _, _ := col.ReplaceOrInsert(items[i].id, items[i].object, nil, nil)
		if o != items[i].object {
			t.Fatal("shoot!")
		}
	}
}

func BenchmarkGet(t *testing.B) {
	rand.Seed(time.Now().UnixNano())
	items := make([]testPointItem, t.N)
	for i := 0; i < t.N; i++ {
		items[i] = testPointItem{
			fmt.Sprintf("%d", i),
			geojson.SimplePoint{
				Y: rand.Float64()*180 - 90,
				X: rand.Float64()*360 - 180,
			},
		}
	}
	col := New()
	for i := 0; i < t.N; i++ {
		col.ReplaceOrInsert(items[i].id, items[i].object, nil, nil)
	}
	t.ResetTimer()
	for _, i := range rand.Perm(t.N) {
		o, _, _ := col.Get(items[i].id)
		if o != items[i].object {
			t.Fatal("shoot!")
		}
	}
}

func BenchmarkRemove(t *testing.B) {
	rand.Seed(time.Now().UnixNano())
	items := make([]testPointItem, t.N)
	for i := 0; i < t.N; i++ {
		items[i] = testPointItem{
			fmt.Sprintf("%d", i),
			geojson.SimplePoint{
				Y: rand.Float64()*180 - 90,
				X: rand.Float64()*360 - 180,
			},
		}
	}
	col := New()
	for i := 0; i < t.N; i++ {
		col.ReplaceOrInsert(items[i].id, items[i].object, nil, nil)
	}
	t.ResetTimer()
	for _, i := range rand.Perm(t.N) {
		o, _, _ := col.Remove(items[i].id)
		if o != items[i].object {
			t.Fatal("shoot!")
		}
	}
}
