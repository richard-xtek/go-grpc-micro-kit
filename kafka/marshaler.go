package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

//  ...
const (
	// UUIDHeaderKey ...
	UUIDHeaderKey = "_message_uuid"
	// EventTypeHeaderKey ...
	EventTypeHeaderKey = "_event_type"
)

// Marshaler marshals Watermill's message to Kafka message.
type Marshaler interface {
	Marshal(topic string, msg *Message) (*sarama.ProducerMessage, error)
}

// Unmarshaler unmarshals Kafka's message to Watermill's message.
type Unmarshaler interface {
	Unmarshal(*sarama.ConsumerMessage) (*Message, error)
}

// MarshalerUnmarshaler ...
type MarshalerUnmarshaler interface {
	Marshaler
	Unmarshaler
}

// DefaultMarshaler ...
type DefaultMarshaler struct{}

// Marshal ...
func (DefaultMarshaler) Marshal(topic string, msg *Message) (*sarama.ProducerMessage, error) {
	if value := msg.Metadata.Get(UUIDHeaderKey); value != "" {
		return nil, errors.Errorf("metadata %s is reserved for message UUID", UUIDHeaderKey)
	}

	headers := []sarama.RecordHeader{{
		Key:   []byte(UUIDHeaderKey),
		Value: []byte(msg.UUID),
	}}

	for key, value := range msg.Metadata {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	return &sarama.ProducerMessage{
		Topic:   topic,
		Value:   sarama.ByteEncoder(msg.Payload),
		Headers: headers,
	}, nil
}

// Unmarshal ...
func (DefaultMarshaler) Unmarshal(kafkaMsg *sarama.ConsumerMessage) (*Message, error) {
	var (
		messageID string
		eventType EventType
	)
	metadata := make(Metadata, len(kafkaMsg.Headers))

	for _, header := range kafkaMsg.Headers {
		switch string(header.Key) {
		case UUIDHeaderKey:
			messageID = string(header.Value)
			break
		case EventTypeHeaderKey:
			eventType = EventType(header.Value)
			break
		default:
			metadata.Set(string(header.Key), string(header.Value))
		}
	}

	msg := NewMessage(messageID, kafkaMsg.Value)
	msg.Metadata = metadata
	msg.EventType = eventType

	return msg, nil
}

// GeneratePartitionKey ...
type GeneratePartitionKey func(topic string, msg *Message) (string, error)

// kafkaJsonWithPartitioning ...
type kafkaJSONWithPartitioning struct {
	DefaultMarshaler

	generatePartitionKey GeneratePartitionKey
}

// NewWithPartitioningMarshaler ...
func NewWithPartitioningMarshaler(generatePartitionKey GeneratePartitionKey) MarshalerUnmarshaler {
	return kafkaJSONWithPartitioning{generatePartitionKey: generatePartitionKey}
}

func (j kafkaJSONWithPartitioning) Marshal(topic string, msg *Message) (*sarama.ProducerMessage, error) {
	kafkaMsg, err := j.DefaultMarshaler.Marshal(topic, msg)
	if err != nil {
		return nil, err
	}

	key, err := j.generatePartitionKey(topic, msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate partition key")
	}
	kafkaMsg.Key = sarama.ByteEncoder(key)

	return kafkaMsg, nil
}
