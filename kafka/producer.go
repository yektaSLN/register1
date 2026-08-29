package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"login/logger"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkago.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        true,

		Completion: func(
			messages []kafkago.Message,
			err error,
		) {
			if err != nil {
				log := logger.New(os.Stderr)

				log.Error().
					Err(err).
					Msg("failed to deliver kafka message")
			}
		},
	}

	return &Producer{
		writer: writer,
	}
}

func (p *Producer) Publish(
	ctx context.Context,
	eventType string,
	payload any,
) error {

	event := Event{
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}

	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal kafka event: %w",
			err,
		)
	}

	// Persist event to file.
	if err := logger.WriteToFile(
		append(message, '\n'),
	); err != nil {
		return fmt.Errorf(
			"failed to write kafka event to log file: %w",
			err,
		)
	}

	// Async Kafka write.
	if err := p.writer.WriteMessages(
		ctx,
		kafkago.Message{
			Key:   []byte(eventType),
			Value: message,
		},
	); err != nil {
		return fmt.Errorf(
			"failed to queue kafka event %q: %w",
			eventType,
			err,
		)
	}

	return nil
}

func (p *Producer) PublishRaw(
	ctx context.Context,
	eventType string,
	message []byte,
) error {

	// Persist raw HTTP log to file.
	if err := logger.WriteToFile(
		append(message, '\n'),
	); err != nil {
		return fmt.Errorf(
			"failed to write raw kafka event to log file: %w",
			err,
		)
	}

	// Async Kafka write.
	if err := p.writer.WriteMessages(
		ctx,
		kafkago.Message{
			Key:   []byte(eventType),
			Value: message,
		},
	); err != nil {
		return fmt.Errorf(
			"failed to queue raw kafka event %q: %w",
			eventType,
			err,
		)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
