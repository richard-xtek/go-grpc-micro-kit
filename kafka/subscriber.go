package kafka

import (
	"context"
	"sync"
	"time"

	ctx_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf/ctx-logf"

	"github.com/Shopify/sarama"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Subscriber ...
type Subscriber struct {
	config SubscriberConfig
	logger log.Factory

	closing       chan struct{}
	subscribersWg sync.WaitGroup

	closed bool
}

// NewSubscriber creates a new Kafka Subscriber.
func NewSubscriber(
	config SubscriberConfig,
	logger log.Factory,
) (*Subscriber, error) {
	config.setDefaults()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("subscriber_uuid", uuid.NewV4().String()))

	return &Subscriber{
		config: config,
		logger: logger,

		closing: make(chan struct{}),
	}, nil
}

// SubscriberConfig ...
type SubscriberConfig struct {
	// Kafka brokers list.
	Brokers []string

	// Unmarshaler is used to unmarshal messages from Kafka format into Watermill format.
	Unmarshaler Unmarshaler

	// OverwriteSaramaConfig holds additional sarama settings.
	OverwriteSaramaConfig *sarama.Config

	// Kafka consumer group.
	// When empty, all messages from all partitions will be returned.
	ConsumerGroup string

	// How long after Nack message should be redelivered.
	NackResendSleep time.Duration

	// How long about unsuccessful reconnecting next reconnect will occur.
	ReconnectRetrySleep time.Duration

	InitializeTopicDetails *sarama.TopicDetail
}

// NoSleep can be set to SubscriberConfig.NackResendSleep and SubscriberConfig.ReconnectRetrySleep.
const NoSleep time.Duration = -1

func (c *SubscriberConfig) setDefaults() {
	if c.OverwriteSaramaConfig == nil {
		c.OverwriteSaramaConfig = DefaultSaramaSubscriberConfig()
	}
	if c.NackResendSleep == 0 {
		c.NackResendSleep = time.Millisecond * 100
	}
	if c.ReconnectRetrySleep == 0 {
		c.ReconnectRetrySleep = time.Second
	}
}

// Validate ...
func (c SubscriberConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("missing brokers")
	}
	if c.Unmarshaler == nil {
		return errors.New("missing unmarshaler")
	}

	return nil
}

// DefaultSaramaSubscriberConfig creates default Sarama config used by Watermill.
//
// Custom config can be passed to NewSubscriber and NewPublisher.
//
//		saramaConfig := DefaultSaramaSubscriberConfig()
//		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
//
//		subscriberConfig.OverwriteSaramaConfig = saramaConfig
//
//		subscriber, err := NewSubscriber(subscriberConfig, logger)
//		// ...
//
func DefaultSaramaSubscriberConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0
	config.Consumer.Return.Errors = true
	config.ClientID = "watermill"

	return config
}

// Subscribe subscribers for messages in Kafka.
//
// There are multiple subscribers spawned
func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan *Message, error) {
	if s.closed {
		return nil, errors.New("subscriber closed")
	}

	s.subscribersWg.Add(1)

	// logFields := watermill.LogFields{
	// 	"provider":            "kafka",
	// 	"topic":               topic,
	// 	"consumer_group":      s.config.ConsumerGroup,
	// 	"kafka_consumer_uuid": watermill.NewShortUUID(),
	// }

	logFields := []zapcore.Field{
		zap.String("provider", "kafka"),
		zap.String("topic", topic),
		zap.String("consumer_group", s.config.ConsumerGroup),
		zap.String("kafka_consumer_uuid", uuid.NewV4().String()),
	}
	s.logger.Bg().Info("Subscribing to Kafka topic", logFields...)

	// we don't want to have buffered channel to not consume message from Kafka when consumer is not consuming
	output := make(chan *Message, 0)

	consumeClosed, err := s.consumeMessages(ctx, topic, output, logFields)
	if err != nil {
		s.subscribersWg.Done()
		return nil, err
	}

	go func() {
		// blocking, until s.closing is closed
		s.handleReconnects(ctx, topic, output, consumeClosed, logFields)
		close(output)
		s.subscribersWg.Done()
	}()

	return output, nil
}

