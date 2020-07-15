package subscriber

import (
	"context"
	"errors"
	"fmt"

	"github.com/richard-xtek/go-grpc-micro-kit/kafka"

	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"go.uber.org/zap"
)

// EventFuncHandler ...
type EventFuncHandler func(ctx context.Context, payload interface{}) error

var (
	handleRegistry *HandleRegistry

	marshalerProtobuf = new(kafka.ProtobufMarshaler)

	// ErrFnEventHandleNotFound ...
	ErrFnEventHandleNotFound = errors.New("Function handle not found")
	// ErrPbStructNotFound ...
	ErrPbStructNotFound = errors.New("Protobuf struct not found")
)

// HandleRegistry ...
type HandleRegistry struct {
	registries      map[kafka.EventType]EventFuncHandler
	protobufMapping map[kafka.EventType]interface{}
}

// GetHandleRegistry ...
func GetHandleRegistry() *HandleRegistry {
	if handleRegistry == nil {
		handleRegistry = &HandleRegistry{
			registries:      make(map[kafka.EventType]EventFuncHandler),
			protobufMapping: make(map[kafka.EventType]interface{}),
		}
	}
	return handleRegistry
}

// Register ...
func (r *HandleRegistry) Register(eventType kafka.EventType, fnHandler EventFuncHandler, protobuf interface{}) {
	r.registries[eventType] = fnHandler
	r.protobufMapping[eventType] = protobuf
}

// GetHandlerByEventType ...
func (r *HandleRegistry) GetHandlerByEventType(eventType kafka.EventType) (EventFuncHandler, error) {
	if fn, ok := r.registries[eventType]; ok {
		return fn, nil
	}
	return nil, ErrFnEventHandleNotFound
}

// GetPbStructByEventType ...
func (r *HandleRegistry) GetPbStructByEventType(eventType kafka.EventType) (interface{}, error) {
	if pbStruct, ok := r.protobufMapping[eventType]; ok {
		return pbStruct, nil
	}
	return nil, ErrPbStructNotFound
}

// ExecuteHandler ...
func ExecuteHandler(msg *kafka.Message, logger log.Factory) (err error) {
	defer func() {
		if errRecover := recover(); errRecover != nil {
			var ok bool
			err, ok = errRecover.(error)
			if !ok {
				fmt.Printf("Error execute handler %v, message_uuid %s\n", errRecover, msg.UUID)
			}
		}
	}()

	registry := GetHandleRegistry()
	data, err := registry.GetPbStructByEventType(msg.EventType)
	if err != nil {
		logger.Bg().Error("GetPbStructByEventType", zap.String("message_uuid", msg.UUID), zap.Error(err))
		return err
	}

	if err := marshalerProtobuf.Unmarshal(msg, data); err != nil {
		logger.Bg().Error("Unmarshal", zap.String("message_uuid", msg.UUID), zap.Error(err))
		return err
	}

	fnHandle, err := registry.GetHandlerByEventType(msg.EventType)
	if err != nil {
		return err
	}

	if err := fnHandle(msg.Context(), data); err != nil {
		return err
	}

	return err
}
