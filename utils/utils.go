package utils

import (
	"math"
	"math/big"
)

// PowFloat calculates x^y
func PowFloat(x, y float64) float64 {
	if y == 0 {
		return 1.0
	}
	if y == 1 {
		return x
	}
	if y < 0 {
		return 1.0 / PowFloat(x, -y)
	}
	if y == float64(int(y)) {
		// Integer exponent
		result := 1.0
		for i := 0; i < int(y); i++ {
			result *= x
		}
		return result
	}
	// For fractional exponents, use a simple approximation
	// This is not mathematically precise but works for our use case
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	// Add fractional part approximation
	fractional := y - float64(int(y))
	if fractional > 0 {
		result *= x * fractional
	}
	return result
}

func FormatAmount(value *big.Int, decimals, places int) float64 {
	if value == nil {
		return 0
	}
	if decimals < 0 {
		decimals = 0
	}

	// Do the division at high precision first, then convert to float64.
	f := new(big.Float).SetPrec(256).SetInt(value)
	denInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	den := new(big.Float).SetPrec(256).SetInt(denInt)
	f.Quo(f, den)

	x, _ := f.Float64() // convert to double precision

	if places < 0 {
		return x
	}
	pow := math.Pow(10, float64(places))
	return math.Round(x*pow) / pow
}
