package math

import (
	"fmt"
	"math/big"
)

func Mul(x, y *big.Int) *big.Int {
	return big.NewInt(0).Mul(x, y)
}

func Div(x, y *big.Int) *big.Int {
	return big.NewInt(0).Div(x, y)
}

// Add ...
func Add(x, y *big.Int) *big.Int {
	return big.NewInt(0).Add(x, y)
}

// Sub ...
func Sub(x, y *big.Int) *big.Int {
	return big.NewInt(0).Sub(x, y)
}

// Neg ...
func Neg(x *big.Int) *big.Int {
	return big.NewInt(0).Neg(x)
}

// Avg ...
func Avg(x *big.Int, y *big.Int) *big.Int {
	return Div(Add(x, y), big.NewInt(2))
}

// ToBigInt ...
func ToBigInt(s string) *big.Int {
	res := big.NewInt(0)
	res.SetString(s, 10)
	return res
}

// FloatToBigInt ...
func FloatToBigInt(n float64, mul *big.Int) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(n)

	coin := new(big.Float)
	coin.SetInt(mul)

	bigval.Mul(bigval, coin)

	result := new(big.Int)
	bigval.Int(result) // store converted number in result

	return result
}

// BigIntToDBString ...
func BigIntToDBString(n *big.Int) string {
	layer := fmt.Sprintf("%.18d", n)
	return layer
}

// DBStringToBigInt ...
func DBStringToBigInt(s string) *big.Int {
	// set value
	dbNum := big.NewInt(0)
	dbNum.SetString(s, 10)
	return dbNum
}

// Exp ...
func Exp(x, y *big.Int) *big.Int {
	return big.NewInt(0).Exp(x, y, nil)
}

// BigIntToBigFloat ...
func BigIntToBigFloat(a *big.Int) *big.Float {
	b := new(big.Float).SetInt(a)
	return b
}

// ToBigFraction ...
func ToBigFraction(a, b *big.Int) *big.Rat {
	return big.NewRat(1, 1).SetFrac(a, b)
}

// DivideToFloat ...
func DivideToFloat(a, b *big.Int) float64 {
	res, _ := big.NewRat(1, 1).SetFrac(a, b).Float64()

	return res
}

// ToDecimal ...
func ToDecimal(value *big.Int) float64 {
	bigFloatValue := BigIntToBigFloat(value)
	result := DivFloat(bigFloatValue, big.NewFloat(1e18))

	floatValue, _ := result.Float64()
	return floatValue
}

// DivFloat ...
func DivFloat(x, y *big.Float) *big.Float {
	return big.NewFloat(0).Quo(x, y)
}

// Max ...
func Max(a, b *big.Int) *big.Int {
	if a.Cmp(b) == 1 {
		return a
	} else {
		return b
	}
}

// IsZero ...
func IsZero(x *big.Int) bool {
	if x.Cmp(big.NewInt(0)) == 0 {
		return true
	} else {
		return false
	}
}

// IsEqual ...
func IsEqual(x, y *big.Int) bool {
	if x.Cmp(y) == 0 {
		return true
	} else {
		return false
	}
}

// IsNotEqual ...
func IsNotEqual(x, y *big.Int) bool {
	if x.Cmp(y) != 0 {
		return true
	} else {
		return false
	}
}

// IsGreaterThan ...
func IsGreaterThan(x, y *big.Int) bool {
	if x.Cmp(y) == 1 || x.Cmp(y) == 0 {
		return true
	} else {
		return false
	}
}

// IsStrictlyGreaterThan ...
func IsStrictlyGreaterThan(x, y *big.Int) bool {
	if x.Cmp(y) == 1 {
		return true
	} else {
		return false
	}
}

func IsSmallerThan(x, y *big.Int) bool {
	if x.Cmp(y) == -1 || x.Cmp(y) == 0 {
		return true
	} else {
		return false
	}
}

func IsStrictlySmallerThan(x, y *big.Int) bool {
	if x.Cmp(y) == -1 {
		return true
	} else {
		return false
	}
}

func IsEqualOrGreaterThan(x, y *big.Int) bool {
	return (IsEqual(x, y) || IsGreaterThan(x, y))
}

func IsEqualOrSmallerThan(x, y *big.Int) bool {
	return (IsEqual(x, y) || IsSmallerThan(x, y))
}
