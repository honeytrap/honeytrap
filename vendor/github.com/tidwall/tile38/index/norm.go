package index

import "math"

// normPoint takes the latitude and longitude of one point and return the x,y position on a world map.
// The map bounds are minimum -180,-90 and maximum 180,90. These values are x,y; not lat,lon.
func normPoint(lat, lon float64) (x, y float64, normd bool) {
	// Check if the rect is completely in bounds.
	// This is likely to be the vast majority of cases.
	if lon >= -180 && lon <= 180 && lat >= -90 && lat <= 90 {
		return lon, lat, false
	}
	lat = math.Mod(lat, 360)
	for lat < -90 || lat > 90 {
		if lat < -90 {
			lat = -90 - (90 + lat)
			lon = 180 + lon
		}
		if lat > 90 {
			lat = 90 + (90 - lat)
			lon = 180 + lon
		}
	}
	lon = math.Mod(lon, 360)
	for lon < -180 {
		lon += 360
	}
	for lon > 180 {
		lon -= 360
	}
	return lon, lat, true
}

// normRect takes the latitude and longitude of two points which define a rectangle and returns an array of x,y rectangles on a world map.
// The map bounds are minimum -180,-90 and maximum 180,90. These values are x,y; not lat,lon.
func normRect(swLat, swLon, neLat, neLon float64) (mins, maxs [][]float64, normd bool) {
	mins, maxs, normd = normRectStep(swLat, swLon, neLat, neLon, nil, nil, false)
	return mins, maxs, normd
}

func normRectStep(swLat, swLon, neLat, neLon float64, mins, maxs [][]float64, normd bool) (minsOut, maxsOut [][]float64, normdOut bool) {
	// Make sure that the northeast point is greater than the southwest point.
	if neLat < swLat {
		swLat, neLat, normd = neLat, swLat, true
	}
	if neLon < swLon {
		swLon, neLon, normd = neLon, swLon, true
	}
	if swLon < -180 || neLon > 180 {
		// The rect is horizontally out of bounds.
		if neLon-swLon > 360 {
			// The rect goes around the world. Just normalize to -180 to 180.
			swLon = -180
			neLon = 180
		} else if swLon < -180 && neLon < -180 {
			// The rect is way left. Move it into range.
			// TODO: replace loops with math/mod.
			for {
				swLon += 360
				neLon += 360
				if swLon >= -180 || neLon >= -180 {
					break
				}
			}
		} else if swLon > 180 && neLon > 180 {
			// The rect is way right. Move it into range.
			// TODO: replace loops with math/mod.
			for {
				swLon -= 360
				neLon -= 360
				if swLon <= 180 || neLon <= 180 {
					break
				}
			}
		} else {
			// The rect needs to be split into two.
			if swLon < -180 {
				mins, maxs, normd = normRectStep(swLat, 180+(180+swLon), neLat, 180, mins, maxs, normd)
				mins, maxs, normd = normRectStep(swLat, -180, neLat, neLon, mins, maxs, normd)
			} else if neLon > 180 {
				mins, maxs, normd = normRectStep(swLat, swLon, neLat, 180, mins, maxs, normd)
				mins, maxs, normd = normRectStep(swLat, -180, neLat, -180+(neLon-180), mins, maxs, normd)
			} else {
				panic("should not be reached")
			}
			return mins, maxs, true
		}
		return normRectStep(swLat, swLon, neLat, neLon, mins, maxs, true)
	} else if swLat < -90 || neLat > 90 {
		// The rect is vertically out of bounds.
		if neLat-swLat > 360 {
			// The rect goes around the world.  Just normalize to -180 to 180.
			swLat = -180
			neLat = 180
		} else if swLat < -90 && neLat < -90 {
			swLat = -90 + (-90 - swLat)
			neLat = -90 + (-90 - neLat)
			swLon = swLon - 180
			neLon = neLon - 180
		} else if swLat > 90 && neLat > 90 {
			swLat = 90 - (swLat - 90)
			neLat = 90 - (neLat - 90)
			swLon = swLon - 180
			neLon = neLon - 180
		} else {
			if neLat > 90 {
				mins, maxs, normd = normRectStep(swLat, swLon, 90, neLon, mins, maxs, normd)
				mins, maxs, normd = normRectStep(90-(neLat-90), swLon-180, 90, neLon-180, mins, maxs, normd)
			} else if swLat < -90 {
				mins, maxs, normd = normRectStep(-90, swLon, neLat, neLon, mins, maxs, normd)
				mins, maxs, normd = normRectStep(-90, swLon-180, -90-(90+swLat), neLon-180, mins, maxs, normd)
			} else {
				panic("should not be reached")
			}
			return mins, maxs, true
		}
		return normRectStep(swLat, swLon, neLat, neLon, mins, maxs, true)
	} else {
		// rect is completely in bounds.
		mins = append(mins, []float64{swLon, swLat})
		maxs = append(maxs, []float64{neLon, neLat})
		return mins, maxs, normd
	}
}
