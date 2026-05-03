package utils

import "math"

// RoundBaht rounds a money value (baht) to 2 decimal places (satang precision).
func RoundBaht(v float64) float64 {
	return math.Round(v*100) / 100
}

// RoundGram rounds a gold weight (grams) to 6 decimal places.
// 1e6 chosen so that sub-milligram increments are preserved while killing
// the long floating-point tail produced by amount/goldPrice divisions.
func RoundGram(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}
