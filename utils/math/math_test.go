package math

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloatToGeneralFormat(t *testing.T) {
	c1 := 10.03
	valueInDB := FloatToGeneralFormat(c1)
	exp1 := big.NewInt(1003000000)

	c2 := 0.9
	valueInDB2 := FloatToGeneralFormat(c2)
	exp2 := big.NewInt(90000000)

	c3 := 3.0
	valueInDB3 := FloatToGeneralFormat(c3)
	exp3 := big.NewInt(300000000)

	c4 := 92233720368.0
	valueInDB4 := FloatToGeneralFormat(c4)
	exp4 := big.NewInt(9223372036800000000)

	c5 := 3.0007
	valueInDB5 := FloatToGeneralFormat(c5)
	exp5 := big.NewInt(300070000)

	// fmt.Println(exp1, valueInDB)
	assert.Equal(t, exp1, valueInDB, "Should be 1003000000")
	assert.Equal(t, exp2, valueInDB2, "Should be 90000000")
	assert.Equal(t, exp3, valueInDB3, "Should be 300000000")
	assert.Equal(t, exp4, valueInDB4, "Should be 9223372036800000000")
	assert.Equal(t, exp5, valueInDB5, "Should be 300070000")

}

func TestGeneralFormatToFloat(t *testing.T) {
	valueInDB := big.NewInt(31293100000)

	exp := 312.931

	rs := GeneralFormatToFloat(valueInDB)

	assert.Equal(t, exp, rs, "Should be 312.931")
}
