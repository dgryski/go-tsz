package main

import (
	"math"
)

func Round(f float64) float64 {
	if f < 0 {
		return math.Ceil(f - 0.5)
	}
	return math.Floor(f + .5)
}

func RoundNum(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
