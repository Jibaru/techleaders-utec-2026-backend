// Package kafka holds the Kafka adapter for the mail.Sender port.
//
// Send() does not deliver the email — it produces a JSON message to a Kafka
// topic. A separate process (cmd/mailworker) consumes that topic and uses
// the smtp adapter to actually send. Same Sender interface as the other
// adapters; only the wire changes.
package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

// Message is the wire format on the topic. cmd/mailworker decodes the same shape.
type Message struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type Sender struct {
	writer *kafka.Writer
}

func New(brokers []string, topic string) *Sender {
	return &Sender{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 50 * time.Millisecond,
		},
	}
}

func (s *Sender) Send(ctx context.Context, to, subject, body string) error {
	payload, err := json.Marshal(Message{To: to, Subject: subject, Body: body})
	if err != nil {
		return err
	}
	return s.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(to),
		Value: payload,
		Time:  time.Now(),
	})
}

func (s *Sender) Close() error {
	return s.writer.Close()
}