func (s *Subscriber) handleReconnects(
	ctx context.Context,
	topic string,
	output chan *Message,
	consumeClosed chan struct{},
	logFields []zap.Field,
) {
	for {
		// nil channel will cause deadlock
		if consumeClosed != nil {
			<-consumeClosed
			s.logger.Bg().Debug("consumeMessages stopped", logFields...)
		} else {
			s.logger.Bg().Debug("empty consumeClosed", logFields...)
		}

		select {
		// it's important to don't exit before consumeClosed,
		// to not trigger s.subscribersWg.Done() before consumer is closed
		case <-s.closing:
			s.logger.Bg().Debug("Closing subscriber, no reconnect needed", logFields...)
			return
		case <-ctx.Done():
			s.logger.Bg().Debug("Ctx cancelled, no reconnect needed", logFields...)
			return
		default:
			s.logger.Bg().Debug("Not closing, reconnecting", logFields...)
		}

		s.logger.Bg().Info("Reconnecting consumer", logFields...)

		var err error
		consumeClosed, err = s.consumeMessages(ctx, topic, output, logFields)
		if err != nil {
			s.logger.Bg().Error("Cannot reconnect messages consumer", logFields...)

			if s.config.ReconnectRetrySleep != NoSleep {
				time.Sleep(s.config.ReconnectRetrySleep)
			}
			continue
		}
	}
}

func (s *Subscriber) consumeMessages(
	ctx context.Context,
	topic string,
	output chan *Message,
	logFields []zap.Field,
) (consumeMessagesClosed chan struct{}, err error) {
	s.logger.Bg().Info("Starting consuming", logFields...)

	// Start with a client
	client, err := sarama.NewClient(s.config.Brokers, s.config.OverwriteSaramaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Sarama client")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-s.closing:
			s.logger.Bg().Debug("Closing subscriber, cancelling consumeMessages", logFields...)
			cancel()
		case <-ctx.Done():
			// avoid goroutine leak
		}
	}()

	if s.config.ConsumerGroup == "" {
		consumeMessagesClosed, err = s.consumeWithoutConsumerGroups(ctx, client, topic, output, logFields)
	} else {
		consumeMessagesClosed, err = s.consumeGroupMessages(ctx, client, topic, output, logFields)
	}
	if err != nil {
		s.logger.Bg().Debug(
			"Starting consume failed, cancelling context",
			append(logFields, zap.Error(err))...,
		)
		cancel()
		return nil, err
	}

	go func() {
		<-consumeMessagesClosed
		if err := client.Close(); err != nil {
			s.logger.Bg().Error("Cannot close client", append(logFields, zap.Error(err))...)
		} else {
			s.logger.Bg().Debug("Client closed", logFields...)
		}
	}()

	return consumeMessagesClosed, nil
}

func (s *Subscriber) consumeGroupMessages(
	ctx context.Context,
	client sarama.Client,
	topic string,
	output chan *Message,
	logFields []zap.Field,
) (chan struct{}, error) {
	// Start a new consumer group
	group, err := sarama.NewConsumerGroupFromClient(s.config.ConsumerGroup, client)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create consumer group client")
	}

	groupClosed := make(chan struct{})

	handleGroupErrorsCtx, cancelHandleGroupErrors := context.WithCancel(context.Background())
	handleGroupErrorsDone := s.handleGroupErrors(handleGroupErrorsCtx, group, logFields)

	handler := consumerGroupHandler{
		ctx:              ctx,
		messageHandler:   s.createMessagesHandler(output),
		logger:           s.logger,
		closing:          s.closing,
		messageLogFields: logFields,
	}

	go func() {
		err := group.Consume(ctx, []string{topic}, handler)

		if err != nil {
			if err == sarama.ErrUnknown {
				// this is info, because it is often just noise
				s.logger.Bg().Info("Received unknown Sarama error", append(logFields, zap.Error(err))...)
			} else {
				s.logger.Bg().Error("Group consume error", append(logFields, zap.Error(err))...)
			}
		} else {
			s.logger.Bg().Debug("Consume stopped without any error", append(logFields, zap.Error(err))...)
		}

		cancelHandleGroupErrors()
		<-handleGroupErrorsDone

		if err := group.Close(); err != nil {
			s.logger.Bg().Info("Group close with error", append(logFields, zap.Error(err))...)
		}

		s.logger.Bg().Info("Consuming done", logFields...)
		close(groupClosed)
	}()

	return groupClosed, nil
}

func (s *Subscriber) handleGroupErrors(
	ctx context.Context,
	group sarama.ConsumerGroup,
	logFields []zap.Field,
) chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		errs := group.Errors()

		for {
			select {
			case err := <-errs:
				if err == nil {
					continue
				}

				s.logger.Bg().Error("Sarama internal error", append(logFields, zap.Error(err))...)
			case <-ctx.Done():
				return
			}
		}
	}()

	return done
}

