// https://msdn.microsoft.com/en-us/library/bb259689.aspx

package bing

import "math"

const (
	// EarthRadius is the radius of the earth
	EarthRadius = 6378137.0
	// MinLatitude is the min lat
	MinLatitude = -85.05112878
	// MaxLatitude is the max lat
	MaxLatitude = 85.05112878
	// MinLongitude is the min lon
	MinLongitude = -180.0
	// MaxLongitude is the max lon
	MaxLongitude = 180.0
	// TileSize is the size of a tile
	TileSize = 256
	// MaxLevelOfDetail is the max level of detail
	MaxLevelOfDetail = 38
)

// Clips a number to the specified minimum and maximum values.
// Param 'n' is the number to clip.
// Param 'minValue' is the minimum allowable value.
// Param 'maxValue' is the maximum allowable value.
// Returns the clipped value.
func clip(n, minValue, maxValue float64) float64 {
	if n < minValue {
		return minValue
	}
	if n > maxValue {
		return maxValue
	}
	return n
}

// MapSize determines the map width and height (in pixels) at a specified level of detail.
// Param 'levelOfDetail' is the level of detail, from 1 (lowest detail) to N (highest detail).
// Returns the map width and height in pixels.
func MapSize(levelOfDetail uint64) uint64 {
	return TileSize << levelOfDetail
}

// // Determines the ground resolution (in meters per pixel) at a specified latitude and level of detail.
// // Param 'latitude' is the Latitude (in degrees) at which to measure the ground resolution.
// // Param 'levelOfDetail' is the Level of detail, from 1 (lowest detail) to N (highest detail).
// // Returns the ground resolution, in meters per pixel.
// func GroundResolution(latitude float64, levelOfDetail uint64) float64 {
// 	latitude = clip(latitude, MinLatitude, MaxLatitude)
// 	return math.Cos(latitude*math.Pi/180) * 2 * math.Pi * EarthRadius / float64(MapSize(levelOfDetail))
// }

// // Determines the map scale at a specified latitude, level of detail, and screen resolution.
// // Param 'latitude' is the latitude (in degrees) at which to measure the map scale.
// // Param 'levelOfDetail' is the level of detail, from 1 (lowest detail) to N (highest detail).
// // Param 'screenDpi' is the resolution of the screen, in dots per inch.
// // Returns the map scale, expressed as the denominator N of the ratio 1 : N.
// func MapScale(latitude float64, levelOfDetail, screenDpi uint64) float64 {
// 	return GroundResolution(latitude, levelOfDetail) * float64(screenDpi) / 0.0254
// }

// LatLongToPixelXY converts a point from latitude/longitude WGS-84 coordinates (in degrees) into pixel XY coordinates at a specified level of detail.
// Param 'latitude' is the latitude of the point, in degrees.
// Param 'longitude' is the longitude of the point, in degrees.
// Param 'levelOfDetail' is the level of detail, from 1 (lowest detail) to N (highest detail).
// Return value 'pixelX' is the output parameter receiving the X coordinate in pixels.
// Return value 'pixelY' is the output parameter receiving the Y coordinate in pixels.
func LatLongToPixelXY(latitude, longitude float64, levelOfDetail uint64) (pixelX, pixelY int64) {
	latitude = clip(latitude, MinLatitude, MaxLatitude)
	longitude = clip(longitude, MinLongitude, MaxLongitude)
	x := (longitude + 180) / 360
	sinLatitude := math.Sin(latitude * math.Pi / 180)
	y := 0.5 - math.Log((1+sinLatitude)/(1-sinLatitude))/(4*math.Pi)
	mapSize := float64(MapSize(levelOfDetail))
	pixelX = int64(clip(x*mapSize+0.5, 0, mapSize-1))
	pixelY = int64(clip(y*mapSize+0.5, 0, mapSize-1))
	return
}

