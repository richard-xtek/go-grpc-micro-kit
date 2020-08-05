package math

import (
	"fmt"
	"math/big"
	"reflect"
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

func TestFromDecimalString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Int
		wantErr bool
	}{
		struct {
			name    string
			args    args
			want    *big.Int
			wantErr bool
		}{
			name:    " 1",
			args:    args{str: "3.0007"},
			want:    big.NewInt(300070000),
			wantErr: false,
		},
		{
			name:    " 1",
			args:    args{str: "0.9"},
			want:    big.NewInt(90000000),
			wantErr: false,
		},
		{
			name:    " 1",
			args:    args{str: "3.0"},
			want:    big.NewInt(300000000),
			wantErr: false,
		},
		{
			name:    " 1",
			args:    args{str: "1.245678899333"},
			want:    big.NewInt(124567889),
			wantErr: false,
		},
		{
			name:    " 1",
			args:    args{str: "92233720368.0"},
			want:    big.NewInt(9223372036800000000),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromDecimalString(tt.args.str)
			fmt.Printf("got: %s\n", got.String())
			if (err != nil) != tt.wantErr {
				t.Errorf("FromDecimalString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromDecimalString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDecimalString(t *testing.T) {
	type args struct {
		num *big.Int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		struct {
			name string
			args args
			want string
		}{
			name: "1",
			args: args{
				num: big.NewInt(90000000),
			},
			want: "0.9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToDecimalString(tt.args.num); got != tt.want {
				t.Errorf("ToDecimalString() = %v, want %v", got, tt.want)
			}
		})
	}
}