func (s *Subscriber) consumeWithoutConsumerGroups(
	ctx context.Context,
	client sarama.Client,
	topic string,
	output chan *Message,
	logFields []zap.Field,
) (chan struct{}, error) {
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create client")
	}

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get partitions")
	}

	partitionConsumersWg := &sync.WaitGroup{}

	for _, partition := range partitions {
		partitionLogFields := append(logFields, zap.Int32("kafka_partition", partition))
		partitionConsumer, err := consumer.ConsumePartition(topic, partition, s.config.OverwriteSaramaConfig.Consumer.Offsets.Initial)
		if err != nil {
			if err := client.Close(); err != nil && err != sarama.ErrClosedClient {
				s.logger.Bg().Error("Cannot close client", append(partitionLogFields, zap.Error(err))...)
			}
			return nil, errors.Wrap(err, "failed to start consumer for partition")
		}

		messageHandler := s.createMessagesHandler(output)

		partitionConsumersWg.Add(1)
		go s.consumePartition(ctx, partitionConsumer, messageHandler, partitionConsumersWg, partitionLogFields)
	}

	closed := make(chan struct{})
	go func() {
		partitionConsumersWg.Wait()
		close(closed)
	}()

	return closed, nil
}

func (s *Subscriber) consumePartition(
	ctx context.Context,
	partitionConsumer sarama.PartitionConsumer,
	messageHandler messageHandler,
	partitionConsumersWg *sync.WaitGroup,
	logFields []zap.Field,
) {
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			s.logger.Bg().Error("Cannot close partition consumer", append(logFields, zap.Error(err))...)
		}
		partitionConsumersWg.Done()
		s.logger.Bg().Debug("consumePartition stopped", logFields...)

	}()

	kafkaMessages := partitionConsumer.Messages()

	for {
		select {
		case kafkaMsg := <-kafkaMessages:
			if kafkaMsg == nil {
				s.logger.Bg().Debug("kafkaMsg is closed, stopping consumePartition", logFields...)
				return
			}
			if err := messageHandler.processMessage(ctx, kafkaMsg, nil, logFields); err != nil {
				return
			}
		case <-s.closing:
			s.logger.Bg().Debug("Subscriber is closing, stopping consumePartition", logFields...)
			return

		case <-ctx.Done():
			s.logger.Bg().Debug("Ctx was cancelled, stopping consumePartition", logFields...)
			return
		}
	}
}

func (s *Subscriber) createMessagesHandler(output chan *Message) messageHandler {
	return messageHandler{
		outputChannel:   output,
		unmarshaler:     s.config.Unmarshaler,
		nackResendSleep: s.config.NackResendSleep,
		logger:          s.logger,
		closing:         s.closing,
	}
}

// Close ...
func (s *Subscriber) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true
	close(s.closing)
	s.subscribersWg.Wait()

	s.logger.Bg().Debug("Kafka subscriber closed")

	return nil
}

type consumerGroupHandler struct {
	ctx              context.Context
	messageHandler   messageHandler
	logger           log.Factory
	closing          chan struct{}
	messageLogFields []zap.Field
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	kafkaMessages := claim.Messages()

	logFields := append(h.messageLogFields,
		zap.Int32("kafka_partition",
			claim.Partition()), zap.Int64("kafka_initial_offset", claim.InitialOffset()),
	)

	h.logger.Bg().Debug("Consume claimed", logFields...)

	for {
		select {
		case kafkaMsg, ok := <-kafkaMessages:
			if !ok {
				h.logger.Bg().Debug("kafkaMessages is closed, stopping consumerGroupHandler", logFields...)
				return nil
			}
			if err := h.messageHandler.processMessage(h.ctx, kafkaMsg, sess, logFields); err != nil {
				return err
			}

		case <-h.closing:
			h.logger.Bg().Debug("Subscriber is closing, stopping consumerGroupHandler", logFields...)
			return nil

		case <-h.ctx.Done():
			h.logger.Bg().Debug("Ctx was cancelled, stopping consumerGroupHandler", logFields...)
			return nil
		}
	}
}

type messageHandler struct {
	outputChannel chan<- *Message
	unmarshaler   Unmarshaler

	nackResendSleep time.Duration

	logger  log.Factory
	closing chan struct{}
}

