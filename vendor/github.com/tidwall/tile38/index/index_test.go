package index

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func randf(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func randPoint() (lat float64, lon float64) {
	// intentionally go out of range.
	return randf(-100, 100), randf(-190, 190)
}

func randRect() (swLat, swLon, neLat, neLon float64) {
	swLat, swLon = randPoint()
	// intentionally go out of range even more.
	neLat = randf(swLat-10, swLat+10)
	neLon = randf(swLon-10, swLon+10)
	return
}

func wp(swLat, swLon, neLat, neLon float64) *FlexItem {
	return &FlexItem{
		MinX: swLon,
		MinY: swLat,
		MaxX: neLon,
		MaxY: neLat,
	}
}

func TestRandomInserts(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	l := 200000
	tr := New()
	start := time.Now()
	i := 0
	for ; i < l/2; i++ {
		swLat, swLon := randPoint()
		tr.Insert(wp(swLat, swLon, swLat, swLon))
	}
	inspdur := time.Now().Sub(start)

	start = time.Now()
	for ; i < l; i++ {
		swLat, swLon, neLat, neLon := randRect()
		tr.Insert(wp(swLat, swLon, neLat, neLon))
	}
	insrdur := time.Now().Sub(start)
	count := 0

	count = tr.Count()
	if count != l {
		t.Fatalf("count == %d, expect %d", count, l)
	}
	count = 0
	items := make([]Item, 0, l)
	tr.Search(0, -90, -180, 90, 180, 0, 0, func(item Item) bool {
		count++
		items = append(items, item)
		return true
	})
	if count != l {
		t.Fatalf("count == %d, expect %d", count, l)
	}
	start = time.Now()
	count1 := 0
	tr.Search(0, 33, -115, 34, -114, 0, 0, func(item Item) bool {
		count1++
		return true
	})
	searchdur1 := time.Now().Sub(start)

	start = time.Now()
	count2 := 0

	tr.Search(0, 33-180, -115-360, 34-180, -114-360, 0, 0, func(item Item) bool {
		count2++
		return true
	})
	searchdur2 := time.Now().Sub(start)

	start = time.Now()
	count3 := 0
	tr.Search(0, -10, 170, 20, 200, 0, 0, func(item Item) bool {
		count3++
		return true
	})
	searchdur3 := time.Now().Sub(start)

	fmt.Printf("Randomly inserted %d points in %s.\n", l/2, inspdur.String())
	fmt.Printf("Randomly inserted %d rects in %s.\n", l/2, insrdur.String())
	fmt.Printf("Searched %d items in %s.\n", count1, searchdur1.String())
	fmt.Printf("Searched %d items in %s.\n", count2, searchdur2.String())
	fmt.Printf("Searched %d items in %s.\n", count3, searchdur3.String())

	tr.Search(0, -10, 170, 20, 200, 0, 0, func(item Item) bool {
		lat1, lon1, _, lat2, lon2, _ := item.Rect()
		if lat1 == lat2 && lon1 == lon2 {
			return false
		}
		return true
	})

	tr.Search(0, -10, 170, 20, 200, 0, 0, func(item Item) bool {
		lat1, lon1, _, lat2, lon2, _ := item.Rect()
		if lat1 != lat2 || lon1 != lon2 {
			return false
		}
		return true
	})

	// Remove all of the elements
	for _, item := range items {
		tr.Remove(item)
	}

	count = tr.Count()
	if count != 0 {
		t.Fatalf("count == %d, expect %d", count, 0)
	}

	tr.RemoveAll()
	/*	if tr.getQTreeItem(nil) != nil {
			t.Fatal("getQTreeItem(nil) should return nil")
		}
	*/
	if tr.getRTreeItem(nil) != nil {
		t.Fatal("getRTreeItem(nil) should return nil")
	}
}

func TestMemory(t *testing.T) {
	rand.Seed(0)
	l := 100000
	tr := New()
	for i := 0; i < l; i++ {
		swLat, swLon, neLat, neLon := randRect()
		if rand.Int()%2 == 0 { // one in three chance that the rect is actually a point.
			neLat, neLon = swLat, swLon
		}
		tr.Insert(wp(swLat, swLon, neLat, neLon))
	}
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	const PtrSize = 32 << uintptr(uint64(^uintptr(0))>>63)
	fmt.Printf("Memory consumption is %d bytes/object. Pointers are %d bytes.\n", int(m.HeapAlloc)/tr.Count(), PtrSize/8)
}

func TestInsertVarious(t *testing.T) {
	var count int
	tr := New()
	item := wp(33, -115, 33, -115)
	tr.Insert(item)
	count = tr.Count()
	if count != 1 {
		t.Fatalf("count = %d, expect 1", count)
	}
	tr.Remove(item)
	count = tr.Count()
	if count != 0 {
		t.Fatalf("count = %d, expect 0", count)
	}
	tr.Insert(item)
	count = tr.Count()
	if count != 1 {
		t.Fatalf("count = %d, expect 1", count)
	}
	found := false
	tr.Search(0, -90, -180, 90, 180, 0, 0, func(item2 Item) bool {
		if item2 == item {
			found = true
		}
		return true
	})
	if !found {
		t.Fatal("did not find item")
	}
}

func BenchmarkInsertRect(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	tr := New()
	for i := 0; i < b.N; i++ {
		swLat, swLon, neLat, neLon := randRect()
		tr.Insert(wp(swLat, swLon, neLat, neLon))
	}
}

func BenchmarkInsertPoint(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	tr := New()
	for i := 0; i < b.N; i++ {
		swLat, swLon, _, _ := randRect()
		tr.Insert(wp(swLat, swLon, swLat, swLon))
	}
}

func BenchmarkInsertEither(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	tr := New()
	for i := 0; i < b.N; i++ {
		swLat, swLon, neLat, neLon := randRect()
		if rand.Int()%3 == 0 { // one in three chance that the rect is actually a point.
			neLat, neLon = swLat, swLon
		}
		tr.Insert(wp(swLat, swLon, neLat, neLon))
	}
}

// func BenchmarkSearchRect(b *testing.B) {
// 	rand.Seed(time.Now().UnixNano())
// 	tr := New()
// 	for i := 0; i < 100000; i++ {
// 		swLat, swLon, neLat, neLon := randRect()
// 		tr.Insert(swLat, swLon, neLat, neLon)
// 	}
// 	b.ResetTimer()
// 	count := 0
// 	//for i := 0; i < b.N; i++ {
// 	tr.Search(0, -180, 90, 180, func(id int) bool {
// 		count++
// 		return true
// 	})
// 	//}
// 	println(count)
// }
