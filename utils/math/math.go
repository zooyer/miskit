package math

import "math"

func Ceil(f float64, n int) float64 {
	pow := math.Pow10(n)
	if f < 0 {
		return math.Trunc((f/pow)*pow) / pow
	}
	return math.Trunc((f+0.5/pow)*pow) / pow
}

func Floor(f float64, n int) float64 {
	pow := math.Pow10(n)
	if f > 0 {
		return math.Trunc((f/pow)*pow) / pow
	}
	return math.Trunc((f-0.5/pow)*pow) / pow
}

func Round(f float64, n int) float64 {
	pow := math.Pow10(n)
	if f > 0 {
		return math.Trunc((f+0.5/pow)*pow) / pow
	}
	return math.Trunc((f-0.5/pow)*pow) / pow
}
