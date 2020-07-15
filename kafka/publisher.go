package kafka

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"go.uber.org/zap"
)

// Publisher ...
type Publisher struct {
	config   PublisherConfig
	producer sarama.SyncProducer
	logger   log.Factory

	closed bool
}

// NewPublisher creates a new Kafka Publisher.
func NewPublisher(
	config PublisherConfig,
	logger log.Factory,
) (*Publisher, error) {
	config.setDefaults()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(config.Brokers, config.OverwriteSaramaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Kafka producer")
	}

	return &Publisher{
		config:   config,
		producer: producer,
		logger:   logger,
	}, nil
}

// PublisherConfig ...
type PublisherConfig struct {
	// Kafka brokers list.
	Brokers []string

	// Marshaler is used to marshal messages from Watermill format into Kafka format.
	Marshaler Marshaler

	// OverwriteSaramaConfig holds additional sarama settings.
	OverwriteSaramaConfig *sarama.Config
}

func (c *PublisherConfig) setDefaults() {
	if c.OverwriteSaramaConfig == nil {
		c.OverwriteSaramaConfig = DefaultSaramaSyncPublisherConfig()
	}
}

// Validate ...
func (c PublisherConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("missing brokers")
	}
	if c.Marshaler == nil {
		return errors.New("missing marshaler")
	}

	return nil
}

// DefaultSaramaSyncPublisherConfig ...
func DefaultSaramaSyncPublisherConfig() *sarama.Config {
	config := sarama.NewConfig()

	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Version = sarama.V1_0_0_0
	config.Metadata.Retry.Backoff = time.Second * 2
	config.ClientID = "watermill"

	return config
}

// Publish publishes message to Kafka.
//
// Publish is blocking and wait for ack from Kafka.
// When one of messages delivery fails - function is interrupted.
func (p *Publisher) Publish(topic string, msgs ...*Message) error {
	if p.closed {
		return errors.New("publisher closed")
	}

	logFields := []zap.Field{zap.String("topic", topic)}

	for _, msg := range msgs {
		logMsgFields := append(logFields, zap.String("message_uuid", msg.UUID))
		p.logger.Bg().Debug("Sending message to Kafka", logMsgFields...)

		kafkaMsg, err := p.config.Marshaler.Marshal(topic, msg)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal message %s", msg.UUID)
		}

		partition, offset, err := p.producer.SendMessage(kafkaMsg)
		if err != nil {
			return errors.Wrapf(err, "cannot produce message %s", msg.UUID)
		}

		logMsgFields = append(logMsgFields, zap.Int32("kafka_partition", partition), zap.Int64("kafka_partition_offset", offset))

		p.logger.Bg().Debug("Message sent to Kafka", logFields...)
	}

	return nil
}

// Close ...
func (p *Publisher) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true

	if err := p.producer.Close(); err != nil {
		return errors.Wrap(err, "cannot close Kafka producer")
	}

	return nil
}
