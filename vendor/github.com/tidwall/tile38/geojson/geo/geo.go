package geo

import "math"

const earthRadius = 6371e3

func toRadians(deg float64) float64 { return deg * math.Pi / 180 }
func toDegrees(rad float64) float64 { return rad * 180 / math.Pi }

// DistanceTo return the distance in meteres between two point.
func DistanceTo(latA, lonA, latB, lonB float64) (meters float64) {
	φ1 := toRadians(latA)
	λ1 := toRadians(lonA)
	φ2 := toRadians(latB)
	λ2 := toRadians(lonB)
	Δφ := φ2 - φ1
	Δλ := λ2 - λ1
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

// DestinationPoint return the destination from a point based on a distance and bearing.
func DestinationPoint(lat, lon, meters, bearingDegrees float64) (destLat, destLon float64) {
	// see http://williams.best.vwh.net/avform.htm#LL
	δ := meters / earthRadius // angular distance in radians
	θ := toRadians(bearingDegrees)
	φ1 := toRadians(lat)
	λ1 := toRadians(lon)
	φ2 := math.Asin(math.Sin(φ1)*math.Cos(δ) + math.Cos(φ1)*math.Sin(δ)*math.Cos(θ))
	λ2 := λ1 + math.Atan2(math.Sin(θ)*math.Sin(δ)*math.Cos(φ1), math.Cos(δ)-math.Sin(φ1)*math.Sin(φ2))
	λ2 = math.Mod(λ2+3*math.Pi, 2*math.Pi) - math.Pi // normalise to -180..+180°
	return toDegrees(φ2), toDegrees(λ2)
}
