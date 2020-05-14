package math

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBStringToBigInt(t *testing.T) {
	t.Run("Testing", func(t *testing.T) {
		s := "000000000100000000"
		num := DBStringToBigInt(s)
		assert.Equal(t, big.NewInt(1e8), num, "not equal so it error")
	})
}

func TestBigIntToDBString(t *testing.T) {
	t.Run("Testing", func(t *testing.T) {
		n := big.NewInt(1e8)
		s := BigIntToDBString(n)
		assert.Equal(t, 18, len(s), "length should be 18")
	})
}
