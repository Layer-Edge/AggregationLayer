package utils

import (
	"math"
	"math/big"
)

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