// PixelXYToLatLong converts a pixel from pixel XY coordinates at a specified level of detail into latitude/longitude WGS-84 coordinates (in degrees).
// Param 'pixelX' is the X coordinate of the point, in pixels.
// Param 'pixelY' is the Y coordinates of the point, in pixels.
// Param 'levelOfDetail' is the level of detail, from 1 (lowest detail) to N (highest detail).
// Return value 'latitude' is the output parameter receiving the latitude in degrees.
// Return value 'longitude' is the output parameter receiving the longitude in degrees.
func PixelXYToLatLong(pixelX, pixelY int64, levelOfDetail uint64) (latitude, longitude float64) {
	mapSize := float64(MapSize(levelOfDetail))
	x := (clip(float64(pixelX), 0, mapSize-1) / mapSize) - 0.5
	y := 0.5 - (clip(float64(pixelY), 0, mapSize-1) / mapSize)
	latitude = 90 - 360*math.Atan(math.Exp(-y*2*math.Pi))/math.Pi
	longitude = 360 * x
	return
}

// PixelXYToTileXY converts pixel XY coordinates into tile XY coordinates of the tile containing the specified pixel.
// Param 'pixelX' is the pixel X coordinate.
// Param 'pixelY' is the pixel Y coordinate.
// Return value 'tileX' is the output parameter receiving the tile X coordinate.
// Return value 'tileY' is the output parameter receiving the tile Y coordinate.
func PixelXYToTileXY(pixelX, pixelY int64) (tileX, tileY int64) {
	return pixelX >> 8, pixelY >> 8
}

// TileXYToPixelXY converts tile XY coordinates into pixel XY coordinates of the upper-left pixel of the specified tile.
// Param 'tileX' is the tile X coordinate.
// Param 'tileY' is the tile Y coordinate.
// Return value 'pixelX' is the output parameter receiving the pixel X coordinate.
// Return value 'pixelY' is the output parameter receiving the pixel Y coordinate.
func TileXYToPixelXY(tileX, tileY int64) (pixelX, pixelY int64) {
	return tileX << 8, tileY << 8
}

/// TileXYToQuadKey converts tile XY coordinates into a QuadKey at a specified level of detail.
/// Param 'tileX' is the tile X coordinate.
/// Param 'tileY' is the tile Y coordinate.
/// Param 'levelOfDetail' is the Level of detail, from 1 (lowest detail) to N (highest detail).
/// Returns a string containing the QuadKey.
func TileXYToQuadKey(tileX, tileY int64, levelOfDetail uint64) string {
	quadKey := make([]byte, levelOfDetail)
	for i, j := levelOfDetail, 0; i > 0; i, j = i-1, j+1 {
		mask := int64(1 << (i - 1))
		if (tileX & mask) != 0 {
			if (tileY & mask) != 0 {
				quadKey[j] = '3'
			} else {
				quadKey[j] = '1'
			}
		} else if (tileY & mask) != 0 {
			quadKey[j] = '2'
		} else {
			quadKey[j] = '0'
		}
	}
	return string(quadKey)
}

/// QuadKeyToTileXY converts a QuadKey into tile XY coordinates.
/// Param 'quadKey' is the quadKey of the tile.
/// Return value 'tileX' is the output parameter receiving the tile X coordinate.
/// Return value 'tileY is the output parameter receiving the tile Y coordinate.
/// Return value 'levelOfDetail' is the output parameter receiving the level of detail.
func QuadKeyToTileXY(quadKey string) (tileX, tileY int64, levelOfDetail uint64) {
	levelOfDetail = uint64(len(quadKey))
	for i := levelOfDetail; i > 0; i-- {
		mask := int64(1 << (i - 1))
		switch quadKey[levelOfDetail-i] {
		case '0':
		case '1':
			tileX |= mask
		case '2':
			tileY |= mask
		case '3':
			tileX |= mask
			tileY |= mask
		default:
			panic("Invalid QuadKey digit sequence.")
		}
	}
	return
}
