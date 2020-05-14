package math

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"

	converttype "xtek/exchange/user/base/utils/type"
)

// Round ...
func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// ToFixed ...
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}

// RandFloats ...
func RandFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}
	return res
}

// FloatToGeneralFormat ...
func FloatToGeneralFormat(v float64) *big.Int {
	// mul := Exp(big.NewInt(10), big.NewInt(8))
	// 	return FloatToBigInt(v, mul)
	vl := converttype.ParseValueFromDecimalToString(v, 8)
	ss := strings.Split(vl, ".")
	s1, _ := converttype.StringToInt64(ss[0])
	s2, _ := converttype.StringToInt64(ss[1])
	rs := big.NewInt(0)
	if s1 == 0 {
		rs = big.NewInt(s2)
	} else {
		formatDecimal := fmt.Sprintf("%.8d", s2)
		str := ss[0] + formatDecimal
		cvString, _ := converttype.StringToInt64(str)
		rs = big.NewInt(cvString)
	}
	return rs
}

// GeneralFormatToFloat ...
func GeneralFormatToFloat(n *big.Int) float64 {
	nominator := n
	denominator := Exp(big.NewInt(10), big.NewInt(8))
	rs := DivideToFloat(nominator, denominator)
	return rs
}

// GetMul ...
func GetMul() *big.Int {
	return Exp(big.NewInt(10), big.NewInt(8))
}

// MulQtyAndPricePoint ...
func MulQtyAndPricePoint(qty, pp *big.Int) *big.Int {
	denominator := Exp(big.NewInt(10), big.NewInt(8))
	return Div(Mul(qty, pp), denominator)
}