func newLoggerForCall(ctx context.Context, logger log.Factory, start time.Time) context.Context {
	f := ctx_logf.TagsToFields(ctx)
	f = append(f, zap.String("event.start_time", start.Format(time.RFC3339)))
	if d, ok := ctx.Deadline(); ok {
		f = append(f, zap.String("event.request.deadline", d.Format(time.RFC3339)))
	}
	callLog := logger.With(f...)

	return ctx_logf.ToContext(ctx, callLog)
}

func (h messageHandler) processMessage(
	ctx context.Context,
	kafkaMsg *sarama.ConsumerMessage,
	sess sarama.ConsumerGroupSession,
	messageLogFields []zap.Field,
) error {
	receivedMsgLogFields := append(messageLogFields, zap.Int64("kafka_partition_offset", kafkaMsg.Offset), zap.Int32("kafka_partition", kafkaMsg.Partition))

	h.logger.Bg().Debug("Received message from Kafka", receivedMsgLogFields...)

	ctx = setPartitionToCtx(ctx, kafkaMsg.Partition)
	ctx = setPartitionOffsetToCtx(ctx, kafkaMsg.Offset)
	ctx = setMessageTimestampToCtx(ctx, kafkaMsg.Timestamp)
	msg, err := h.unmarshaler.Unmarshal(kafkaMsg)
	if err != nil {
		// resend will make no sense, stopping consumerGroupHandler
		return errors.Wrap(err, "message unmarshal failed")
	}

	ctx = setMessageUUIDKeyToCtx(ctx, msg.UUID)
	ctx = newLoggerForCall(ctx, h.logger, time.Now())

	ctx, cancelCtx := context.WithCancel(ctx)

	msg.SetContext(ctx)
	defer cancelCtx()

	receivedMsgLogFields = append(receivedMsgLogFields, zap.String("message_uuid", msg.UUID))

ResendLoop:
	for {
		select {
		case h.outputChannel <- msg:
			h.logger.Bg().Debug("Message sent to consumer", receivedMsgLogFields...)
		case <-h.closing:
			h.logger.Bg().Warn("Closing, message discarded", receivedMsgLogFields...)
			return nil
		case <-ctx.Done():
			h.logger.Bg().Warn("Closing, ctx cancelled before sent to consumer", receivedMsgLogFields...)
			return nil
		}

		select {
		case <-msg.Acked():
			if sess != nil {
				sess.MarkMessage(kafkaMsg, "")
			}
			h.logger.Bg().Debug("Message Acked", receivedMsgLogFields...)
			break ResendLoop
		case <-msg.Nacked():
			h.logger.Bg().Debug("Message Nacked", receivedMsgLogFields...)

			// reset acks, etc.
			msg = msg.Copy()
			if h.nackResendSleep != NoSleep {
				time.Sleep(h.nackResendSleep)
			}

			continue ResendLoop
		case <-h.closing:
			h.logger.Bg().Warn("Closing, message discarded before ack", receivedMsgLogFields...)
			return nil
		case <-ctx.Done():
			h.logger.Bg().Warn("Closing, ctx cancelled before ack", receivedMsgLogFields...)
			return nil
		}
	}

	return nil
}

// SubscribeInitialize ...
func (s *Subscriber) SubscribeInitialize(topic string) (err error) {
	if s.config.InitializeTopicDetails == nil {
		return errors.New("s.config.InitializeTopicDetails is empty, cannot SubscribeInitialize")
	}

	clusterAdmin, err := sarama.NewClusterAdmin(s.config.Brokers, s.config.OverwriteSaramaConfig)
	if err != nil {
		return errors.Wrap(err, "cannot create cluster admin")
	}
	defer func() {
		if closeErr := clusterAdmin.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	if err := clusterAdmin.CreateTopic(topic, s.config.InitializeTopicDetails, false); err != nil {
		return errors.Wrap(err, "cannot create topic")
	}

	s.logger.Bg().Info("Created Kafka topic", zap.String("topic", topic))

	return nil
}

// PartitionOffset ...
type PartitionOffset map[int32]int64

// PartitionOffset ...
func (s *Subscriber) PartitionOffset(topic string) (PartitionOffset, error) {
	client, err := sarama.NewClient(s.config.Brokers, s.config.OverwriteSaramaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Sarama client")
	}

	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	partitions, err := client.Partitions(topic)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get topic partitions")
	}

	partitionOffset := make(PartitionOffset, len(partitions))
	for _, partition := range partitions {
		offset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return nil, err
		}

		partitionOffset[partition] = offset
	}

	return partitionOffset, nil
}
