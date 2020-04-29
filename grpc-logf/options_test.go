package grpc_logf_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/richard-xtek/go-grpc-micro-kit/grpc-logf"
	"go.uber.org/zap/zapcore"
)

func TestDurationToTimeMillisField(t *testing.T) {
	val := grpc_logf.DurationToTimeMillisField(time.Microsecond * 100)
	assert.Equal(t, val.Type, zapcore.Float32Type, "should be a float type")
	assert.Equal(t, math.Float32frombits(uint32(val.Integer)), float32(0.1), "sub millisecond values should be correct")
}
