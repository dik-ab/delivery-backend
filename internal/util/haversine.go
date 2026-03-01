package util

import "math"

const earthRadiusKm = 6371.0

// CalculateDistance calculates the great circle distance between two points
// on the earth (specified in decimal degrees) using the Haversine formula.
// Returns distance in kilometers.
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180.0
	lon1Rad := lon1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	lon2Rad := lon2 * math.Pi / 180.0

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Asin(math.Sqrt(a))

	return earthRadiusKm * c
}

// IsWithinRadius checks if a point is within a certain radius of another point
func IsWithinRadius(lat1, lon1, lat2, lon2 float64, radiusKm float64) bool {
	distance := CalculateDistance(lat1, lon1, lat2, lon2)
	return distance <= radiusKm
}
