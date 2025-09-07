package kafkaio

import (
	"context"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func NewWriter(broker, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topic,
		Balancer:               &kafka.Murmur2Balancer{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           200 * time.Millisecond,
	}
}

func NewReader(broker, groupID, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		GroupID:        groupID,
		Topic:          topic,
		CommitInterval: time.Second, // balance latency/throughput
		MinBytes:       1,           // 1B
		MaxBytes:       10e6,        // 10MB
		MaxWait:        250 * time.Millisecond,
	})
}

func CloseQuiet(c interface{ Close() error }) {
	_ = c.Close()
}

func CommitQuiet(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	_ = r.CommitMessages(ctx, m)
}
