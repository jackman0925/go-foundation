package geox

import "math"

const earthRadiusMeters = 6371000.0

// DistanceMeters returns the approximate distance between two WGS84 points.
func DistanceMeters(lon1 float64, lat1 float64, lon2 float64, lat2 float64) float64 {
	rad := math.Pi / 180.0
	x := (lon2 - lon1) * rad * math.Cos((lat1+lat2)/2*rad)
	y := (lat2 - lat1) * rad
	return math.Sqrt(x*x+y*y) * earthRadiusMeters
}
