package kafka

import (
	"context"
	"time"
)

type contextKey int

const (
	_ contextKey = iota
	partitionContextKey
	partitionOffsetContextKey
	timestampContextKey
	messageUUIDKey
)

func setPartitionToCtx(ctx context.Context, partition int32) context.Context {
	return context.WithValue(ctx, partitionContextKey, partition)
}

// MessagePartitionFromCtx returns Kafka partition of the consumed message
func MessagePartitionFromCtx(ctx context.Context) (int32, bool) {
	partition, ok := ctx.Value(partitionContextKey).(int32)
	return partition, ok
}

func setPartitionOffsetToCtx(ctx context.Context, offset int64) context.Context {
	return context.WithValue(ctx, partitionOffsetContextKey, offset)
}

// MessagePartitionOffsetFromCtx returns Kafka partition offset of the consumed message
func MessagePartitionOffsetFromCtx(ctx context.Context) (int64, bool) {
	offset, ok := ctx.Value(partitionOffsetContextKey).(int64)
	return offset, ok
}

func setMessageTimestampToCtx(ctx context.Context, timestamp time.Time) context.Context {
	return context.WithValue(ctx, timestampContextKey, timestamp)
}

// MessageTimestampFromCtx returns Kafka internal timestamp of the consumed message
func MessageTimestampFromCtx(ctx context.Context) (time.Time, bool) {
	timestamp, ok := ctx.Value(timestampContextKey).(time.Time)
	return timestamp, ok
}

func setMessageUUIDKeyToCtx(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, messageUUIDKey, uuid)

}

// MessageUUIDFromCtx returns Kafka internal timestamp of the consumed message
func MessageUUIDFromCtx(ctx context.Context) (string, bool) {
	str, ok := ctx.Value(messageUUIDKey).(string)
	return str, ok
}
