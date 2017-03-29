package bing

import "errors"

// LatLongToQuad iterates through all of the quads parts until levelOfDetail is reached.
func LatLongToQuad(latitude, longitude float64, levelOfDetail uint64, iterator func(part int) bool) {
	pixelX, pixelY := LatLongToPixelXY(latitude, longitude, levelOfDetail)
	tileX, tileY := PixelXYToTileXY(pixelX, pixelY)
	for i := levelOfDetail; i > 0; i-- {
		if !iterator(partForTileXY(tileX, tileY, i)) {
			break
		}
	}
}

func partForTileXY(tileX, tileY int64, levelOfDetail uint64) int {
	mask := int64(1 << (levelOfDetail - 1))
	if (tileX & mask) != 0 {
		if (tileY & mask) != 0 {
			return 3
		}
		return 1
	} else if (tileY & mask) != 0 {
		return 2
	}
	return 0
}

// TileXYToBounds returns the bounds around a tile.
func TileXYToBounds(tileX, tileY int64, levelOfDetail uint64) (minLat, minLon, maxLat, maxLon float64) {
	size := int64(1 << levelOfDetail)
	pixelX, pixelY := TileXYToPixelXY(tileX, tileY)
	maxLat, minLon = PixelXYToLatLong(pixelX, pixelY, levelOfDetail)
	pixelX, pixelY = TileXYToPixelXY(tileX+1, tileY+1)
	minLat, maxLon = PixelXYToLatLong(pixelX, pixelY, levelOfDetail)
	if tileX%size == 0 {
		minLon = MinLongitude
	}
	if tileX%size == size-1 {
		maxLon = MaxLongitude
	}
	if tileY <= 0 {
		maxLat = MaxLatitude
	}
	if tileY >= size-1 {
		minLat = MinLatitude
	}
	return
}

// QuadKeyToBounds converts a quadkey to bounds
func QuadKeyToBounds(quadkey string) (minLat, minLon, maxLat, maxLon float64, err error) {
	for i := 0; i < len(quadkey); i++ {
		switch quadkey[i] {
		case '0', '1', '2', '3':
		default:
			err = errors.New("invalid quadkey")
			return
		}
	}
	minLat, minLon, maxLat, maxLon = TileXYToBounds(QuadKeyToTileXY(quadkey))
	return
}
