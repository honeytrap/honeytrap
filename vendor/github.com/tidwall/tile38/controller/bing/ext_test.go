package bing

import (
	"math/rand"
	"testing"
	"time"
)

func TestIteratorFuzz(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10000; i++ {
		latitude := rand.Float64()*180.0 - 90.0
		longitude := rand.Float64()*380.0 - 180.0
		levelOfDetail := uint64((rand.Int() % MaxLevelOfDetail) + 1)
		pixelX, pixelY := LatLongToPixelXY(latitude, longitude, levelOfDetail)
		tileX, tileY := PixelXYToTileXY(pixelX, pixelY)
		quad1 := TileXYToQuadKey(tileX, tileY, levelOfDetail)
		l := rand.Int() % len(quad1)
		i := 0
		quad2 := ""
		LatLongToQuad(latitude, longitude, levelOfDetail, func(part int) bool {
			if i == l {
				return false
			}
			quad2 += string(byte(part) + '0')
			i++
			return true
		})
		if quad2 != quad1[:l] {
			t.Fatalf("[%d,%d] quad2 == %s, expect %s", i, levelOfDetail, quad2, quad1[:l])
		}
	}
}

func TestExt(t *testing.T) {
	// tileX, tileY, levelOfDetail := int64(0), int64(0), uint64(0)
	// parts := strings.Split(os.Getenv("TEST_TILE"), ",")
	// if len(parts) == 3 {
	// 	tileX, _ = strconv.ParseInt(parts[0], 10, 64)
	// 	tileY, _ = strconv.ParseInt(parts[1], 10, 64)
	// 	levelOfDetail, _ = strconv.ParseUint(parts[2], 10, 64)
	// }
	// minLat, minLon, maxLat, maxLon := TileXYToBounds(tileX, tileY, levelOfDetail)
	// fmt.Printf("\x1b[32m== Tile Boundaries ==\x1b[0m\n")
	// fmt.Printf("\x1b[31m%d,%d,%d\x1b[0m\n", tileX, tileY, levelOfDetail)
	// fmt.Printf("\x1b[31mWGS84 datum (longitude/latitude):\x1b[0m\n")
	// fmt.Printf("%v %v\n%v %v\n\n", minLon, minLat, maxLon, maxLat)

	//fmt.Printf("\x1b[32m\x1b[0m\n%v %v\n%v %v\n\n", minLon, minLat, maxLon, maxLat)
	// minLat, minLon, maxLat, maxLon = TileXYToBounds(1, 0, 1)
	// fmt.Printf("\x1b[32m1,0\x1b[0m\n%v %v\n%v %v\n\n", minLon, minLat, maxLon, maxLat)
	// minLat, minLon, maxLat, maxLon = TileXYToBounds(0, 1, 1)
	// fmt.Printf("\x1b[32m0,1\x1b[0m\n%v %v\n%v %v\n\n", minLon, minLat, maxLon, maxLat)
	// minLat, minLon, maxLat, maxLon = TileXYToBounds(1, 1, 1)
	// fmt.Printf("\x1b[32m1,1\x1b[0m\n%v %v\n%v %v\n\n", minLon, minLat, maxLon, maxLat)

	// minLat, minLon, maxLat, maxLon = TileXYToBounds(1, 0, 1)
	// fmt.Printf("1,0: %f,%f  %f,%f\n", minLat, minLon, maxLat, maxLon)
	// minLat, minLon, maxLat, maxLon = TileXYToBounds(0, 1, 1)
	// fmt.Printf("0,1: %f,%f  %f,%f\n", minLat, minLon, maxLat, maxLon)
	// minLat, minLon, maxLat, maxLon = TileXYToBounds(1, 1, 1)
	// fmt.Printf("1,1: %f,%f  %f,%f\n", minLat, minLon, maxLat, maxLon)
}
