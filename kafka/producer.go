package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
		Async:        false,
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
		return fmt.Errorf("failed to marshal kafka event: %w", err)
	}

	err = p.writer.WriteMessages(
		ctx,
		kafkago.Message{
			Key:   []byte(eventType),
			Value: message,
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to publish kafka event %q: %w",
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

	err := p.writer.WriteMessages(
		ctx,
		kafkago.Message{
			Key:   []byte(eventType),
			Value: message,
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to publish raw kafka event %q: %w",
			eventType,
			err,
		)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
